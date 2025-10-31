package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/karthikbhandary2/jwt-go/helpers"
)


func Authenticate() gin.HandlerFunc {
	return func (c *gin.Context)  {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("first_name", claims.FirstName)
		c.Set("last_name", claims.LastName)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.UserType)
		c.Next()
	
	}
}