package main

import (
	"log"
	"seedance-client/handlers"
	"seedance-client/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	models.InitDB()

	// Initialize Services
	handlers.InitService()

	r := gin.Default()

	// Static files
	r.Static("/static", "./static")
	r.Static("/uploads", "./uploads")
	r.LoadHTMLGlob("templates/*")

	// Middleware to load API key from cookie
	r.Use(handlers.LoadAPIKeyFromCookie())

	// Routes
	r.GET("/", handlers.ListProjects)
	r.POST("/projects", handlers.CreateProject)
	r.POST("/projects/delete/:id", handlers.DeleteProject)
	r.GET("/projects/:id", handlers.GetProject)
	r.POST("/settings/apikey", handlers.UpdateAPIKey)

	// Storyboard Routes
	r.POST("/projects/:id/storyboards", handlers.CreateStoryboard)
	r.POST("/storyboards/delete/:sid", handlers.DeleteStoryboard)
	r.POST("/storyboards/:sid/generate", handlers.GenerateVideo)
	r.GET("/storyboards/:sid/status", handlers.GetStoryboardStatus)
	r.POST("/storyboards/:sid/update", handlers.UpdateStoryboard)

	log.Println("Server starting on :23313")
	if err := r.Run(":23313"); err != nil {
		log.Fatal(err)
	}
}
