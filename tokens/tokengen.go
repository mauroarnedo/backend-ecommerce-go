package tokens

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mauroarnedo/ecommerce/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_Name string
	Last_Name  string
	Uid        string
	jwt.RegisteredClaims
}

var UserData *mongo.Collection = database.UserData(database.Client, "users")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func TokenGenerator(email string, firstname string, lastname string, uid string) (signedtoken string, signedrefreshtroken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_Name: firstname,
		Last_Name:  lastname,
		Uid:        uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24))),
		},
	}
	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(168))),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panicln(err)
		return
	}
	return token, refreshToken, err
}
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})

	if err != nil {
		msg = err.Error()
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		msg = "The token is invalid"
		return
	}
	if claims.ExpiresAt.Unix() < time.Now().Local().Unix() {
		msg = "Token is expired"
		return
	}
	return claims, msg
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})
	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updateat", Value: updated_at})
	upsert := true
	filter := bson.M{"user_id": userID}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := UserData.UpdateOne(ctx, filter, bson.D{
		{Key: "$set", Value: updateObj},
	},
		&opt)
	defer cancel()
	if err != nil {
		log.Panic(err)
		return
	}
}
