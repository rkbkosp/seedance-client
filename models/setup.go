package models

import (
	"seedance-client/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open(config.DBPath()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(&Project{}, &Storyboard{}, &Take{}, &Setting{})

	// Migrate existing Storyboard data to Takes
	// We use raw SQL to avoid needing the old struct definition
	// Copy data from storyboards table to takes table for storyboards that don't have takes yet
	DB.Exec(`
		INSERT INTO takes (
			storyboard_id, prompt, first_frame_path, last_frame_path, model_id, 
			ratio, duration, generate_audio, task_id, status, video_url, 
			last_frame_url, service_tier, token_usage, expires_after, created_at, is_good
		)
		SELECT 
			id, prompt, first_frame_path, last_frame_path, model_id, 
			ratio, duration, generate_audio, task_id, status, video_url, 
			last_frame_url, service_tier, token_usage, expires_after, created_at, 0
		FROM storyboards 
		WHERE id NOT IN (SELECT DISTINCT storyboard_id FROM takes)
		  AND prompt IS NOT NULL AND prompt != '' -- Ensure we don't migrate empty/already migrated rows if columns cleared
	`)
}
