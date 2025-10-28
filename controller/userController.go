package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/karthikbhandary2/jwt-go/database"
	"github.com/karthikbhandary2/jwt-go/helpers"
	"github.com/karthikbhandary2/jwt-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)


var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) (string, error) {
	return "", nil
}

func VerifyPassword(hashedPassword, password string) error {
	return nil
}

func Signup() gin.HandlerFunc{
	return func (c *gin.Context)  {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user model.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{"email":user.Email})
		if err != nil {
			log.Panic(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "this email or phone number already exists"})
		}

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()

		token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.FirstName, *user.LastName, *user.UserType, user.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while generating tokens"})
			return
		}

		user.Token = &token
		user.RefreshToken = &refreshToken

		resultInsertionNumber, err := userCollection.InsertOne(ctx, user)
		if err != nil {
			msg := fmt.Sprintf("user item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() {

}

func GetUsers() {

}

func GetUser(userId string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err :=helpers.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user model.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the user"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}