package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/karthikbhandary2/jwt-go/controller"
)

func AuthRoutes(ir *gin.Engine) {
	ir.POST("users/signup", controller.Signup())
	ir.POST("users/login", controller.Login())
}