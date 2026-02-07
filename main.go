package main

import (
	"html/template"
	"log"
	"seedance-client/handlers"
	"seedance-client/models"
	"seedance-client/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Database
	models.InitDB()

	// Initialize Services
	handlers.InitService()

	// Start background asset downloader
	services.StartBackgroundDownloader()

	r := gin.Default()

	// Load embedded templates
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}).ParseFS(templatesFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	// Static files - uploads and downloads need to be served from disk
	r.Static("/uploads", "./uploads")
	r.Static("/downloads", "./downloads")

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
	// Storyboard Routes
	r.POST("/projects/:id/storyboards", handlers.CreateStoryboard)
	r.POST("/storyboards/delete/:sid", handlers.DeleteStoryboard)
	r.POST("/storyboards/:sid/update", handlers.UpdateStoryboard)
	r.GET("/storyboards/:sid/takes", handlers.ListTakes)

	// Take Routes
	r.POST("/takes/:tid/generate", handlers.GenerateTakeVideo)
	r.GET("/takes/:tid/status", handlers.GetTakeStatus)
	r.GET("/takes/:tid", handlers.GetTake)
	r.POST("/takes/:tid/toggle_good", handlers.ToggleGoodTake)
	r.POST("/takes/delete/:tid", handlers.DeleteTake)

	log.Println("Server starting on :23313")
	if err := r.Run(":23313"); err != nil {
		log.Fatal(err)
	}
}
