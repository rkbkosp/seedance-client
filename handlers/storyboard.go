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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	serviceTier := c.PostForm("service_tier")
	if serviceTier == "" {
		serviceTier = "standard" // Default
	}

	// Create Container Storyboard
	storyboard := models.Storyboard{
		ProjectID: uint(projectID),
		CreatedAt: time.Now(),
	}
	models.DB.Create(&storyboard)

	// Create Initial Take
	take := models.Take{
		StoryboardID:  storyboard.ID,
		Prompt:        prompt,
		ModelID:       modelID,
		Ratio:         ratio,
		Duration:      duration,
		GenerateAudio: generateAudio,
		ServiceTier:   serviceTier,
		Status:        "Draft",
		CreatedAt:     time.Now(),
	}

	if serviceTier == "flex" {
		take.ExpiresAfter = 86400 // 24 hours
	}

	// Handle File Uploads
	firstFrame, _ := c.FormFile("first_frame")
	if firstFrame != nil {
		filename := uuid.New().String() + filepath.Ext(firstFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(firstFrame, dst)
		take.FirstFramePath = dst
	}

	lastFrame, _ := c.FormFile("last_frame")
	if lastFrame != nil {
		filename := uuid.New().String() + filepath.Ext(lastFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(lastFrame, dst)
		take.LastFramePath = dst
	}

	models.DB.Create(&take)
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", projectID))
}

func DeleteStoryboard(c *gin.Context) {
	id := c.Param("sid")
	var sb models.Storyboard
	models.DB.First(&sb, id)
	models.DB.Delete(&sb) // Cascades to Takes
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", sb.ProjectID))
}

func GenerateTakeVideo(c *gin.Context) {
	id := c.Param("tid")
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Take not found"})
		return
	}

	var firstFrameURL, lastFrameURL string

	if take.FirstFramePath != "" {
		b64, err := imageToBase64(take.FirstFramePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process first frame"})
			return
		}
		firstFrameURL = b64
	}

	if take.LastFramePath != "" {
		b64, err := imageToBase64(take.LastFramePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process last frame"})
			return
		}
		lastFrameURL = b64
	}

	taskID, err := volcService.CreateVideoTask(take.ModelID, take.Prompt, firstFrameURL, lastFrameURL, take.Ratio, take.Duration, take.GenerateAudio, take.ServiceTier, take.ExpiresAfter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		take.Status = "Failed"
		models.DB.Save(&take)
		return
	}

	take.TaskID = taskID
	take.Status = "Running"
	models.DB.Save(&take)

	c.JSON(http.StatusOK, gin.H{"status": "submitted", "task_id": taskID})
}

func GetTakeStatus(c *gin.Context) {
	id := c.Param("tid")
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Take not found"})
		return
	}

	if take.TaskID == "" {
		c.JSON(http.StatusOK, gin.H{"status": take.Status})
		return
	}

	pollInterval := 3000
	if take.ServiceTier == "flex" {
		if time.Since(take.CreatedAt) > 10*time.Minute {
			pollInterval = 60000
		} else {
			pollInterval = 10000
		}
	}

	resp, err := volcService.GetTaskStatus(take.TaskID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": take.Status, "error": err.Error()})
		return
	}

	if resp.Status != "" {
		switch strings.ToLower(resp.Status) {
		case "succeeded":
			take.Status = "Succeeded"
		case "failed":
			take.Status = "Failed"
		case "running":
			take.Status = "Running"
		case "queued":
			take.Status = "Queued"
		default:
			take.Status = resp.Status
		}
	}

	if take.Status == "Succeeded" {
		if resp.Content.VideoURL != "" {
			take.VideoURL = resp.Content.VideoURL
		}
		if resp.Content.LastFrameURL != "" {
			take.LastFrameURL = resp.Content.LastFrameURL
		}
		take.TokenUsage = resp.Usage.CompletionTokens
	}

	models.DB.Save(&take)
	c.JSON(http.StatusOK, gin.H{
		"status":         take.Status,
		"video_url":      take.VideoURL,
		"last_frame_url": take.LastFrameURL,
		"poll_interval":  pollInterval,
	})
}

// Helper (unchanged)
func imageToBase64(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	mimeType := "image/png"
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
	if err := models.DB.Preload("Takes", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).First(&sb, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Storyboard not found"})
		return
	}

	// Create New Take based on latest or active?
	// We'll create a new Take with fields from Form, defaulting to copying latest take if form is missing files?
	// Actually, the form submit allows partial updates? No, the edit form sends everything.
	// But files (frames) are optional. If not provided, we should copy from the previous take?
	// The requirement is "Config saved as a Take".
	// Yes, if I edit a take and don't change the image, I expect the new take to still have the image.

	// Find "Previous" Take to copy from
	var prevTake models.Take
	if len(sb.Takes) > 0 {
		prevTake = sb.Takes[len(sb.Takes)-1] // Assuming order
	}

	newTake := models.Take{
		StoryboardID:   sb.ID,
		Prompt:         prevTake.Prompt,
		ModelID:        prevTake.ModelID,
		Ratio:          prevTake.Ratio,
		Duration:       prevTake.Duration,
		ServiceTier:    prevTake.ServiceTier,
		ExpiresAfter:   prevTake.ExpiresAfter,
		FirstFramePath: prevTake.FirstFramePath,
		LastFramePath:  prevTake.LastFramePath,
		Status:         "Draft",
		CreatedAt:      time.Now(),
	}

	// Update with Form Data
	if val := c.PostForm("prompt"); val != "" {
		newTake.Prompt = val
	}
	if val := c.PostForm("model_id"); val != "" {
		newTake.ModelID = val
	}
	if val := c.PostForm("ratio"); val != "" {
		newTake.Ratio = val
	}
	if val := c.PostForm("duration"); val != "" {
		newTake.Duration, _ = strconv.Atoi(val)
	}
	// Checkbox: If checked, value is "true". If unchecked/missing, value is empty -> false.
	newTake.GenerateAudio = c.PostForm("generate_audio") == "true"
	if val := c.PostForm("service_tier"); val != "" {
		newTake.ServiceTier = val
		if val == "flex" {
			newTake.ExpiresAfter = 86400
		} else {
			newTake.ExpiresAfter = 0
		}
	}

	// Handle Deletion Flags (Before uploads, so uploads can overwrite)
	if c.PostForm("delete_first_frame") == "true" {
		newTake.FirstFramePath = ""
	}
	if c.PostForm("delete_last_frame") == "true" {
		newTake.LastFramePath = ""
	}

	// Handle optional file uploads (Overwrite copied paths)
	firstFrame, _ := c.FormFile("first_frame")
	if firstFrame != nil {
		filename := uuid.New().String() + filepath.Ext(firstFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(firstFrame, dst)
		newTake.FirstFramePath = dst
	}

	lastFrame, _ := c.FormFile("last_frame")
	if lastFrame != nil {
		filename := uuid.New().String() + filepath.Ext(lastFrame.Filename)
		dst := filepath.Join("uploads", filename)
		c.SaveUploadedFile(lastFrame, dst)
		newTake.LastFramePath = dst
	}

	models.DB.Create(&newTake)
	c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d", sb.ProjectID))
}

// Additional handlers for Takes
func ListTakes(c *gin.Context) {
	sid := c.Param("sid")
	var takes []models.Take
	if err := models.DB.Where("storyboard_id = ?", sid).Order("created_at asc").Find(&takes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, takes)
}

func ToggleGoodTake(c *gin.Context) {
	id := c.Param("tid")
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Take not found"})
		return
	}

	// Calculate new state
	newState := !take.IsGood

	// If we are marking it as Good, we must unmark all others in this storyboard
	if newState {
		models.DB.Model(&models.Take{}).Where("storyboard_id = ? AND id != ?", take.StoryboardID, take.ID).Update("is_good", false)
	}

	take.IsGood = newState
	models.DB.Save(&take)

	c.JSON(http.StatusOK, gin.H{"is_good": take.IsGood})
}

func DeleteTake(c *gin.Context) {
	id := c.Param("tid")

	// 1. Get the take to know its StoryboardID
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Take not found"})
		return
	}
	storyboardID := take.StoryboardID

	// 2. Delete the take
	if err := models.DB.Delete(&take).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete take"})
		return
	}

	// 3. Check for remaining takes
	var count int64
	models.DB.Model(&models.Take{}).Where("storyboard_id = ?", storyboardID).Count(&count)

	// 4. If no takes left, delete the container
	storyboardDeleted := false
	if count == 0 {
		if err := models.DB.Delete(&models.Storyboard{}, storyboardID).Error; err == nil {
			storyboardDeleted = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"storyboard_deleted": storyboardDeleted,
		"remaining_takes":    count,
	})
}

func GetTake(c *gin.Context) {
	id := c.Param("tid")
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Take not found"})
		return
	}
	c.JSON(http.StatusOK, take)
}
