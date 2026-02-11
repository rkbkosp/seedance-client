package main

import (
	"embed"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"seedance-client/config"
	"seedance-client/models"
	"seedance-client/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Initialize data directory (must be first)
	config.InitDataDir()

	// Initialize Database
	models.InitDB()

	// Start background asset downloader
	services.StartBackgroundDownloader()

	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "Spark (火种)",
		Width:            1280,
		Height:           850,
		MinWidth:         900,
		MinHeight:        600,
		DisableResize:    false,
		Fullscreen:       false,
		WindowStartState: options.Normal,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: NewFileHandler(),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}

// FileHandler serves local upload/download files from disk
type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path

	// Serve local files from uploads/ and downloads/ directories
	if strings.HasPrefix(urlPath, "/uploads/") || strings.HasPrefix(urlPath, "/downloads/") {
		// Get the relative part (uploads/xxx.png or downloads/xxx.mp4)
		relPath := filepath.Clean(strings.TrimPrefix(urlPath, "/"))

		// Security: ensure the path doesn't escape these directories
		if !strings.HasPrefix(relPath, "uploads") && !strings.HasPrefix(relPath, "downloads") {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Resolve to absolute path in the data directory
		absPath := config.ToAbsolutePath(relPath)

		if _, err := os.Stat(absPath); err == nil {
			http.ServeFile(w, r, absPath)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
