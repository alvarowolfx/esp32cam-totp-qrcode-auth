# QRCode based authentication/access system using ESP32Cam and Go API

This is an IoT pet project of an access systems based on QR Codes. It's separated in 3 parts:

- Web App
  - User creates accounts
  - Requests an TOTP code and saves securely locally
  - Generate time based QRCode using TOTP code
- API
  - User Management
  - TOTP Secret Management
  - Check if QRCode is valid
- Hardware
  - Reads QR Code locally with a camera
  - Sends data to an Webhook
    - If returns 200 - Open Relay
    - It not - Close relay

* EN-US - See demo video at https://twitter.com/alvaroviebrantz/status/1290116219199279104?s=20
* PT-BR - Veja video de demo em https://twitter.com/alvaroviebrantz/status/1290116405824806912?s=20

The live codings session happens in Portuguese (PT-BR) and you can follow on my Twitch channels.

- [Alvaro Viebrantz on Twitch](https://twitch.tv/alvaroviebrantz)

## Project Structure

- Web App - Not Implemented yet
- Go API - `api` folder
  - Authentication and authorization server.
  - Handles the TOTP secrets
  - Use MongoDB as database
- ESP32 QRCode Reader - `firmware` folder
