package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/karthikbhandary2/jwt-go/database"
	helpers "github.com/karthikbhandary2/jwt-go/helpers"
	model "github.com/karthikbhandary2/jwt-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(hash)
}

func VerifyPassword(userPassword, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(providedPassword))
	check := true
	msg := ""
	if err != nil {
		msg = fmt.Sprintf("email or password is incorrect")
		check = false
	}
	return check, msg
	
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

		password := HashPassword(*user.Password)
		user.Password = &password

		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "this email or phone number already exists"})
			return
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
			log.Printf("Database insertion error: %v", err)
			msg := fmt.Sprintf("user item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		log.Printf("User created successfully: %v", resultInsertionNumber.InsertedID)
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc{
	return func (c *gin.Context)  {
		var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		var user model.User
		var foundUser model.User
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}

		isValid, msg := VerifyPassword(*foundUser.Password, *user.Password)

		if isValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return 
		}

		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found"})
			return
		}

		token, refreshToken, _ := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserID, *foundUser.UserType)
		helpers.UpdateAllTokens(token, refreshToken, foundUser.UserID)
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.UserID}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc{
	return func(c *gin.Context) {
		if err :=helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		recordPerPage, err :=strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err :=strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}

		groupStage := bson.D{{
			"$group", bson.D{
				{"_id", "null"},
				{"total_count", bson.D{{"$sum", 1}}},
				{"data", bson.D{{"$push", "$$ROOT"}}},
			},
		}}

		projectStage := bson.D{{
			"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			},
		}}

		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing the users"})
			return
		}

		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])



	}
}

func GetUser() gin.HandlerFunc {
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