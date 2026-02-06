package handlers

import (
	"net/http"
	"seedance-client/models"
	"seedance-client/services"

	"github.com/gin-gonic/gin"
)

func ListProjects(c *gin.Context) {
	var projects []models.Project
	models.DB.Order("created_at desc").Find(&projects)

	// Calculate Global Stats
	var storyboards []models.Storyboard
	models.DB.Where("status = ?", "Succeeded").Find(&storyboards)

	var stats struct {
		TotalVideos       int
		TotalVideos15Pro  int
		TotalVideos10Fast int
		TotalTokenUsage   int
		TotalCost         float64
		TotalSavings      float64
	}

	stats.TotalVideos = len(storyboards)

	for _, sb := range storyboards {
		// Count by Model
		is15Pro := sb.ModelID == "doubao-seedance-1-5-pro-251215"
		if is15Pro {
			stats.TotalVideos15Pro++
		} else {
			stats.TotalVideos10Fast++
		}

		// Token Usage
		stats.TotalTokenUsage += sb.TokenUsage

		// Calculate Cost
		var pricePerMillion float64

		// Determine Price based on Model, Audio, and Tier
		if is15Pro {
			// 1.5 Pro
			if sb.GenerateAudio {
				if sb.ServiceTier == "flex" {
					pricePerMillion = 8.0
				} else {
					pricePerMillion = 16.0
				}
			} else { // Silent
				if sb.ServiceTier == "flex" {
					pricePerMillion = 4.0
				} else {
					pricePerMillion = 8.0
				}
			}
		} else {
			// 1.0 Pro Fast
			if sb.ServiceTier == "flex" {
				pricePerMillion = 2.1
			} else {
				pricePerMillion = 4.2
			}
		}

		cost := (float64(sb.TokenUsage) / 1000000.0) * pricePerMillion
		stats.TotalCost += cost

		// Calculate Savings
		var platformPrice float64
		if is15Pro {
			platformPrice = 2.56
		} else {
			platformPrice = 0.64
		}
		stats.TotalSavings += (platformPrice - cost)
	}

	c.HTML(http.StatusOK, "projects.html", gin.H{
		"Projects": projects,
		"Stats":    stats,
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

// ExportProject exports all succeeded storyboard videos as a ZIP bundle with FCPXML
func ExportProject(c *gin.Context) {
	id := c.Param("id")
	var project models.Project
	if err := models.DB.Preload("Storyboards").First(&project, id).Error; err != nil {
		c.String(404, "Project not found")
		return
	}

	// Generate export data
	exports := services.PrepareExportData(project.Storyboards)
	if len(exports) == 0 {
		c.String(400, "No succeeded videos available for export")
		return
	}

	// Set headers for file download
	filename := services.GetExportFilename(project.Name)
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")

	// Stream ZIP directly to response
	if err := services.CreateExportZIP(c.Writer, project.Name, exports); err != nil {
		// Log error - headers already sent so can't change response
		c.Error(err)
		return
	}
}
