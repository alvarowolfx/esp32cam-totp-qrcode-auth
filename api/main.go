package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"com.aviebrantz.qrcode_auth/database"
	"com.aviebrantz.qrcode_auth/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	jwtware "github.com/gofiber/jwt"
	"github.com/pquerna/otp/totp"
)

var databaseUri = "mongodb://localhost:27017"
var jwtSecret = "supersecret"

func login(c *fiber.Ctx) {
	email := c.FormValue("email")
	password := c.FormValue("password")
	user, err := repository.CheckUser(email, password)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.SendStatus(fiber.StatusInternalServerError)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	c.JSON(fiber.Map{"message": "OK", "token": t})
}

func createAccount(c *fiber.Ctx) {
	email := c.FormValue("email")
	password := c.FormValue("password")
	user, err := repository.CreateAccount(email, password)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	c.JSON(fiber.Map{"id": user.ID})
}

func generateToken(c *fiber.Ctx) {
	jwtUser := c.Locals("user").(*jwt.Token)
	claims := jwtUser.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	user, err := repository.FindUserByID(userID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      user.ID,
		AccountName: user.Email,
	})

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	secret := key.Secret()

	err = repository.UpdateUserSecret(user.ID, secret)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	c.JSON(fiber.Map{"message": "OK", "token": secret})
}

func sign(content, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(content))
	signed := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signed[0:8]
}

func generateQrCode(c *fiber.Ctx) {
	jwtUser := c.Locals("user").(*jwt.Token)
	claims := jwtUser.Claims.(jwt.MapClaims)
	userID := claims["sub"].(string)

	passcode, err := repository.GetPasscodeForUserID(userID)

	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": err.Error()})
		return
	}

	signedUserID := sign(userID, passcode)

	code := userID + "." + signedUserID
	c.JSON(fiber.Map{"code": code, "passcode": passcode})
}

func checkQrCode(c *fiber.Ctx) {
	code := c.Query("code")
	fmt.Println("Request with code = " + code)

	parts := strings.Split(code, ".")
	if len(parts) != 2 {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": "Invalid format"})
		return
	}

	userID := parts[0]
	signedUserId := parts[1]

	passcode, err := repository.GetPasscodeForUserID(userID)
	if err != nil {
		c.Status(fiber.StatusBadRequest)
		c.JSON(fiber.Map{"message": "Failed to get passcode"})
		return
	}

	counterSignedId := sign(userID, passcode)
	fmt.Println(userID)
	fmt.Println(signedUserId)
	fmt.Println(counterSignedId)

	if strings.Compare(signedUserId, counterSignedId) != 0 {
		c.Status(fiber.StatusForbidden)
		c.JSON(fiber.Map{"message": "Code doesn't match"})
		return
	}

	c.JSON(fiber.Map{"message": "OK"})
}

func main() {
	app := fiber.New()

	database.Connect(databaseUri)

	app.Post("/auth", login)
	app.Post("/user", createAccount)
	app.Get("/check", checkQrCode)

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(jwtSecret),
	}))

	app.Get("/auth/totp", generateToken)

	// Temporary endpoint
	app.Get("/auth/totp/qrcode", generateQrCode)

	app.Listen(8080)

	database.Disconnect()

}
