package services

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"seedance-client/models"
	"strings"
)

// FCPXML 1.9 compatible structures for DaVinci Resolve
type FCPXML struct {
	XMLName   xml.Name  `xml:"fcpxml"`
	Version   string    `xml:"version,attr"`
	Resources Resources `xml:"resources"`
	Library   Library   `xml:"library"`
}

type Resources struct {
	Format Format  `xml:"format"`
	Assets []Asset `xml:"asset"`
}

type Format struct {
	ID            string `xml:"id,attr"`
	Name          string `xml:"name,attr"`
	FrameDuration string `xml:"frameDuration,attr"`
	Width         int    `xml:"width,attr"`
	Height        int    `xml:"height,attr"`
}

type Asset struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Src      string `xml:"src,attr"`
	Start    string `xml:"start,attr"`
	Duration string `xml:"duration,attr"`
	HasVideo int    `xml:"hasVideo,attr"`
	HasAudio int    `xml:"hasAudio,attr"`
	Format   string `xml:"format,attr"`
}

type Library struct {
	Event Event `xml:"event"`
}

type Event struct {
	Name    string  `xml:"name,attr"`
	Project Project `xml:"project"`
}

type Project struct {
	Name     string   `xml:"name,attr"`
	Sequence Sequence `xml:"sequence"`
}

type Sequence struct {
	Format string `xml:"format,attr"`
	Spine  Spine  `xml:"spine"`
}

type Spine struct {
	Clips []Clip `xml:"clip"`
}

type Clip struct {
	Name     string `xml:"name,attr"`
	Offset   string `xml:"offset,attr"`
	Duration string `xml:"duration,attr"`
	Start    string `xml:"start,attr"`
	Ref      string `xml:"ref,attr"`
}

// ExportData holds information about files to be included in the export
type ExportData struct {
	Filename string
	VideoURL string
	Duration int
}

// sanitizeFilename removes special characters and limits length for filenames
func sanitizeFilename(s string) string {
	// Replace non-alphanumeric characters (except spaces) with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s\u4e00-\u9fff]`)
	s = reg.ReplaceAllString(s, "")
	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Limit to 30 characters
	if len(s) > 30 {
		s = s[:30]
	}
	// Remove trailing underscores
	s = strings.TrimRight(s, "_")
	if s == "" {
		s = "video"
	}
	return s
}

// GenerateFCPXML creates FCPXML 1.9 content for DaVinci Resolve
func GenerateFCPXML(projectName string, exports []ExportData) ([]byte, error) {
	// Create format (24fps, 1280x720)
	format := Format{
		ID:            "r0",
		Name:          "FFVideoFormat720p24",
		FrameDuration: "100/2400s", // 24fps
		Width:         1280,
		Height:        720,
	}

	// Create assets and clips
	var assets []Asset
	var clips []Clip
	var offsetSeconds int

	for i, exp := range exports {
		assetID := fmt.Sprintf("r%d", i+1)
		clipName := strings.TrimSuffix(exp.Filename, ".mp4")

		// Asset
		asset := Asset{
			ID:       assetID,
			Name:     clipName,
			Src:      "./" + exp.Filename,
			Start:    "0s",
			Duration: fmt.Sprintf("%ds", exp.Duration),
			HasVideo: 1,
			HasAudio: 1,
			Format:   "r0",
		}
		assets = append(assets, asset)

		// Clip
		clip := Clip{
			Name:     clipName,
			Offset:   fmt.Sprintf("%ds", offsetSeconds),
			Duration: fmt.Sprintf("%ds", exp.Duration),
			Start:    "0s",
			Ref:      assetID,
		}
		clips = append(clips, clip)

		offsetSeconds += exp.Duration
	}

	fcpxml := FCPXML{
		Version: "1.9",
		Resources: Resources{
			Format: format,
			Assets: assets,
		},
		Library: Library{
			Event: Event{
				Name: "AI_Generated",
				Project: Project{
					Name: projectName,
					Sequence: Sequence{
						Format: "r0",
						Spine: Spine{
							Clips: clips,
						},
					},
				},
			},
		},
	}

	// Marshal with XML header
	output, err := xml.MarshalIndent(fcpxml, "", "    ")
	if err != nil {
		return nil, err
	}

	header := []byte(xml.Header + "<!DOCTYPE fcpxml>\n")
	return append(header, output...), nil
}

// PrepareExportData generates the list of files to export from succeeded storyboards
func PrepareExportData(storyboards []models.Storyboard) []ExportData {
	var exports []ExportData
	index := 1

	for _, sb := range storyboards {
		// Only include succeeded storyboards with valid video URLs
		if sb.Status != "Succeeded" || sb.VideoURL == "" {
			continue
		}

		// Create filename: {index}_{sanitized_prompt}.mp4
		sanitized := sanitizeFilename(sb.Prompt)
		filename := fmt.Sprintf("%03d_%s.mp4", index, sanitized)

		exports = append(exports, ExportData{
			Filename: filename,
			VideoURL: sb.VideoURL,
			Duration: sb.Duration,
		})
		index++
	}

	return exports
}

// CreateExportZIP creates a ZIP file containing all videos and FCPXML
func CreateExportZIP(w io.Writer, projectName string, exports []ExportData) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Generate and add FCPXML
	fcpxmlData, err := GenerateFCPXML(projectName, exports)
	if err != nil {
		return fmt.Errorf("failed to generate FCPXML: %w", err)
	}

	fcpxmlFile, err := zipWriter.Create("project.fcpxml")
	if err != nil {
		return fmt.Errorf("failed to create fcpxml in zip: %w", err)
	}
	if _, err := fcpxmlFile.Write(fcpxmlData); err != nil {
		return fmt.Errorf("failed to write fcpxml: %w", err)
	}

	// Download and add each video
	client := &http.Client{}
	for _, exp := range exports {
		if err := addVideoToZip(zipWriter, client, exp.Filename, exp.VideoURL); err != nil {
			return fmt.Errorf("failed to add video %s: %w", exp.Filename, err)
		}
	}

	return nil
}

// addVideoToZip downloads a video and adds it to the ZIP
func addVideoToZip(zw *zip.Writer, client *http.Client, filename, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download video: status %d", resp.StatusCode)
	}

	// Create file in zip
	videoFile, err := zw.Create(filename)
	if err != nil {
		return err
	}

	// Stream video content to zip
	_, err = io.Copy(videoFile, resp.Body)
	return err
}

// ExportProjectToZip is a convenience function that handles the full export process
func ExportProjectToZip(w io.Writer, project models.Project) error {
	exports := PrepareExportData(project.Storyboards)

	if len(exports) == 0 {
		return fmt.Errorf("no succeeded videos available for export")
	}

	return CreateExportZIP(w, project.Name, exports)
}

// GetExportFilename returns a safe filename for the ZIP download
func GetExportFilename(projectName string) string {
	safe := sanitizeFilename(projectName)
	return safe + "_export.zip"
}

// Cleanup removes temporary files (if we ever need a temp dir approach in future)
func Cleanup(tempDir string) error {
	if tempDir != "" {
		return os.RemoveAll(tempDir)
	}
	return nil
}

// GetTempDir creates a temporary directory for export operations
func GetTempDir() (string, error) {
	return os.MkdirTemp("", "seedance-export-*")
}

// DownloadVideo downloads a video to a local file
func DownloadVideo(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// CopyFileToZip copies a local file into a zip archive
func CopyFileToZip(zw *zip.Writer, srcPath, destName string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zw.Create(destName)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}

// GetVideoFilename creates the export filename for a storyboard
func GetVideoFilename(index int, prompt string) string {
	sanitized := sanitizeFilename(prompt)
	return fmt.Sprintf("%03d_%s.mp4", index, sanitized)
}

// EnsureDir creates directory if it doesn't exist
func EnsureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// WriteToFile writes content to a file
func WriteToFile(path string, content []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}
