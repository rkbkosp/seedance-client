package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"seedance-client/models"
	"time"

	"github.com/google/uuid"
)

const DownloadsDir = "downloads"

// ensureDownloadsDir creates the downloads directory if it doesn't exist
func ensureDownloadsDir() {
	if _, err := os.Stat(DownloadsDir); os.IsNotExist(err) {
		os.MkdirAll(DownloadsDir, 0755)
	}
}

// DownloadAsset downloads a file from URL and saves it locally
func DownloadAsset(url string, ext string) (string, error) {
	ensureDownloadsDir()

	filename := uuid.New().String() + ext
	localPath := filepath.Join(DownloadsDir, filename)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(localPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return localPath, nil
}

// DownloadTakeAssets downloads video and last frame for a take
func DownloadTakeAssets(take *models.Take) error {
	if take.Status != "Succeeded" || take.VideoURL == "" {
		return nil
	}

	// Skip if already completed
	if take.DownloadStatus == "completed" {
		// Verify files exist
		if take.LocalVideoPath != "" {
			if _, err := os.Stat(take.LocalVideoPath); err == nil {
				return nil // Already downloaded and exists
			}
		}
	}

	// Mark as downloading
	take.DownloadStatus = "downloading"
	models.DB.Save(take)

	// Download video
	if take.VideoURL != "" && take.LocalVideoPath == "" {
		localPath, err := DownloadAsset(take.VideoURL, ".mp4")
		if err != nil {
			take.DownloadStatus = "failed"
			models.DB.Save(take)
			return fmt.Errorf("video download failed: %w", err)
		}
		take.LocalVideoPath = localPath
	}

	// Download last frame
	if take.LastFrameURL != "" && take.LocalLastFramePath == "" {
		localPath, err := DownloadAsset(take.LastFrameURL, ".png")
		if err != nil {
			// Video downloaded but frame failed - still mark partial success
			log.Printf("Last frame download failed for take %d: %v", take.ID, err)
		} else {
			take.LocalLastFramePath = localPath
		}
	}

	take.DownloadStatus = "completed"
	models.DB.Save(take)
	return nil
}

// DownloadTakeAssetsAsync downloads assets in background
func DownloadTakeAssetsAsync(takeID uint) {
	go func() {
		var take models.Take
		if err := models.DB.First(&take, takeID).Error; err != nil {
			log.Printf("Failed to find take %d for download: %v", takeID, err)
			return
		}
		if err := DownloadTakeAssets(&take); err != nil {
			log.Printf("Failed to download assets for take %d: %v", takeID, err)
		} else {
			log.Printf("Downloaded assets for take %d", takeID)
		}
	}()
}

// ScanAndDownloadMissing scans all succeeded takes and downloads missing assets
func ScanAndDownloadMissing() {
	log.Println("Starting asset scan...")

	var takes []models.Take
	models.DB.Where("status = ?", "Succeeded").Find(&takes)

	downloadCount := 0
	for _, take := range takes {
		needsDownload := false

		// Check if video needs download
		if take.VideoURL != "" {
			if take.LocalVideoPath == "" {
				needsDownload = true
			} else if _, err := os.Stat(take.LocalVideoPath); os.IsNotExist(err) {
				needsDownload = true
				take.LocalVideoPath = "" // Reset so it gets re-downloaded
			}
		}

		// Check if last frame needs download
		if take.LastFrameURL != "" {
			if take.LocalLastFramePath == "" {
				needsDownload = true
			} else if _, err := os.Stat(take.LocalLastFramePath); os.IsNotExist(err) {
				needsDownload = true
				take.LocalLastFramePath = "" // Reset
			}
		}

		// Check pending/failed downloads
		if take.DownloadStatus == "pending" || take.DownloadStatus == "failed" || take.DownloadStatus == "" {
			needsDownload = true
		}

		if needsDownload {
			if err := DownloadTakeAssets(&take); err != nil {
				log.Printf("Scan: failed to download take %d: %v", take.ID, err)
			} else {
				downloadCount++
			}
		}
	}

	log.Printf("Asset scan complete. Downloaded %d assets.", downloadCount)
}

// StartBackgroundDownloader starts the background asset scanner
func StartBackgroundDownloader() {
	ensureDownloadsDir()

	// Run initial scan after a short delay
	go func() {
		time.Sleep(2 * time.Second)
		ScanAndDownloadMissing()
	}()
}

// GetEffectiveVideoURL returns local path if available, otherwise remote URL
func GetEffectiveVideoURL(take *models.Take) string {
	if take.LocalVideoPath != "" {
		if _, err := os.Stat(take.LocalVideoPath); err == nil {
			return "/" + take.LocalVideoPath
		}
	}
	return take.VideoURL
}

// GetEffectiveLastFrameURL returns local path if available, otherwise remote URL
func GetEffectiveLastFrameURL(take *models.Take) string {
	if take.LocalLastFramePath != "" {
		if _, err := os.Stat(take.LocalLastFramePath); err == nil {
			return "/" + take.LocalLastFramePath
		}
	}
	return take.LastFrameURL
}
