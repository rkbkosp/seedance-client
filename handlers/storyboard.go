package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seedance-client/models"
	"seedance-client/services"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var volcService *services.VolcEngineService

func InitService() {
	volcService = services.NewVolcEngineService()
}

// ensureUploadsDir checks if uploads directory exists
func ensureUploadsDir() {
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}
}

func CreateStoryboard(c *gin.Context) {
	ensureUploadsDir()

	projectID, _ := strconv.Atoi(c.Param("id"))
	prompt := c.PostForm("prompt")
	modelID := c.PostForm("model_id")
	ratio := c.PostForm("ratio")
	durationStr := c.PostForm("duration")
	duration, _ := strconv.Atoi(durationStr)
	generateAudio := c.PostForm("generate_audio") == "true"

	storyboard := models.Storyboard{
		ProjectID:     uint(projectID),
		Prompt:        prompt,
		ModelID:       modelID,
		Ratio:         ratio,
		Duration:      duration,
		GenerateAudio: generateAudio,
		Status:        "Draft",
	}

	// Handle File Uploads
	firstFrame, _ := c.FormFile("first_frame")
	if firstFrame != nil {
		filename := uuid.New().String() + filepath.Ext(firstFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(firstFrame, dst)
		storyboard.FirstFramePath = dst
	}

	lastFrame, _ := c.FormFile("last_frame")
	if lastFrame != nil {
		filename := uuid.New().String() + filepath.Ext(lastFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(lastFrame, dst)
		storyboard.LastFramePath = dst
	}

	models.DB.Create(&storyboard)
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", projectID))
}

func DeleteStoryboard(c *gin.Context) {
	id := c.Param("sid")
	var sb models.Storyboard
	models.DB.First(&sb, id)
	models.DB.Delete(&sb)
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", sb.ProjectID))
}

func GenerateVideo(c *gin.Context) {
	id := c.Param("sid")
	var sb models.Storyboard
	if err := models.DB.First(&sb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storyboard not found"})
		return
	}

	// Prepare Image URLs (Need to be accessible by Volcengine if using URL, but API docs support base64 too?)
	// User request says: "支持上传本地图片，自动转 Base64 或上传到服务端".
	// The API docs createReq show "ImageURL" struct. Looking at "图生视频base64" in user request,
	// it shows "url": "data:image/png;base64,...". So we can pass Base64 data URI as the URL.

	var firstFrameURL, lastFrameURL string

	if sb.FirstFramePath != "" {
		b64, err := imageToBase64(sb.FirstFramePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process first frame"})
			return
		}
		firstFrameURL = b64
	}

	if sb.LastFramePath != "" {
		b64, err := imageToBase64(sb.LastFramePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process last frame"})
			return
		}
		lastFrameURL = b64
	}

	taskID, err := volcService.CreateVideoTask(sb.ModelID, sb.Prompt, firstFrameURL, lastFrameURL, sb.Ratio, sb.Duration, sb.GenerateAudio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		sb.Status = "Failed"
		models.DB.Save(&sb)
		return
	}

	sb.TaskID = taskID
	sb.Status = "Running" // Or Queued
	models.DB.Save(&sb)

	c.JSON(http.StatusOK, gin.H{"status": "submitted", "task_id": taskID})
}

func GetStoryboardStatus(c *gin.Context) {
	id := c.Param("sid")
	var sb models.Storyboard
	if err := models.DB.First(&sb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storyboard not found"})
		return
	}

	if sb.TaskID == "" {
		c.JSON(http.StatusOK, gin.H{"status": sb.Status})
		return
	}

	// Call API to check status
	resp, err := volcService.GetTaskStatus(sb.TaskID)
	// API docs for callback: "queued", "running", "succeeded", "failed", "expired".
	// GetTask response content should have Status.
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": sb.Status, "error": err.Error()}) // Don't fail the request, just return old status
		return
	}

	// Update DB - normalize status to Title case for template compatibility
	if resp.Status != "" {
		// Normalize status: API returns lowercase, template expects Title case
		switch strings.ToLower(resp.Status) {
		case "succeeded":
			sb.Status = "Succeeded"
		case "failed":
			sb.Status = "Failed"
		case "running":
			sb.Status = "Running"
		case "queued":
			sb.Status = "Queued"
		default:
			sb.Status = resp.Status
		}
	}

	if sb.Status == "Succeeded" {
		// Content is a struct with VideoURL and LastFrameURL
		if resp.Content.VideoURL != "" {
			sb.VideoURL = resp.Content.VideoURL
		}
		if resp.Content.LastFrameURL != "" {
			sb.LastFrameURL = resp.Content.LastFrameURL
		}
	}

	models.DB.Save(&sb)
	c.JSON(http.StatusOK, gin.H{
		"status":         sb.Status,
		"video_url":      sb.VideoURL,
		"last_frame_url": sb.LastFrameURL,
	})
}

// Helper
func imageToBase64(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read entire file into byte slice
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Determine mime type (simple extension check)
	mimeType := "image/png" // default
	if strings.HasSuffix(strings.ToLower(path), ".jpg") || strings.HasSuffix(strings.ToLower(path), ".jpeg") {
		mimeType = "image/jpeg"
	}

	b64 := base64.StdEncoding.EncodeToString(bytes)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, b64), nil
}

func UpdateStoryboard(c *gin.Context) {
	ensureUploadsDir()

	id := c.Param("sid")
	var sb models.Storyboard
	if err := models.DB.First(&sb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storyboard not found"})
		return
	}

	// Update fields
	prompt := c.PostForm("prompt")
	if prompt != "" {
		sb.Prompt = prompt
	}

	modelID := c.PostForm("model_id")
	if modelID != "" {
		sb.ModelID = modelID
	}

	ratio := c.PostForm("ratio")
	if ratio != "" {
		sb.Ratio = ratio
	}

	durationStr := c.PostForm("duration")
	if durationStr != "" {
		duration, _ := strconv.Atoi(durationStr)
		sb.Duration = duration
	}

	generateAudioStr := c.PostForm("generate_audio")
	if generateAudioStr != "" {
		sb.GenerateAudio = generateAudioStr == "true"
	}

	// Handle optional file uploads
	firstFrame, _ := c.FormFile("first_frame")
	if firstFrame != nil {
		// Delete old file if exists
		if sb.FirstFramePath != "" {
			os.Remove(sb.FirstFramePath)
		}
		filename := uuid.New().String() + filepath.Ext(firstFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(firstFrame, dst)
		sb.FirstFramePath = dst
	}

	lastFrame, _ := c.FormFile("last_frame")
	if lastFrame != nil {
		// Delete old file if exists
		if sb.LastFramePath != "" {
			os.Remove(sb.LastFramePath)
		}
		filename := uuid.New().String() + filepath.Ext(lastFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(lastFrame, dst)
		sb.LastFramePath = dst
	}

	// Reset status to Draft after editing
	sb.Status = "Draft"
	sb.TaskID = ""
	sb.VideoURL = ""
	sb.LastFrameURL = ""

	models.DB.Save(&sb)
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", sb.ProjectID))
}
