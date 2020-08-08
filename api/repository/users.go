package repository

import (
	"context"
	"encoding/base32"
	"errors"
	"time"

	"com.aviebrantz.qrcode_auth/database"
	"com.aviebrantz.qrcode_auth/model"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var databaseName = "qrcode_auth"
var userCollection = "users"

func FindUserByEmail(email string) (error, *model.User) {
	ctx := context.Background()
	db := database.Client.Database(databaseName)
	collection := db.Collection(userCollection)

	filter := bson.M{"email": email}
	user := &model.User{}
	err := collection.FindOne(ctx, filter).Decode(user)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return err, nil
	}

	user.ID = user.InternalID.Hex()

	return nil, user
}

func FindUserByID(userID string) (*model.User, error) {
	ctx := context.Background()
	db := database.Client.Database(databaseName)
	collection := db.Collection(userCollection)

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": oid}
	user := &model.User{}
	err = collection.FindOne(ctx, filter).Decode(user)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	user.ID = user.InternalID.Hex()

	return user, nil
}

func UpdateUserSecret(userID, secret string) error {
	ctx := context.Background()
	db := database.Client.Database(databaseName)
	collection := db.Collection(userCollection)

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	res, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.D{
			{"$set", bson.D{{"secret", secret}}},
		},
	)

	if res.ModifiedCount < 1 {
		return errors.New("User not found")
	}

	if err != nil {
		return err
	}

	return nil
}

func CreateAccount(email, password string) (*model.User, error) {
	db := database.Client.Database(databaseName)
	collection := db.Collection(userCollection)
	ctx := context.Background()

	err, userFound := FindUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if userFound != nil && userFound.Email != "" {
		return nil, errors.New("User already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return nil, err
	}
	encryptedPassword := string(hash)

	res, err := collection.InsertOne(ctx, bson.M{"email": email, "password": encryptedPassword})
	if err != nil {
		return nil, err
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("Error getting user id")
	}

	user := &model.User{
		ID:    oid.Hex(),
		Email: email,
	}
	return user, nil
}

func CheckUser(email, password string) (*model.User, error) {
	err, user := FindUserByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("User not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("Email and password doesn't match")
	}

	return user, nil
}

func GetPasscodeForUserID(userID string) (string, error) {
	user, err := FindUserByID(userID)
	if err != nil {
		return "", err
	}

	secret := base32.StdEncoding.EncodeToString([]byte(user.Secret))
	passcode, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA512,
	})

	return passcode, err
}
