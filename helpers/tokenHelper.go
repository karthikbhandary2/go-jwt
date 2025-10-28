package helpers

import (
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/karthikbhandary2/jwt-go/database"
	"go.mongodb.org/mongo-driver/mongo"
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


func GenerateAllToekens(email string, firstName string, lastName string, uid string, userType string) (signedToken string, signedRefreshToken string, err error) {
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

	signedToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(SECRET_KEY)
	if err != nil {
		log.Panic(err)
		return
	}

	signedRefreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(SECRET_KEY)
	if err != nil {
		log.Panic(err)
		return
	}

	return signedToken, signedRefreshToken, nil
}