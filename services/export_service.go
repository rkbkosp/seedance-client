package services

import (
	"archive/zip"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"math"
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
func GenerateFCPXML(projectName string, exports []ExportData, width, height int, frameDuration string) ([]byte, error) {
	// Create format
	format := Format{
		ID:            "r0",
		Name:          fmt.Sprintf("FFVideoFormat%dp", height),
		FrameDuration: frameDuration,
		Width:         width,
		Height:        height,
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
		// Find best take
		var bestTake *models.Take
		for i := len(sb.Takes) - 1; i >= 0; i-- {
			if sb.Takes[i].IsGood {
				bestTake = &sb.Takes[i]
				break
			}
		}
		if bestTake == nil && len(sb.Takes) > 0 {
			bestTake = &sb.Takes[len(sb.Takes)-1]
		}

		if bestTake == nil {
			continue
		}

		// Only include succeeded storyboards with valid video URLs
		if bestTake.Status != "Succeeded" || bestTake.VideoURL == "" {
			continue
		}

		// Create filename: {index}_{sanitized_prompt}.mp4
		sanitized := sanitizeFilename(bestTake.Prompt)
		filename := fmt.Sprintf("%03d_%s.mp4", index, sanitized)

		exports = append(exports, ExportData{
			Filename: filename,
			VideoURL: bestTake.VideoURL,
			Duration: bestTake.Duration,
		})
		index++
	}

	return exports
}

// CreateExportZIP creates a ZIP file containing all videos and FCPXML
func CreateExportZIP(w io.Writer, projectName string, exports []ExportData) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Default properties
	width, height := 1280, 720
	frameDuration := "100/2400s" // 24fps default

	// Inspect the first video to get properties if available
	if len(exports) > 0 {
		tempFile, err := os.CreateTemp("", "scan_video_*.mp4")
		if err == nil {
			defer os.Remove(tempFile.Name()) // Clean up
			defer tempFile.Close()

			// Download first video header (or full file if small)
			// For simplicity and correctness, downloading full file to temp is safest to ensure we can seek if needed.
			// Since videos are short, this is acceptable overhead.
			if err := DownloadVideo(exports[0].VideoURL, tempFile.Name()); err == nil {
				// Re-open for reading
				f, err := os.Open(tempFile.Name())
				if err == nil {
					defer f.Close()
					w, h, fd, err := GetVideoProperties(f)
					if err == nil {
						width = w
						height = h
						frameDuration = fd
					}
				}
			}
		}
	}

	// Generate and add FCPXML
	fcpxmlData, err := GenerateFCPXML(projectName, exports, width, height, frameDuration)
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

// GetVideoProperties attempts to parse MP4 atoms to find resolution and frame rate
func GetVideoProperties(r io.ReadSeeker) (int, int, string, error) {
	// Simple atom walker
	buf := make([]byte, 8)
	var timescale uint32
	var width, height int

	// Reset seeker
	r.Seek(0, io.SeekStart)

	// We need to find moov -> trak -> mdia -> mdhd (for timescale)
	// and moov -> trak -> tkhd (for width/height)
	// For FPS, standard FCPXML notation is "100/{fps*100}s" or similar.
	// Actually FCPXML uses integer math. 25fps = 100/2500s. 24fps = 100/2400s.
	// timescale / sample_duration = fps.

	// Since full parsing is complex, we will search for specific atoms by walking.

	// Helper to read atoms
	readAtom := func() (string, int64, error) {
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", 0, err
		}
		size := uint32(binary.BigEndian.Uint32(buf[0:4]))
		name := string(buf[4:8])
		var size64 int64 = int64(size)
		if size == 1 {
			// Extended size
			sysBuf := make([]byte, 8)
			if _, err := io.ReadFull(r, sysBuf); err != nil {
				return "", 0, err
			}
			size64 = int64(binary.BigEndian.Uint64(sysBuf))
		}
		return name, size64, nil
	}

	// Recursive finder is hard without struct definitions.
	// We will implement a simplified scanner that looks for 'moov' and reads it into memory if small enough,
	// or seeks into it.

	// Better approach: Use a proper function that takes a limit.

	fileSize, _ := r.Seek(0, io.SeekEnd)
	r.Seek(0, io.SeekStart)

	var parseAtoms func(offset int64, limit int64) error
	parseAtoms = func(offset int64, limit int64) error {
		current := offset
		for current < offset+limit {
			if _, err := r.Seek(current, io.SeekStart); err != nil {
				return err
			}
			name, size, err := readAtom()
			if err != nil {
				return err
			}

			if size == 0 {
				// Extend to end of file
				size = fileSize - current
			}

			payloadStart := current + 8
			if size == 1 {
				payloadStart += 8
			}
			payloadSize := size - (payloadStart - current)

			switch name {
			case "moov", "trak", "mdia", "minf", "stbl":
				// Container atoms, recurse
				if err := parseAtoms(payloadStart, payloadSize); err != nil {
					return err
				}
			case "mdhd":
				// Media Header
				// Version 1 byte, Flags 3 bytes
				// Creation time 4/8, Mod time 4/8, Timescale 4, Duration 4/8
				r.Seek(payloadStart, io.SeekStart)
				header := make([]byte, 4)
				io.ReadFull(r, header)
				version := header[0]

				if version == 1 {
					r.Seek(payloadStart+4+8+8, io.SeekStart) // Skip version+flags + create + mod (8 bytes each)
				} else {
					r.Seek(payloadStart+4+4+4, io.SeekStart) // Skip version+flags + create + mod (4 bytes each)
				}

				// Read Timescale
				io.ReadFull(r, header)
				timescale = binary.BigEndian.Uint32(header)

				// Read Duration (just for reference, not strict for FPS)
				if version == 1 {
					durBuf := make([]byte, 8)
					io.ReadFull(r, durBuf)
					// duration = uint64...
				} else {
					io.ReadFull(r, header)
					// duration = binary.BigEndian.Uint32(header)
				}

			case "tkhd":
				// Track Header
				// Version 1, Flags 3
				// create/mod...
				// Width/Height are at fixed offsets
				// version 0: 84 bytes matrix structure at end-ish
				// We need to look up spec.
				// Width and height are 16.16 fixed point values at the end of tkhd.
				// For version 0: duration is at offset 20 (4 bytes), reserved 12 bytes, layer 2, alt group 2, vol 2, reserved 2, matrix 36 bytes, width 4 bytes, height 4 bytes.
				// Offset to width: 4(ver/flags) + 4(create) + 4(mod) + 4(trackid) + 4(reserved) + 4(duration) + 8(reserved) + 2(layer) + 2(alt) + 2(vol) + 2(reserved) + 36(matrix) = 76 bytes?
				// Let's rely on standard offsets relative to payload.

				r.Seek(payloadStart, io.SeekStart)
				verBuf := make([]byte, 1)
				io.ReadFull(r, verBuf)
				version := verBuf[0]

				// Skip to width/height
				// Ver 0: Width at offset 76 (from start of tkhd data, excluding headers?)
				// Let's count bytes carefully.
				// struct {
				//   ver(1), flags(3)
				//   creation(4), mod(4), track(4), reserved(4), duration(4)
				//   reserved(8), layer(2), alt(2), vol(2), reserved(2)
				//   matrix(36)
				//   width(4), height(4)
				// }
				// 1+3+4+4+4+4+4 + 8+2+2+2+2 + 36 = 76.
				// Width is at 76, Height at 80.

				// Ver 1: creation(8), mod(8), track(4), reserved(4), duration(8)
				// 1+3+8+8+4+4+8 + ... same ...
				// 1+3+8+8+4+4+8 + 8+2+2+2+2 + 36 = 88.

				var seekOffset int64 = 76
				if version == 1 {
					seekOffset = 88
				}

				r.Seek(payloadStart+seekOffset, io.SeekStart)
				dimBuf := make([]byte, 8)
				if _, err := io.ReadFull(r, dimBuf); err == nil {
					wFixed := binary.BigEndian.Uint32(dimBuf[0:4])
					hFixed := binary.BigEndian.Uint32(dimBuf[4:8])
					width = int(wFixed >> 16)
					height = int(hFixed >> 16)
				}

			case "stts":
				// Time-to-sample
				// Used to calculate frame duration
				// ver(1), flags(3), count(4)
				// entries: count, duration
				// We assume constant frame rate, so first entry is enough.

				// Only if we haven't found a valid fps yet?
				// But we need to match it with timescale found in the SAME trak > mdia.
				// This implies we should track context.
				// For simplicity, we assume the first video track's properties apply.

				r.Seek(payloadStart+4, io.SeekStart) // Skip ver/flags
				countBuf := make([]byte, 4)
				io.ReadFull(r, countBuf)
				// count := binary.BigEndian.Uint32(countBuf)

				// Read first entry
				// sample_count(4), sample_delta(4)
				entryBuf := make([]byte, 8)
				io.ReadFull(r, entryBuf)
				// sampleCount := binary.BigEndian.Uint32(entryBuf[0:4])
				sampleDelta := binary.BigEndian.Uint32(entryBuf[4:8])

				if timescale > 0 && sampleDelta > 0 {
					// Found it!
					// Normalize frame duration
					return fmt.Errorf("FOUND_FPS:%s", matchFrameDuration(timescale, sampleDelta))
				}
			}

			current += size
		}
		return nil
	}

	// Run parser
	err := parseAtoms(0, fileSize)

	// Check if specialized error returned (hacky flow control but simple)
	frameDuration := "100/2500s" // Default fallback if not found
	if err != nil && strings.HasPrefix(err.Error(), "FOUND_FPS:") {
		frameDuration = strings.TrimPrefix(err.Error(), "FOUND_FPS:")
		err = nil
	}

	if width == 0 {
		width = 1280
	}
	if height == 0 {
		height = 720
	}

	return width, height, frameDuration, nil
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

// Helper to normalize FPS to standard video rates
func matchFrameDuration(timescale, sampleDelta uint32) string {
	fps := float64(timescale) / float64(sampleDelta)

	// Standard rates tolerance
	tolerance := 0.01

	type StandardRate struct {
		FPS      float64
		Duration string
	}

	rates := []StandardRate{
		{23.976, "1001/24000s"},
		{24.0, "100/2400s"},
		{25.0, "100/2500s"},
		{29.97, "1001/30000s"},
		{30.0, "100/3000s"},
		{50.0, "100/5000s"},
		{59.94, "1001/60000s"},
		{60.0, "100/6000s"},
	}

	for _, r := range rates {
		if math.Abs(fps-r.FPS) < tolerance {
			return r.Duration
		}
	}

	// Fallback to raw exact fraction
	return fmt.Sprintf("%d/%ds", sampleDelta, timescale)
}
