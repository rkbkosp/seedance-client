package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"seedance-client/config"
	"seedance-client/models"
	"seedance-client/services"

	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

// App struct holds application state
type App struct {
	ctx         context.Context
	volcService *services.VolcEngineService
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.volcService = services.NewVolcEngineService()

	// Load saved API key from database
	apiKey := a.GetSavedAPIKey()
	if apiKey != "" {
		a.volcService.SetAPIKey(apiKey)
	}
}

// ============================================================
// Settings
// ============================================================

// GetSavedAPIKey retrieves the API key from database or environment
func (a *App) GetSavedAPIKey() string {
	var setting models.Setting
	if err := models.DB.Where("`key` = ?", "ark_api_key").First(&setting).Error; err != nil {
		return os.Getenv("ARK_API_KEY")
	}
	return setting.Value
}

// UpdateAPIKey saves the API key and updates the service
func (a *App) UpdateAPIKey(apiKey string) {
	if apiKey == "" {
		return
	}
	models.DB.Where("`key` = ?", "ark_api_key").Assign(models.Setting{Value: apiKey}).FirstOrCreate(&models.Setting{Key: "ark_api_key"})
	a.volcService.SetAPIKey(apiKey)
}

// HasAPIKey checks if an API key is configured
func (a *App) HasAPIKey() bool {
	return a.GetSavedAPIKey() != ""
}

// ============================================================
// Projects
// ============================================================

// ProjectStats holds aggregated statistics
type ProjectStats struct {
	TotalVideos     int            `json:"total_videos"`
	TotalTokenUsage int            `json:"total_token_usage"`
	TotalCost       float64        `json:"total_cost"`
	TotalSavings    float64        `json:"total_savings"`
	ModelVideoCount map[string]int `json:"model_video_count"`
}

// ProjectsData is the response for the projects list page
type ProjectsData struct {
	Projects []models.Project `json:"projects"`
	Stats    ProjectStats     `json:"stats"`
}

// GetProjects returns all projects with stats
func (a *App) GetProjects() (*ProjectsData, error) {
	var projects []models.Project
	models.DB.Order("created_at desc").Find(&projects)

	var takes []models.Take
	models.DB.Where("status = ?", "Succeeded").Find(&takes)

	modelVideoCount := make(map[string]int)
	var totalTokenUsage int
	var totalCost float64
	var totalSavings float64

	for _, take := range takes {
		modelVideoCount[take.ModelID]++
		totalTokenUsage += take.TokenUsage
		pricePerMillion := config.GetPricePerMillion(take.ModelID, take.ServiceTier, take.GenerateAudio)
		cost := (float64(take.TokenUsage) / 1000000.0) * pricePerMillion
		totalCost += cost
		platformPrice := config.GetPlatformPrice(take.ModelID)
		totalSavings += (platformPrice - cost)
	}

	return &ProjectsData{
		Projects: projects,
		Stats: ProjectStats{
			TotalVideos:     len(takes),
			TotalTokenUsage: totalTokenUsage,
			TotalCost:       totalCost,
			TotalSavings:    totalSavings,
			ModelVideoCount: modelVideoCount,
		},
	}, nil
}

// CreateProject creates a new project
func (a *App) CreateProject(name string) error {
	if name == "" {
		return fmt.Errorf("project name is required")
	}
	return models.DB.Create(&models.Project{Name: name}).Error
}

// DeleteProject deletes a project by ID
func (a *App) DeleteProject(id uint) error {
	return models.DB.Delete(&models.Project{}, id).Error
}

// ============================================================
// Project Detail types
// ============================================================

// TakeResponse is the JSON-friendly take structure
type TakeResponse struct {
	ID                 uint      `json:"id"`
	StoryboardID       uint      `json:"storyboard_id"`
	Prompt             string    `json:"prompt"`
	FirstFramePath     string    `json:"first_frame_path"`
	LastFramePath      string    `json:"last_frame_path"`
	ModelID            string    `json:"model_id"`
	Ratio              string    `json:"ratio"`
	Duration           int       `json:"duration"`
	GenerateAudio      bool      `json:"generate_audio"`
	TaskID             string    `json:"task_id"`
	Status             string    `json:"status"`
	VideoURL           string    `json:"video_url"`
	LastFrameURL       string    `json:"last_frame_url"`
	LocalVideoPath     string    `json:"local_video_path"`
	LocalLastFramePath string    `json:"local_last_frame_path"`
	DownloadStatus     string    `json:"download_status"`
	ServiceTier        string    `json:"service_tier"`
	TokenUsage         int       `json:"token_usage"`
	ExpiresAfter       int64     `json:"expires_after"`
	IsGood             bool      `json:"is_good"`
	CreatedAt          time.Time `json:"created_at"`
}

func takeToResponse(take *models.Take) TakeResponse {
	return TakeResponse{
		ID:                 take.ID,
		StoryboardID:       take.StoryboardID,
		Prompt:             take.Prompt,
		FirstFramePath:     take.FirstFramePath,
		LastFramePath:      take.LastFramePath,
		ModelID:            take.ModelID,
		Ratio:              take.Ratio,
		Duration:           take.Duration,
		GenerateAudio:      take.GenerateAudio,
		TaskID:             take.TaskID,
		Status:             take.Status,
		VideoURL:           services.GetEffectiveVideoURL(take),
		LastFrameURL:       services.GetEffectiveLastFrameURL(take),
		LocalVideoPath:     take.LocalVideoPath,
		LocalLastFramePath: take.LocalLastFramePath,
		DownloadStatus:     take.DownloadStatus,
		ServiceTier:        take.ServiceTier,
		TokenUsage:         take.TokenUsage,
		ExpiresAfter:       take.ExpiresAfter,
		IsGood:             take.IsGood,
		CreatedAt:          take.CreatedAt,
	}
}

// StoryboardData is the JSON-friendly storyboard structure
type StoryboardData struct {
	ID         uint           `json:"id"`
	ProjectID  uint           `json:"project_id"`
	Takes      []TakeResponse `json:"takes"`
	CreatedAt  time.Time      `json:"created_at"`
	ActiveTake *TakeResponse  `json:"active_take"`
}

// ProjectDetail is the project with storyboards
type ProjectDetail struct {
	ID          uint             `json:"id"`
	Name        string           `json:"name"`
	CreatedAt   time.Time        `json:"created_at"`
	Storyboards []StoryboardData `json:"storyboards"`
}

// ProjectDetailData is the response for the project detail page
type ProjectDetailData struct {
	Project              ProjectDetail  `json:"project"`
	Models               []config.Model `json:"models"`
	AudioSupportedModels []string       `json:"audio_supported_models"`
}

// GetProject returns project detail with storyboards and takes
func (a *App) GetProject(id uint) (*ProjectDetailData, error) {
	var project models.Project
	err := models.DB.Preload("Storyboards", func(db *gorm.DB) *gorm.DB {
		return db.Order("id asc")
	}).Preload("Storyboards.Takes", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).First(&project, id).Error

	if err != nil {
		return nil, fmt.Errorf("project not found")
	}

	var storyboards []StoryboardData
	for _, sb := range project.Storyboards {
		var takes []TakeResponse
		for i := range sb.Takes {
			takes = append(takes, takeToResponse(&sb.Takes[i]))
		}

		sbData := StoryboardData{
			ID:        sb.ID,
			ProjectID: sb.ProjectID,
			Takes:     takes,
			CreatedAt: sb.CreatedAt,
		}

		// Calculate active take: Good Take (latest) > Latest Created
		if len(takes) > 0 {
			var bestTake *TakeResponse
			for j := len(takes) - 1; j >= 0; j-- {
				if takes[j].IsGood {
					t := takes[j]
					bestTake = &t
					break
				}
			}
			if bestTake == nil {
				t := takes[len(takes)-1]
				bestTake = &t
			}
			sbData.ActiveTake = bestTake
		}

		storyboards = append(storyboards, sbData)
	}

	return &ProjectDetailData{
		Project: ProjectDetail{
			ID:          project.ID,
			Name:        project.Name,
			CreatedAt:   project.CreatedAt,
			Storyboards: storyboards,
		},
		Models:               config.GetModels(),
		AudioSupportedModels: config.GetAudioSupportedModelIDs(),
	}, nil
}

// ============================================================
// Storyboards
// ============================================================

// CreateStoryboardParams holds parameters for creating a storyboard
type CreateStoryboardParams struct {
	ProjectID      uint   `json:"project_id"`
	Prompt         string `json:"prompt"`
	ModelID        string `json:"model_id"`
	Ratio          string `json:"ratio"`
	Duration       int    `json:"duration"`
	GenerateAudio  bool   `json:"generate_audio"`
	ServiceTier    string `json:"service_tier"`
	FirstFramePath string `json:"first_frame_path"`
	LastFramePath  string `json:"last_frame_path"`
}

// CreateStoryboard creates a new storyboard with initial take
func (a *App) CreateStoryboard(params CreateStoryboardParams) error {
	os.MkdirAll(config.UploadsDir(), 0755)

	if params.ServiceTier == "" {
		params.ServiceTier = "standard"
	}

	storyboard := models.Storyboard{
		ProjectID: params.ProjectID,
		CreatedAt: time.Now(),
	}
	models.DB.Create(&storyboard)

	take := models.Take{
		StoryboardID:  storyboard.ID,
		Prompt:        params.Prompt,
		ModelID:       params.ModelID,
		Ratio:         params.Ratio,
		Duration:      params.Duration,
		GenerateAudio: params.GenerateAudio,
		ServiceTier:   params.ServiceTier,
		Status:        "Draft",
		CreatedAt:     time.Now(),
	}

	if params.ServiceTier == "flex" {
		take.ExpiresAfter = 86400
	}

	// Use file paths directly (files already saved via SelectImageFile)
	if params.FirstFramePath != "" {
		take.FirstFramePath = params.FirstFramePath
	}
	if params.LastFramePath != "" {
		take.LastFramePath = params.LastFramePath
	}

	return models.DB.Create(&take).Error
}

// DeleteStoryboard deletes a storyboard and returns its project ID
func (a *App) DeleteStoryboard(id uint) (uint, error) {
	var sb models.Storyboard
	if err := models.DB.First(&sb, id).Error; err != nil {
		return 0, err
	}
	projectID := sb.ProjectID
	return projectID, models.DB.Delete(&sb).Error
}

// UpdateStoryboardParams holds parameters for updating a storyboard
type UpdateStoryboardParams struct {
	StoryboardID     uint   `json:"storyboard_id"`
	Prompt           string `json:"prompt"`
	ModelID          string `json:"model_id"`
	Ratio            string `json:"ratio"`
	Duration         int    `json:"duration"`
	GenerateAudio    bool   `json:"generate_audio"`
	ServiceTier      string `json:"service_tier"`
	FirstFramePath   string `json:"first_frame_path"`
	LastFramePath    string `json:"last_frame_path"`
	DeleteFirstFrame bool   `json:"delete_first_frame"`
	DeleteLastFrame  bool   `json:"delete_last_frame"`
}

// UpdateStoryboard creates a new take version for the storyboard
func (a *App) UpdateStoryboard(params UpdateStoryboardParams) error {
	os.MkdirAll(config.UploadsDir(), 0755)

	var sb models.Storyboard
	if err := models.DB.Preload("Takes", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).First(&sb, params.StoryboardID).Error; err != nil {
		return fmt.Errorf("storyboard not found")
	}

	var prevTake models.Take
	if len(sb.Takes) > 0 {
		prevTake = sb.Takes[len(sb.Takes)-1]
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

	if params.Prompt != "" {
		newTake.Prompt = params.Prompt
	}
	if params.ModelID != "" {
		newTake.ModelID = params.ModelID
	}
	if params.Ratio != "" {
		newTake.Ratio = params.Ratio
	}
	if params.Duration > 0 {
		newTake.Duration = params.Duration
	}
	newTake.GenerateAudio = params.GenerateAudio
	if params.ServiceTier != "" {
		newTake.ServiceTier = params.ServiceTier
		if params.ServiceTier == "flex" {
			newTake.ExpiresAfter = 86400
		} else {
			newTake.ExpiresAfter = 0
		}
	}

	// Handle deletion flags
	if params.DeleteFirstFrame {
		newTake.FirstFramePath = ""
	}
	if params.DeleteLastFrame {
		newTake.LastFramePath = ""
	}

	// Handle new file paths (override copied paths)
	if params.FirstFramePath != "" {
		newTake.FirstFramePath = params.FirstFramePath
	}
	if params.LastFramePath != "" {
		newTake.LastFramePath = params.LastFramePath
	}

	return models.DB.Create(&newTake).Error
}

// ============================================================
// Takes
// ============================================================

// GenerateTakeVideo starts video generation for a take
func (a *App) GenerateTakeVideo(id uint) (map[string]interface{}, error) {
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		return nil, fmt.Errorf("take not found")
	}

	var firstFrameURL, lastFrameURL string
	if take.FirstFramePath != "" {
		b64, err := imageToBase64(take.FirstFramePath)
		if err != nil {
			return nil, fmt.Errorf("failed to process first frame: %w", err)
		}
		firstFrameURL = b64
	}
	if take.LastFramePath != "" {
		b64, err := imageToBase64(take.LastFramePath)
		if err != nil {
			return nil, fmt.Errorf("failed to process last frame: %w", err)
		}
		lastFrameURL = b64
	}

	taskID, err := a.volcService.CreateVideoTask(
		take.ModelID, take.Prompt, firstFrameURL, lastFrameURL,
		take.Ratio, take.Duration, take.GenerateAudio,
		take.ServiceTier, take.ExpiresAfter,
	)
	if err != nil {
		take.Status = "Failed"
		models.DB.Save(&take)
		return nil, err
	}

	take.TaskID = taskID
	take.Status = "Running"
	models.DB.Save(&take)

	return map[string]interface{}{
		"status":  "submitted",
		"task_id": taskID,
	}, nil
}

// TakeStatusResult holds the status polling result
type TakeStatusResult struct {
	Status         string `json:"status"`
	VideoURL       string `json:"video_url"`
	LastFrameURL   string `json:"last_frame_url"`
	PollInterval   int    `json:"poll_interval"`
	DownloadStatus string `json:"download_status"`
}

// GetTakeStatus polls the status of a take's video generation
func (a *App) GetTakeStatus(id uint) (*TakeStatusResult, error) {
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		return nil, fmt.Errorf("take not found")
	}

	if take.TaskID == "" {
		return &TakeStatusResult{Status: take.Status}, nil
	}

	pollInterval := 3000
	if take.ServiceTier == "flex" {
		if time.Since(take.CreatedAt) > 10*time.Minute {
			pollInterval = 60000
		} else {
			pollInterval = 10000
		}
	}

	resp, err := a.volcService.GetTaskStatus(take.TaskID)
	if err != nil {
		return &TakeStatusResult{
			Status:       take.Status,
			PollInterval: pollInterval,
		}, nil
	}

	previousStatus := take.Status
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

		if previousStatus != "Succeeded" && take.DownloadStatus != "completed" {
			take.DownloadStatus = "pending"
			models.DB.Save(&take)
			services.DownloadTakeAssetsAsync(take.ID)
		}
	}

	models.DB.Save(&take)

	return &TakeStatusResult{
		Status:         take.Status,
		VideoURL:       services.GetEffectiveVideoURL(&take),
		LastFrameURL:   services.GetEffectiveLastFrameURL(&take),
		PollInterval:   pollInterval,
		DownloadStatus: take.DownloadStatus,
	}, nil
}

// GetTake returns a single take by ID
func (a *App) GetTake(id uint) (*TakeResponse, error) {
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		return nil, fmt.Errorf("take not found")
	}
	resp := takeToResponse(&take)
	return &resp, nil
}

// ListTakes returns all takes for a storyboard
func (a *App) ListTakes(storyboardID uint) ([]TakeResponse, error) {
	var takes []models.Take
	if err := models.DB.Where("storyboard_id = ?", storyboardID).Order("created_at asc").Find(&takes).Error; err != nil {
		return nil, err
	}
	var result []TakeResponse
	for i := range takes {
		result = append(result, takeToResponse(&takes[i]))
	}
	return result, nil
}

// ToggleGoodTake toggles the "good take" marker
func (a *App) ToggleGoodTake(id uint) (bool, error) {
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		return false, fmt.Errorf("take not found")
	}

	newState := !take.IsGood
	if newState {
		models.DB.Model(&models.Take{}).Where("storyboard_id = ? AND id != ?", take.StoryboardID, take.ID).Update("is_good", false)
	}
	take.IsGood = newState
	models.DB.Save(&take)

	return take.IsGood, nil
}

// DeleteTakeResult holds the result of deleting a take
type DeleteTakeResult struct {
	Success           bool  `json:"success"`
	StoryboardDeleted bool  `json:"storyboard_deleted"`
	RemainingTakes    int64 `json:"remaining_takes"`
}

// DeleteTake deletes a take, and its storyboard if empty
func (a *App) DeleteTake(id uint) (*DeleteTakeResult, error) {
	var take models.Take
	if err := models.DB.First(&take, id).Error; err != nil {
		return nil, fmt.Errorf("take not found")
	}
	storyboardID := take.StoryboardID

	if err := models.DB.Delete(&take).Error; err != nil {
		return nil, fmt.Errorf("failed to delete take")
	}

	var count int64
	models.DB.Model(&models.Take{}).Where("storyboard_id = ?", storyboardID).Count(&count)

	storyboardDeleted := false
	if count == 0 {
		if err := models.DB.Delete(&models.Storyboard{}, storyboardID).Error; err == nil {
			storyboardDeleted = true
		}
	}

	return &DeleteTakeResult{
		Success:           true,
		StoryboardDeleted: storyboardDeleted,
		RemainingTakes:    count,
	}, nil
}

// ============================================================
// Export
// ============================================================

// ExportProject exports project videos as a ZIP with FCPXML via save dialog
func (a *App) ExportProject(id uint) error {
	var project models.Project
	if err := models.DB.Preload("Storyboards.Takes").First(&project, id).Error; err != nil {
		return fmt.Errorf("project not found")
	}

	exports := services.PrepareExportData(project.Storyboards)
	if len(exports) == 0 {
		return fmt.Errorf("no succeeded videos available for export")
	}

	filename := services.GetExportFilename(project.Name)

	savePath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		DefaultFilename: filename,
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "ZIP Files (*.zip)", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return err
	}
	if savePath == "" {
		return nil // User cancelled
	}

	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return services.CreateExportZIP(file, project.Name, exports)
}

// ============================================================
// File Operations
// ============================================================

// SelectImageFile opens a native file dialog to select an image
func (a *App) SelectImageFile() (string, error) {
	result, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Image",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Images (*.png, *.jpg, *.jpeg, *.gif, *.webp)", Pattern: "*.png;*.jpg;*.jpeg;*.gif;*.webp"},
		},
	})
	if err != nil {
		return "", err
	}
	if result == "" {
		return "", nil
	}

	ensureDir("uploads")
	filename := uuid.New().String() + filepath.Ext(result)
	dst := filepath.Join(config.UploadsDir(), filename)

	srcFile, err := os.Open(result)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", err
	}

	// Return relative path for DB storage
	return config.ToRelativePath(dst), nil
}

// CopyToUploads copies a local file to the uploads directory and returns the new path.
// Useful for "Use last frame as first frame" where the source is in downloads/.
func (a *App) CopyToUploads(srcPath string) (string, error) {
	ensureDir("uploads")

	// Strip leading slash if present (from URL paths like /downloads/xxx.png)
	srcPath = strings.TrimPrefix(srcPath, "/")
	// Resolve to absolute path in data dir
	srcPath = config.ToAbsolutePath(srcPath)

	if _, err := os.Stat(srcPath); err != nil {
		return "", fmt.Errorf("source file not found: %w", err)
	}

	ext := filepath.Ext(srcPath)
	if ext == "" {
		ext = ".png"
	}
	filename := uuid.New().String() + ext
	dst := filepath.Join(config.UploadsDir(), filename)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", err
	}

	// Return relative path for DB storage
	return config.ToRelativePath(dst), nil
}

// ============================================================
// Helper Functions
// ============================================================

func imageToBase64(path string) (string, error) {
	// Resolve relative path to absolute in data directory
	absPath := config.ToAbsolutePath(path)
	file, err := os.Open(absPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	mimeType := "image/png"
	lower := strings.ToLower(path)
	if strings.HasSuffix(lower, ".jpg") || strings.HasSuffix(lower, ".jpeg") {
		mimeType = "image/jpeg"
	}

	b64 := base64.StdEncoding.EncodeToString(bytes)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, b64), nil
}

func ensureDir(dir string) {
	absDir := config.ToAbsolutePath(dir)
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		os.MkdirAll(absDir, 0755)
	}
}
