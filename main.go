package main

import (
	"html/template"
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

	// Load embedded templates
	tmpl := template.Must(template.New("").ParseFS(templatesFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	// Static files - uploads still need to be served from disk
	r.Static("/uploads", "./uploads")

	// Middleware to load API key from cookie
	r.Use(handlers.LoadAPIKeyFromCookie())

	// Routes
	r.GET("/", handlers.ListProjects)
	r.POST("/projects", handlers.CreateProject)
	r.POST("/projects/delete/:id", handlers.DeleteProject)
	r.GET("/projects/:id", handlers.GetProject)
	r.GET("/projects/:id/export", handlers.ExportProject)
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
