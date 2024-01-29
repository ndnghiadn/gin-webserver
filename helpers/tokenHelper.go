package helper

import (
	"context"
	"gin-webserver/database"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email 		string 
	First_Name 	string 
	Last_Name 	string
	Uid 		string
	User_Type	string
	jwt.StandardClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("JWT")

func GenerateAllToken(email string, first_name string, last_name string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails {
		Email: email,
		First_Name: first_name,
		Last_Name: last_name,
		Uid: uid,
		User_Type: userType,
		StandardClaims: jwt.StandardClaims {
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}

	refreshClaims := &SignedDetails {
		StandardClaims: jwt.StandardClaims {
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100 * time.Second)

	var updateObj primitive.D 

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	Updated_At, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_At})

	upsert := true
	filter := bson.M{"user_id": userId}
	opt := options.UpdateOptions {
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, filter, bson.D{"$set", updateObj}, &opt)

	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
	return
}