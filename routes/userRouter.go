package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/karthikbhandary2/jwt-go/middleware"
	controller "github.com/karthikbhandary2/jwt-go/controller"
)

func UserRoutes(ir *gin.Engine) {
	ir.Use(middleware.Authenticate())
	ir.GET("/users", controller.GetUsers())
	ir.GET("/users/:user_id", controller.GetUser())
}