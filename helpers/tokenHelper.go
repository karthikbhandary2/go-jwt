package helpers

import (
	"context"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/karthikbhandary2/jwt-go/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	UserType  string
	jwt.StandardClaims
}


var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func init() {
	log.Printf("SECRET_KEY loaded: '%s' (length: %d)", SECRET_KEY, len(SECRET_KEY))
}


func GenerateAllTokens(email string, firstName string, lastName string, uid string, userType string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		Uid: uid,
		UserType: userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.TimeFunc().Add(time.Hour * 24).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * 168).Unix(),
		},
	}

	signedToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	signedRefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})

	updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", updated_at})

	upsert := true
	filter := bson.M{"_id": userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_, err := userCollection.UpdateOne(ctx, filter, bson.D{{"$set", updateObj}}, &opt)
	if err != nil {
		log.Panic(err)
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(signedToken, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
		},
	)
	if err != nil {
		log.Println(err)
		return
	}
	claims, ok := token.Claims.(*SignedDetails)
	if !ok {
		log.Println("the token is invalid")
		return 
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		log.Println("Token expired")
		return
	}
	return claims, ""
}