package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `json:"name"`
	CreatedAt   time.Time    `json:"created_at"`
	Storyboards []Storyboard `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE;" json:"storyboards"`
}

type Storyboard struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProjectID      uint      `json:"project_id"`
	Prompt         string    `json:"prompt"`
	FirstFramePath string    `json:"first_frame_path"` // Local path
	LastFramePath  string    `json:"last_frame_path"`  // Local path
	ModelID        string    `json:"model_id"`         // e.g. "doubao-seedance-1-5-pro-251215"
	Ratio          string    `json:"ratio"`            // "16:9", "adaptive"
	Duration       int       `json:"duration"`         // 5
	GenerateAudio  bool      `json:"generate_audio"`   // false
	TaskID         string    `json:"task_id"`          // Volcano Task ID
	Status         string    `json:"status"`           // Queued, Running, Succeeded, Failed
	VideoURL       string    `json:"video_url"`        // Result URL
	LastFrameURL   string    `json:"last_frame_url"`   // Result Last Frame URL
	CreatedAt      time.Time `json:"created_at"`
}

var DB *gorm.DB
