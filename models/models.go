package models

import (
	"time"

	"gorm.io/gorm"
)

// Model version constants
const (
	ModelVersionV1 = "v1.x" // Seedance 1.0/1.5 and earlier
	ModelVersionV2 = "v2.0" // Seedance 2.0
)

// ValidModelVersions returns the list of valid model versions
func ValidModelVersions() []string {
	return []string{ModelVersionV1, ModelVersionV2}
}

// IsValidModelVersion checks if a model version string is valid
func IsValidModelVersion(v string) bool {
	for _, valid := range ValidModelVersions() {
		if v == valid {
			return true
		}
	}
	return false
}

type Project struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	Name         string       `json:"name"`
	ModelVersion string       `gorm:"default:v1.x" json:"model_version"` // "v1.x" or "v2.0"
	AspectRatio  string       `gorm:"default:16:9" json:"aspect_ratio"`  // Fixed ratio for the project
	CreatedAt    time.Time    `json:"created_at"`
	Storyboards  []Storyboard `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"storyboards"`
}

type Storyboard struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID uint      `json:"project_id"`
	Takes     []Take    `gorm:"foreignKey:StoryboardID;constraint:OnDelete:CASCADE;" json:"takes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// V1 storyboard structured metadata
	ShotOrder         int    `gorm:"index" json:"shot_order"`
	ShotNo            string `json:"shot_no"`
	ShotSize          string `json:"shot_size"`
	CameraMovement    string `json:"camera_movement"`
	FrameContent      string `gorm:"type:text" json:"frame_content"`
	CharactersJSON    string `gorm:"type:text" json:"characters_json"` // []EntityRef
	ScenesJSON        string `gorm:"type:text" json:"scenes_json"`     // []EntityRef
	ElementsJSON      string `gorm:"type:text" json:"elements_json"`   // []EntityRef
	StylesJSON        string `gorm:"type:text" json:"styles_json"`     // []EntityRef
	SoundDesign       string `gorm:"type:text" json:"sound_design"`
	EstimatedDuration int    `gorm:"default:5" json:"estimated_duration"` // 5/10 for now
	DurationFine      int    `gorm:"default:0" json:"duration_fine"`      // reserved for fine control

	// Virtual field for the active take (not stored in DB, populated during query)
	ActiveTake *Take `gorm:"-" json:"active_take,omitempty"`
}

type Take struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	StoryboardID       uint      `json:"storyboard_id"`
	Prompt             string    `json:"prompt"`
	FirstFramePath     string    `json:"first_frame_path"`      // Local path (uploaded)
	LastFramePath      string    `json:"last_frame_path"`       // Local path (uploaded)
	ModelID            string    `json:"model_id"`              // e.g. "doubao-seedance-1-5-pro-251215"
	Ratio              string    `json:"ratio"`                 // "16:9", "adaptive"
	Duration           int       `json:"duration"`              // 5
	GenerateAudio      bool      `json:"generate_audio"`        // false
	TaskID             string    `json:"task_id"`               // Volcano Task ID
	Status             string    `json:"status"`                // Queued, Running, Succeeded, Failed
	VideoURL           string    `json:"video_url"`             // Remote Result URL
	LastFrameURL       string    `json:"last_frame_url"`        // Remote Result Last Frame URL
	LocalVideoPath     string    `json:"local_video_path"`      // Local cached video path
	LocalLastFramePath string    `json:"local_last_frame_path"` // Local cached last frame path
	DownloadStatus     string    `json:"download_status"`       // pending, downloading, completed, failed
	ServiceTier        string    `json:"service_tier"`          // "standard" or "flex"
	TokenUsage         int       `json:"token_usage"`           // Usage.CompletionTokens
	ExpiresAfter       int64     `json:"expires_after"`
	IsGood             bool      `json:"is_good"` // "Good Take" marker
	ChainFromPrev      bool      `json:"chain_from_prev"`
	GenerationMode     string    `gorm:"default:standard" json:"generation_mode"` // standard / flat
	CreatedAt          time.Time `json:"created_at"`
}

// AssetCatalog is a project-level reusable asset prompt definition.
// AssetType: character | scene | element | style
type AssetCatalog struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ProjectID    uint           `gorm:"index:idx_catalog_project_type_code,priority:1;index" json:"project_id"`
	AssetType    string         `gorm:"index:idx_catalog_project_type_code,priority:2;index" json:"asset_type"`
	AssetCode    string         `gorm:"index:idx_catalog_project_type_code,priority:3" json:"asset_code"` // user-facing id
	Name         string         `json:"name"`
	Prompt       string         `gorm:"type:text" json:"prompt"`
	StoryboardID *uint          `gorm:"index" json:"storyboard_id,omitempty"` // optional source storyboard
	Versions     []AssetVersion `gorm:"foreignKey:CatalogID;constraint:OnDelete:CASCADE;" json:"versions"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type AssetVersion struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	CatalogID  uint      `gorm:"index" json:"catalog_id"`
	VersionNo  int       `json:"version_no"`
	ImagePath  string    `json:"image_path"`  // local cached image path
	SourceType string    `json:"source_type"` // generated / uploaded
	ModelID    string    `json:"model_id"`
	Prompt     string    `gorm:"type:text" json:"prompt"`
	TaskID     string    `json:"task_id"`
	Status     string    `json:"status"` // Draft / Running / Succeeded / Failed
	IsGood     bool      `json:"is_good"`
	CreatedAt  time.Time `json:"created_at"`
}

// ShotFrameVersion stores start/end frames per storyboard with versioning.
type ShotFrameVersion struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	StoryboardID uint      `gorm:"index" json:"storyboard_id"`
	FrameType    string    `gorm:"index" json:"frame_type"` // start / end
	VersionNo    int       `json:"version_no"`
	ImagePath    string    `json:"image_path"`  // local cached image path
	SourceType   string    `json:"source_type"` // generated / uploaded
	ModelID      string    `json:"model_id"`
	Prompt       string    `gorm:"type:text" json:"prompt"`
	TaskID       string    `json:"task_id"`
	Status       string    `json:"status"` // Draft / Running / Succeeded / Failed
	IsGood       bool      `json:"is_good"`
	CreatedAt    time.Time `json:"created_at"`
}

// Setting stores key-value configuration (e.g. API key)
type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `json:"value"`
}

var DB *gorm.DB
