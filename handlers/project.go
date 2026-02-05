package handlers

import (
	"net/http"
	"seedance-client/models"

	"github.com/gin-gonic/gin"
)

func ListProjects(c *gin.Context) {
	var projects []models.Project
	models.DB.Order("created_at desc").Find(&projects)
	c.HTML(http.StatusOK, "projects.html", gin.H{
		"Projects": projects,
	})
}

func CreateProject(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	project := models.Project{Name: name}
	models.DB.Create(&project)
	c.Redirect(http.StatusFound, "/")
}

func DeleteProject(c *gin.Context) {
	id := c.Param("id")
	models.DB.Delete(&models.Project{}, id)
	c.Redirect(http.StatusFound, "/")
}

func GetProject(c *gin.Context) {
	id := c.Param("id")
	var project models.Project
	if err := models.DB.Preload("Storyboards").First(&project, id).Error; err != nil {
		c.String(http.StatusNotFound, "Project not found")
		return
	}
	c.HTML(http.StatusOK, "storyboard.html", gin.H{
		"Project": project,
	})
}

func UpdateAPIKey(c *gin.Context) {
	apiKey := c.PostForm("api_key")
	if apiKey != "" {
		volcService.SetAPIKey(apiKey)
		// Set cookie for 30 days (in seconds)
		c.SetCookie("ark_api_key", apiKey, 60*60*24*30, "/", "", false, true)
	}
	// Return to the previous page
	referer := c.Request.Header.Get("Referer")
	if referer == "" {
		referer = "/"
	}
	c.Redirect(http.StatusFound, referer)
}

// Middleware to load API key from cookie on each request
func LoadAPIKeyFromCookie() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if service has no API key set (empty from env)
		apiKey, err := c.Cookie("ark_api_key")
		if err == nil && apiKey != "" {
			volcService.SetAPIKey(apiKey)
		}
		c.Next()
	}
}
