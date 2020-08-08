#include <Arduino.h>
#include <WiFi.h>
#include <HTTPClient.h>
#include <ESP32QRCodeReader.h>

#define WIFI_SSID "YOUR_SSID"
#define WIFI_PASSWORD "YOUR_PASSWORD"
#define WEBHOOK_URL "http://your-server-url:8080/check"
#define DOOR_RELAY_PIN 12

ESP32QRCodeReader reader(CAMERA_MODEL_AI_THINKER);
struct QRCodeData qrCodeData;
bool isConnected = false;
char deviceName[24] = "";

void openDoor()
{
  digitalWrite(DOOR_RELAY_PIN, LOW);
}

void closeDoor()
{
  digitalWrite(DOOR_RELAY_PIN, HIGH);
}

bool connectWifi()
{
  if (WiFi.status() == WL_CONNECTED)
  {
    return true;
  }

  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  int maxRetries = 10;
  while (WiFi.status() != WL_CONNECTED)
  {
    delay(500);
    Serial.print(".");
    maxRetries--;
    if (maxRetries <= 0)
    {
      return false;
    }
  }
  Serial.println("");
  Serial.println("WiFi connected");
  return true;
}

void callWebhook(String code)
{
  HTTPClient http;
  http.begin(String(WEBHOOK_URL) + "?code=" + code + "&deviceId=" + String(deviceName));

  int httpCode = http.GET();
  if (httpCode == HTTP_CODE_OK)
  {
    Serial.println("Open door");
    openDoor();
    delay(2000);
    closeDoor();
  }
  else
  {
    Serial.println("Not authorized");
    closeDoor();
  }

  http.end();
}

void setup()
{
  Serial.begin(115200);
  Serial.println();

  uint64_t chipId = ESP.getEfuseMac();
  uint32_t highBytesChipId = (uint32_t)(chipId >> 16); // High 4 bytes
  //uint16_t lowBytesChipId = (uint16_t)chipId; // Low 2 bytes
  snprintf(deviceName, sizeof(deviceName), "READER_%08X", highBytesChipId);

  pinMode(DOOR_RELAY_PIN, OUTPUT);
  closeDoor();

  reader.setup();
  //reader.setDebug(true);
  Serial.println("Setup QRCode Reader");

  reader.begin();
  Serial.println("Begin QR Code reader");

  delay(1000);
}

void loop()
{
  bool connected = connectWifi();
  if (isConnected != connected)
  {
    isConnected = connected;
  }
  if (reader.receiveQrCode(&qrCodeData, 100))
  {
    Serial.println("Found QRCode");
    if (qrCodeData.valid)
    {
      Serial.print("Payload: ");
      Serial.println((const char *)qrCodeData.payload);
      callWebhook(String((const char *)qrCodeData.payload));
    }
    else
    {
      Serial.print("Invalid: ");
      Serial.println((const char *)qrCodeData.payload);
    }
  }
  delay(300);
}