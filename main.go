package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/karthikbhandary2/jwt-go/routes"
	"github.com/karthikbhandary2/jwt-go/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	routes.AuthRoutes(router)

	router.GET("/test-db", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Test database connection
		err := database.Client.Ping(ctx, nil)
		if err != nil {
			c.JSON(500, gin.H{"error": "Database connection failed", "details": err.Error()})
			return
		}
		
		// Test collection access
		userCollection := database.OpenCollection(database.Client, "user")
		count, err := userCollection.CountDocuments(ctx, bson.M{})
		if err != nil {
			c.JSON(500, gin.H{"error": "Collection access failed", "details": err.Error()})
			return
		}
		
		c.JSON(200, gin.H{"message": "Database working", "user_count": count})
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Test endpoint working"})
	})

	routes.UserRoutes(router)
		c.JSON(200, gin.H{"message": "Test endpoint working"})
	})

	router.GET("/api-1", func(c *gin.Context) {
		c.JSON(200, gin.H{"success":"Access granted for api-1"})
	})
	router.GET("/api-2", func (c *gin.Context)  {
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		panic(err)
	}
}