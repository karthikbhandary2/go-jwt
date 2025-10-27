package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/karthikbhandary2/jwt-go/database"
	"github.com/karthikbhandary2/jwt-go/helpers"
	"github.com/kata-containers/kata-containers/src/runtime/virtcontainers/pkg/firecracker/client/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func HashPassword(password string) (string, error) {

}

func VerifyPassword(hashedPassword, password string) error {

}

func Signup() {

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

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while fetching the user"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}