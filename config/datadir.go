package config

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// dataDir is the resolved application data directory
var dataDir string

// InitDataDir initializes the application data directory.
// On macOS:  ~/Library/Application Support/seedance-client
// On Windows: %APPDATA%/seedance-client
// On Linux:  ~/.local/share/seedance-client
func InitDataDir() {
	dataDir = resolveDataDir()
	os.MkdirAll(dataDir, 0755)
	os.MkdirAll(filepath.Join(dataDir, "uploads"), 0755)
	os.MkdirAll(filepath.Join(dataDir, "downloads"), 0755)

	// Auto-migrate old data from working directory (if exists)
	migrateOldData()

	log.Printf("Data directory: %s", dataDir)
}

// GetDataDir returns the application data directory
func GetDataDir() string {
	return dataDir
}

// UploadsDir returns the full path to the uploads directory
func UploadsDir() string {
	return filepath.Join(dataDir, "uploads")
}

// DownloadsDir returns the full path to the downloads directory
func DownloadsDir() string {
	return filepath.Join(dataDir, "downloads")
}

// DBPath returns the full path to the database file
func DBPath() string {
	return filepath.Join(dataDir, "seedance.db")
}

// UploadsPath returns the full path for a file in uploads
func UploadsPath(filename string) string {
	return filepath.Join(dataDir, "uploads", filename)
}

// DownloadsPath returns the full path for a file in downloads
func DownloadsPath(filename string) string {
	return filepath.Join(dataDir, "downloads", filename)
}

// ToRelativePath converts an absolute data dir path to a relative path (uploads/xxx.png)
// for URL serving. Returns the path unchanged if it's already relative.
func ToRelativePath(absPath string) string {
	if rel, err := filepath.Rel(dataDir, absPath); err == nil {
		return rel
	}
	return absPath
}

// ToAbsolutePath converts a relative path (uploads/xxx.png) to an absolute path
// under the data directory. If the path is already absolute, returns it unchanged.
func ToAbsolutePath(relPath string) string {
	relPath = strings.TrimPrefix(relPath, "/")
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(dataDir, relPath)
}

func resolveDataDir() string {
	// Allow override via environment variable
	if dir := os.Getenv("SEEDANCE_DATA_DIR"); dir != "" {
		return dir
	}

	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "seedance-client")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "seedance-client")
	default: // linux and others
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return filepath.Join(xdg, "seedance-client")
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", "seedance-client")
	}
}

// migrateOldData checks for old data in the working directory and moves it
// to the new data directory. This runs once on first launch after upgrade.
func migrateOldData() {
	// Check if migration marker exists
	markerPath := filepath.Join(dataDir, ".migrated")
	if _, err := os.Stat(markerPath); err == nil {
		return // Already migrated
	}

	migrated := false

	// Migrate old DB
	if _, err := os.Stat("seedance.db"); err == nil {
		newDB := DBPath()
		if _, err := os.Stat(newDB); os.IsNotExist(err) {
			if copyFile("seedance.db", newDB) == nil {
				log.Println("Migrated seedance.db to data directory")
				migrated = true
			}
		}
	}

	// Migrate old uploads
	if entries, err := os.ReadDir("uploads"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			src := filepath.Join("uploads", entry.Name())
			dst := UploadsPath(entry.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if copyFile(src, dst) == nil {
					migrated = true
				}
			}
		}
		if migrated {
			log.Println("Migrated uploads/ to data directory")
		}
	}

	// Migrate old downloads
	dlMigrated := false
	if entries, err := os.ReadDir("downloads"); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			src := filepath.Join("downloads", entry.Name())
			dst := DownloadsPath(entry.Name())
			if _, err := os.Stat(dst); os.IsNotExist(err) {
				if copyFile(src, dst) == nil {
					dlMigrated = true
				}
			}
		}
		if dlMigrated {
			log.Println("Migrated downloads/ to data directory")
			migrated = true
		}
	}

	if migrated {
		// Create migration marker
		os.WriteFile(markerPath, []byte("migrated"), 0644)
		log.Printf("Old data migrated to: %s", dataDir)
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
