package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"seedance-client/config"
	"seedance-client/models"
	"seedance-client/services"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

const (
	defaultLLMModelID   = "doubao-seed-1-6-250615"
	defaultImageModelID = "doubao-seedream-4-5-251128"
)

type EntityRef struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Prompt string `json:"prompt"`
}

type AssetVersionResponse struct {
	ID         uint      `json:"id"`
	CatalogID  uint      `json:"catalog_id"`
	VersionNo  int       `json:"version_no"`
	ImagePath  string    `json:"image_path"`
	SourceType string    `json:"source_type"`
	ModelID    string    `json:"model_id"`
	Prompt     string    `json:"prompt"`
	Status     string    `json:"status"`
	IsGood     bool      `json:"is_good"`
	CreatedAt  time.Time `json:"created_at"`
}

type AssetCatalogResponse struct {
	ID           uint                   `json:"id"`
	ProjectID    uint                   `json:"project_id"`
	AssetType    string                 `json:"asset_type"`
	AssetCode    string                 `json:"asset_code"`
	Name         string                 `json:"name"`
	Prompt       string                 `json:"prompt"`
	StoryboardID *uint                  `json:"storyboard_id,omitempty"`
	Versions     []AssetVersionResponse `json:"versions"`
	Active       *AssetVersionResponse  `json:"active"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type ShotFrameVersionResponse struct {
	ID           uint      `json:"id"`
	StoryboardID uint      `json:"storyboard_id"`
	FrameType    string    `json:"frame_type"`
	VersionNo    int       `json:"version_no"`
	ImagePath    string    `json:"image_path"`
	SourceType   string    `json:"source_type"`
	ModelID      string    `json:"model_id"`
	Prompt       string    `json:"prompt"`
	Status       string    `json:"status"`
	IsGood       bool      `json:"is_good"`
	CreatedAt    time.Time `json:"created_at"`
}

type V1ShotData struct {
	ID                uint                       `json:"id"`
	ProjectID         uint                       `json:"project_id"`
	CreatedAt         time.Time                  `json:"created_at"`
	UpdatedAt         time.Time                  `json:"updated_at"`
	ShotOrder         int                        `json:"shot_order"`
	ShotNo            string                     `json:"shot_no"`
	ShotSize          string                     `json:"shot_size"`
	CameraMovement    string                     `json:"camera_movement"`
	FrameContent      string                     `json:"frame_content"`
	Characters        []EntityRef                `json:"characters"`
	Scenes            []EntityRef                `json:"scenes"`
	Elements          []EntityRef                `json:"elements"`
	Styles            []EntityRef                `json:"styles"`
	SoundDesign       string                     `json:"sound_design"`
	EstimatedDuration int                        `json:"estimated_duration"`
	DurationFine      int                        `json:"duration_fine"`
	Takes             []TakeResponse             `json:"takes"`
	ActiveTake        *TakeResponse              `json:"active_take"`
	StartFrames       []ShotFrameVersionResponse `json:"start_frames"`
	EndFrames         []ShotFrameVersionResponse `json:"end_frames"`
	ActiveStartFrame  *ShotFrameVersionResponse  `json:"active_start_frame"`
	ActiveEndFrame    *ShotFrameVersionResponse  `json:"active_end_frame"`
}

type V1ProjectData struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	ModelVersion string    `json:"model_version"`
	AspectRatio  string    `json:"aspect_ratio"`
	CreatedAt    time.Time `json:"created_at"`
}

type V1WorkspaceData struct {
	Project              V1ProjectData          `json:"project"`
	Storyboards          []V1ShotData           `json:"storyboards"`
	AssetCatalogs        []AssetCatalogResponse `json:"asset_catalogs"`
	Models               []config.Model         `json:"models"`
	AudioSupportedModels []string               `json:"audio_supported_models"`
	LLMModelDefault      string                 `json:"llm_model_default"`
	ImageModelDefault    string                 `json:"image_model_default"`
}

type StoryboardSourceFile struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

type DecomposeStoryboardParams struct {
	ProjectID       uint   `json:"project_id"`
	SourceText      string `json:"source_text"`
	LLMModelID      string `json:"llm_model_id"`
	Provider        string `json:"provider"` // ark_default | ark_custom | openai_compatible
	APIKey          string `json:"api_key"`
	BaseURL         string `json:"base_url"`
	ReplaceExisting bool   `json:"replace_existing"`
}

type UpdateShotParams struct {
	StoryboardID      uint        `json:"storyboard_id"`
	ShotNo            string      `json:"shot_no"`
	ShotSize          string      `json:"shot_size"`
	CameraMovement    string      `json:"camera_movement"`
	FrameContent      string      `json:"frame_content"`
	Characters        []EntityRef `json:"characters"`
	Scenes            []EntityRef `json:"scenes"`
	Elements          []EntityRef `json:"elements"`
	Styles            []EntityRef `json:"styles"`
	SoundDesign       string      `json:"sound_design"`
	EstimatedDuration int         `json:"estimated_duration"`
	DurationFine      int         `json:"duration_fine"`
}

type CreateV1ShotParams struct {
	ProjectID        uint `json:"project_id"`
	AfterStoryboardID uint `json:"after_storyboard_id"`
}

type SplitShotParams struct {
	StoryboardID  uint   `json:"storyboard_id"`
	FirstContent  string `json:"first_content"`
	SecondContent string `json:"second_content"`
}

type UpdateAssetCatalogParams struct {
	CatalogID uint   `json:"catalog_id"`
	Name      string `json:"name"`
	Prompt    string `json:"prompt"`
}

type GenerateAssetImageParams struct {
	CatalogID   uint     `json:"catalog_id"`
	ModelID     string   `json:"model_id"`
	Prompt      string   `json:"prompt"`
	InputImages []string `json:"input_images"`
}

type GenerateShotFrameParams struct {
	StoryboardID uint     `json:"storyboard_id"`
	FrameType    string   `json:"frame_type"` // start/end
	ModelID      string   `json:"model_id"`
	Prompt       string   `json:"prompt"`
	InputImages  []string `json:"input_images"`
}

type UploadShotFrameParams struct {
	StoryboardID uint   `json:"storyboard_id"`
	FrameType    string `json:"frame_type"` // start/end
}

type llmDecomposeShot struct {
	ShotNo            string      `json:"shot_no"`
	ShotSize          string      `json:"shot_size"`
	CameraMovement    string      `json:"camera_movement"`
	FrameContent      string      `json:"frame_content"`
	Characters        []EntityRef `json:"characters"`
	Scenes            []EntityRef `json:"scenes"`
	SpecialElements   []EntityRef `json:"special_elements"`
	VisualStyle       []EntityRef `json:"visual_style"`
	SoundDesign       string      `json:"sound_design"`
	EstimatedDuration int         `json:"estimated_duration"`
}

type llmDecomposeResponse struct {
	Shots []llmDecomposeShot `json:"shots"`
}

// ============================================================
// Workspace Query
// ============================================================

func (a *App) GetV1Workspace(projectID uint) (*V1WorkspaceData, error) {
	var project models.Project
	if err := models.DB.First(&project, projectID).Error; err != nil {
		return nil, fmt.Errorf("project not found")
	}

	if project.AspectRatio == "" {
		project.AspectRatio = "16:9"
	}

	var storyboards []models.Storyboard
	if err := models.DB.Preload("Takes", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Where("project_id = ?", projectID).Order("shot_order asc, id asc").Find(&storyboards).Error; err != nil {
		return nil, err
	}

	storyboardIDs := make([]uint, 0, len(storyboards))
	for _, sb := range storyboards {
		storyboardIDs = append(storyboardIDs, sb.ID)
	}

	frameMap := map[uint]map[string][]models.ShotFrameVersion{}
	if len(storyboardIDs) > 0 {
		var frames []models.ShotFrameVersion
		models.DB.Where("storyboard_id IN ?", storyboardIDs).Order("frame_type asc, version_no asc, created_at asc").Find(&frames)
		for _, f := range frames {
			if _, ok := frameMap[f.StoryboardID]; !ok {
				frameMap[f.StoryboardID] = map[string][]models.ShotFrameVersion{}
			}
			frameMap[f.StoryboardID][f.FrameType] = append(frameMap[f.StoryboardID][f.FrameType], f)
		}
	}

	shotList := make([]V1ShotData, 0, len(storyboards))
	for _, sb := range storyboards {
		takes := make([]TakeResponse, 0, len(sb.Takes))
		for i := range sb.Takes {
			takes = append(takes, takeToResponse(&sb.Takes[i]))
		}

		var activeTake *TakeResponse
		if len(takes) > 0 {
			for i := len(takes) - 1; i >= 0; i-- {
				if takes[i].IsGood {
					t := takes[i]
					activeTake = &t
					break
				}
			}
			if activeTake == nil {
				t := takes[len(takes)-1]
				activeTake = &t
			}
		}

		startFrames := toFrameVersionResponses(frameMap[sb.ID]["start"])
		endFrames := toFrameVersionResponses(frameMap[sb.ID]["end"])

		shot := V1ShotData{
			ID:                sb.ID,
			ProjectID:         sb.ProjectID,
			CreatedAt:         sb.CreatedAt,
			UpdatedAt:         sb.UpdatedAt,
			ShotOrder:         sb.ShotOrder,
			ShotNo:            sb.ShotNo,
			ShotSize:          sb.ShotSize,
			CameraMovement:    sb.CameraMovement,
			FrameContent:      sb.FrameContent,
			Characters:        parseEntityRefs(sb.CharactersJSON),
			Scenes:            parseEntityRefs(sb.ScenesJSON),
			Elements:          parseEntityRefs(sb.ElementsJSON),
			Styles:            parseEntityRefs(sb.StylesJSON),
			SoundDesign:       sb.SoundDesign,
			EstimatedDuration: normalizeDuration(sb.EstimatedDuration),
			DurationFine:      sb.DurationFine,
			Takes:             takes,
			ActiveTake:        activeTake,
			StartFrames:       startFrames,
			EndFrames:         endFrames,
			ActiveStartFrame:  chooseActiveFrame(startFrames),
			ActiveEndFrame:    chooseActiveFrame(endFrames),
		}
		shotList = append(shotList, shot)
	}

	var catalogs []models.AssetCatalog
	if err := models.DB.Preload("Versions", func(db *gorm.DB) *gorm.DB {
		return db.Order("version_no asc, created_at asc")
	}).Where("project_id = ?", projectID).Order("asset_type asc, asset_code asc, id asc").Find(&catalogs).Error; err != nil {
		return nil, err
	}

	assetCatalogs := make([]AssetCatalogResponse, 0, len(catalogs))
	for _, c := range catalogs {
		versions := make([]AssetVersionResponse, 0, len(c.Versions))
		for _, v := range c.Versions {
			versions = append(versions, toAssetVersionResponse(v))
		}

		catalogResp := AssetCatalogResponse{
			ID:           c.ID,
			ProjectID:    c.ProjectID,
			AssetType:    c.AssetType,
			AssetCode:    c.AssetCode,
			Name:         c.Name,
			Prompt:       c.Prompt,
			StoryboardID: c.StoryboardID,
			Versions:     versions,
			Active:       chooseActiveAssetVersion(versions),
			UpdatedAt:    c.UpdatedAt,
		}
		assetCatalogs = append(assetCatalogs, catalogResp)
	}

	return &V1WorkspaceData{
		Project: V1ProjectData{
			ID:           project.ID,
			Name:         project.Name,
			ModelVersion: project.ModelVersion,
			AspectRatio:  project.AspectRatio,
			CreatedAt:    project.CreatedAt,
		},
		Storyboards:          shotList,
		AssetCatalogs:        assetCatalogs,
		Models:               config.GetModels(),
		AudioSupportedModels: config.GetAudioSupportedModelIDs(),
		LLMModelDefault:      getDefaultLLMModel(),
		ImageModelDefault:    defaultImageModelID,
	}, nil
}

// ============================================================
// Source File Import
// ============================================================

func (a *App) SelectStoryboardSourceFile() (*StoryboardSourceFile, error) {
	result, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Storyboard Source File",
		Filters: []wailsRuntime.FileFilter{
			{DisplayName: "Storyboard Source (*.md;*.markdown;*.txt;*.csv;*.tsv;*.xlsx)", Pattern: "*.md;*.markdown;*.txt;*.csv;*.tsv;*.xlsx"},
		},
	})
	if err != nil {
		return nil, err
	}
	if result == "" {
		return &StoryboardSourceFile{}, nil
	}

	ext := strings.ToLower(filepath.Ext(result))
	var content string
	switch ext {
	case ".md", ".markdown", ".txt":
		b, err := os.ReadFile(result)
		if err != nil {
			return nil, err
		}
		content = string(b)
	case ".csv":
		b, err := os.ReadFile(result)
		if err != nil {
			return nil, err
		}
		content, err = csvBytesToMarkdown(b, ',')
		if err != nil {
			return nil, err
		}
	case ".tsv":
		b, err := os.ReadFile(result)
		if err != nil {
			return nil, err
		}
		content, err = csvBytesToMarkdown(b, '\t')
		if err != nil {
			return nil, err
		}
	case ".xlsx":
		md, err := xlsxToMarkdown(result)
		if err != nil {
			return nil, err
		}
		content = md
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	return &StoryboardSourceFile{
		Filename: filepath.Base(result),
		Content:  content,
	}, nil
}

// ============================================================
// LLM Decomposition
// ============================================================

func (a *App) DecomposeStoryboardWithLLM(params DecomposeStoryboardParams) (*V1WorkspaceData, error) {
	if params.ProjectID == 0 {
		return nil, fmt.Errorf("project_id is required")
	}
	if strings.TrimSpace(params.SourceText) == "" {
		return nil, fmt.Errorf("source_text is required")
	}
	if params.LLMModelID == "" {
		params.LLMModelID = getDefaultLLMModel()
	}
	if params.Provider == "" {
		params.Provider = "ark_default"
	}

	var project models.Project
	if err := models.DB.First(&project, params.ProjectID).Error; err != nil {
		return nil, fmt.Errorf("project not found")
	}
	if project.AspectRatio == "" {
		project.AspectRatio = "16:9"
	}

	decoded, err := a.callStoryboardDecomposeLLM(params.Provider, params.APIKey, params.BaseURL, params.LLMModelID, params.SourceText)
	if err != nil {
		return nil, err
	}
	if len(decoded.Shots) == 0 {
		return nil, fmt.Errorf("no shots returned by LLM")
	}

	tx := models.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if params.ReplaceExisting {
		var oldSBs []models.Storyboard
		tx.Where("project_id = ?", params.ProjectID).Find(&oldSBs)
		if len(oldSBs) > 0 {
			ids := make([]uint, 0, len(oldSBs))
			for _, sb := range oldSBs {
				ids = append(ids, sb.ID)
			}
			tx.Where("storyboard_id IN ?", ids).Delete(&models.ShotFrameVersion{})
		}
		tx.Where("project_id = ?", params.ProjectID).Delete(&models.Storyboard{})
		tx.Where("project_id = ?", params.ProjectID).Delete(&models.AssetCatalog{})
	}

	orderBase := 1
	if !params.ReplaceExisting {
		var maxOrder int
		tx.Model(&models.Storyboard{}).Where("project_id = ?", params.ProjectID).Select("COALESCE(MAX(shot_order),0)").Scan(&maxOrder)
		orderBase = maxOrder + 1
	}

	defaultModel := config.GetDefaultModel()
	defaultModelID := ""
	if defaultModel != nil {
		defaultModelID = defaultModel.ID
	}
	if defaultModelID == "" {
		defaultModelID = "doubao-seedance-1-5-pro-251215"
	}

	for i, shot := range decoded.Shots {
		charRefs := normalizeRefs("character", shot.Characters, i+1)
		sceneRefs := normalizeRefs("scene", shot.Scenes, i+1)
		elementRefs := normalizeRefs("element", shot.SpecialElements, i+1)
		styleRefs := normalizeRefs("style", shot.VisualStyle, i+1)

		charJSON, _ := refsToJSON(charRefs)
		sceneJSON, _ := refsToJSON(sceneRefs)
		elementJSON, _ := refsToJSON(elementRefs)
		styleJSON, _ := refsToJSON(styleRefs)

		sb := models.Storyboard{
			ProjectID:         params.ProjectID,
			ShotOrder:         orderBase + i,
			ShotNo:            strings.TrimSpace(shot.ShotNo),
			ShotSize:          strings.TrimSpace(shot.ShotSize),
			CameraMovement:    strings.TrimSpace(shot.CameraMovement),
			FrameContent:      strings.TrimSpace(shot.FrameContent),
			CharactersJSON:    charJSON,
			ScenesJSON:        sceneJSON,
			ElementsJSON:      elementJSON,
			StylesJSON:        styleJSON,
			SoundDesign:       strings.TrimSpace(shot.SoundDesign),
			EstimatedDuration: normalizeDuration(shot.EstimatedDuration),
			DurationFine:      0,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		if err := tx.Create(&sb).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		if err := syncCatalogRefsTx(tx, params.ProjectID, sb.ID, "character", charRefs); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := syncCatalogRefsTx(tx, params.ProjectID, sb.ID, "scene", sceneRefs); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := syncCatalogRefsTx(tx, params.ProjectID, sb.ID, "element", elementRefs); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := syncCatalogRefsTx(tx, params.ProjectID, sb.ID, "style", styleRefs); err != nil {
			tx.Rollback()
			return nil, err
		}

		initialPrompt := composeShotPrompt(sb.FrameContent, charRefs, sceneRefs, elementRefs, styleRefs, sb.SoundDesign)
		take := models.Take{
			StoryboardID:   sb.ID,
			Prompt:         initialPrompt,
			ModelID:        defaultModelID,
			Ratio:          project.AspectRatio,
			Duration:       sb.EstimatedDuration,
			GenerateAudio:  false,
			ServiceTier:    "standard",
			GenerationMode: "standard",
			Status:         "Draft",
			CreatedAt:      time.Now(),
		}
		if err := tx.Create(&take).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return a.GetV1Workspace(params.ProjectID)
}

// ============================================================
// Shot Operations
// ============================================================

func (a *App) CreateV1Shot(params CreateV1ShotParams) (uint, error) {
	if params.ProjectID == 0 {
		return 0, fmt.Errorf("project_id is required")
	}

	var project models.Project
	if err := models.DB.First(&project, params.ProjectID).Error; err != nil {
		return 0, fmt.Errorf("project not found")
	}

	ratio := strings.TrimSpace(project.AspectRatio)
	if ratio == "" {
		ratio = "16:9"
	}

	defaultModel := config.GetDefaultModel()
	defaultModelID := ""
	if defaultModel != nil {
		defaultModelID = strings.TrimSpace(defaultModel.ID)
	}
	if defaultModelID == "" {
		defaultModelID = "doubao-seedance-1-5-pro-251215"
	}

	tx := models.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	insertOrder := 1
	if params.AfterStoryboardID > 0 {
		var after models.Storyboard
		if err := tx.Where("id = ? AND project_id = ?", params.AfterStoryboardID, params.ProjectID).First(&after).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return 0, fmt.Errorf("reference shot not found")
			}
			return 0, err
		}

		insertOrder = after.ShotOrder + 1
		if err := tx.Model(&models.Storyboard{}).
			Where("project_id = ? AND shot_order >= ?", params.ProjectID, insertOrder).
			Update("shot_order", gorm.Expr("shot_order + 1")).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
	} else {
		var maxOrder int
		if err := tx.Model(&models.Storyboard{}).
			Where("project_id = ?", params.ProjectID).
			Select("COALESCE(MAX(shot_order),0)").
			Scan(&maxOrder).Error; err != nil {
			tx.Rollback()
			return 0, err
		}
		insertOrder = maxOrder + 1
	}

	now := time.Now()
	sb := models.Storyboard{
		ProjectID:         params.ProjectID,
		ShotOrder:         insertOrder,
		EstimatedDuration: 5,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	if err := tx.Create(&sb).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	initialPrompt := composeShotPrompt(sb.FrameContent, nil, nil, nil, nil, sb.SoundDesign)
	take := models.Take{
		StoryboardID:   sb.ID,
		Prompt:         initialPrompt,
		ModelID:        defaultModelID,
		Ratio:          ratio,
		Duration:       normalizeDuration(sb.EstimatedDuration),
		GenerateAudio:  false,
		ServiceTier:    "standard",
		GenerationMode: "standard",
		Status:         "Draft",
		CreatedAt:      now,
	}
	if err := tx.Create(&take).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := resequenceStoryboardsTx(tx, params.ProjectID); err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return sb.ID, nil
}

func (a *App) UpdateShotMetadata(params UpdateShotParams) error {
	var sb models.Storyboard
	if err := models.DB.First(&sb, params.StoryboardID).Error; err != nil {
		return fmt.Errorf("shot not found")
	}

	charRefs := normalizeRefs("character", params.Characters, sb.ShotOrder)
	sceneRefs := normalizeRefs("scene", params.Scenes, sb.ShotOrder)
	elementRefs := normalizeRefs("element", params.Elements, sb.ShotOrder)
	styleRefs := normalizeRefs("style", params.Styles, sb.ShotOrder)

	charJSON, _ := refsToJSON(charRefs)
	sceneJSON, _ := refsToJSON(sceneRefs)
	elementJSON, _ := refsToJSON(elementRefs)
	styleJSON, _ := refsToJSON(styleRefs)

	sb.ShotNo = strings.TrimSpace(params.ShotNo)
	sb.ShotSize = strings.TrimSpace(params.ShotSize)
	sb.CameraMovement = strings.TrimSpace(params.CameraMovement)
	sb.FrameContent = strings.TrimSpace(params.FrameContent)
	sb.CharactersJSON = charJSON
	sb.ScenesJSON = sceneJSON
	sb.ElementsJSON = elementJSON
	sb.StylesJSON = styleJSON
	sb.SoundDesign = strings.TrimSpace(params.SoundDesign)
	sb.EstimatedDuration = normalizeDuration(params.EstimatedDuration)
	sb.DurationFine = params.DurationFine
	sb.UpdatedAt = time.Now()

	tx := models.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Save(&sb).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := syncCatalogRefsTx(tx, sb.ProjectID, sb.ID, "character", charRefs); err != nil {
		tx.Rollback()
		return err
	}
	if err := syncCatalogRefsTx(tx, sb.ProjectID, sb.ID, "scene", sceneRefs); err != nil {
		tx.Rollback()
		return err
	}
	if err := syncCatalogRefsTx(tx, sb.ProjectID, sb.ID, "element", elementRefs); err != nil {
		tx.Rollback()
		return err
	}
	if err := syncCatalogRefsTx(tx, sb.ProjectID, sb.ID, "style", styleRefs); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (a *App) DeleteV1Shot(storyboardID uint) error {
	var sb models.Storyboard
	if err := models.DB.First(&sb, storyboardID).Error; err != nil {
		return fmt.Errorf("shot not found")
	}

	tx := models.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Where("storyboard_id = ?", storyboardID).Delete(&models.ShotFrameVersion{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.Storyboard{}, storyboardID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := resequenceStoryboardsTx(tx, sb.ProjectID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (a *App) MergeShotWithNext(storyboardID uint) error {
	var current models.Storyboard
	if err := models.DB.First(&current, storyboardID).Error; err != nil {
		return fmt.Errorf("shot not found")
	}

	var next models.Storyboard
	if err := models.DB.Where("project_id = ? AND shot_order > ?", current.ProjectID, current.ShotOrder).Order("shot_order asc, id asc").First(&next).Error; err != nil {
		return fmt.Errorf("no next shot to merge")
	}

	mergedChars := mergeRefs(parseEntityRefs(current.CharactersJSON), parseEntityRefs(next.CharactersJSON))
	mergedScenes := mergeRefs(parseEntityRefs(current.ScenesJSON), parseEntityRefs(next.ScenesJSON))
	mergedElements := mergeRefs(parseEntityRefs(current.ElementsJSON), parseEntityRefs(next.ElementsJSON))
	mergedStyles := mergeRefs(parseEntityRefs(current.StylesJSON), parseEntityRefs(next.StylesJSON))

	charJSON, _ := refsToJSON(mergedChars)
	sceneJSON, _ := refsToJSON(mergedScenes)
	elementJSON, _ := refsToJSON(mergedElements)
	styleJSON, _ := refsToJSON(mergedStyles)

	current.FrameContent = strings.TrimSpace(strings.TrimSpace(current.FrameContent) + "\n" + strings.TrimSpace(next.FrameContent))
	current.SoundDesign = strings.TrimSpace(strings.TrimSpace(current.SoundDesign) + "\n" + strings.TrimSpace(next.SoundDesign))
	current.CharactersJSON = charJSON
	current.ScenesJSON = sceneJSON
	current.ElementsJSON = elementJSON
	current.StylesJSON = styleJSON
	if current.ShotNo != "" && next.ShotNo != "" {
		current.ShotNo = current.ShotNo + "+" + next.ShotNo
	}
	totalDuration := current.EstimatedDuration + next.EstimatedDuration
	if totalDuration <= 5 {
		current.EstimatedDuration = 5
	} else {
		current.EstimatedDuration = 10
	}
	current.UpdatedAt = time.Now()

	tx := models.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := tx.Save(&current).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("storyboard_id = ?", next.ID).Delete(&models.ShotFrameVersion{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&models.Storyboard{}, next.ID).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := resequenceStoryboardsTx(tx, current.ProjectID); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (a *App) SplitShot(params SplitShotParams) (uint, error) {
	var sb models.Storyboard
	if err := models.DB.Preload("Takes", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).First(&sb, params.StoryboardID).Error; err != nil {
		return 0, fmt.Errorf("shot not found")
	}

	firstContent := strings.TrimSpace(params.FirstContent)
	secondContent := strings.TrimSpace(params.SecondContent)
	if firstContent == "" && secondContent == "" {
		return 0, fmt.Errorf("split content cannot both be empty")
	}
	if firstContent == "" {
		firstContent = autoSplitFirstPart(sb.FrameContent)
	}
	if secondContent == "" {
		secondContent = autoSplitSecondPart(sb.FrameContent)
	}
	if strings.TrimSpace(secondContent) == "" {
		return 0, fmt.Errorf("second shot content is empty")
	}

	tx := models.DB.Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}

	if err := tx.Model(&models.Storyboard{}).
		Where("project_id = ? AND shot_order > ?", sb.ProjectID, sb.ShotOrder).
		Update("shot_order", gorm.Expr("shot_order + 1")).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	sb.FrameContent = firstContent
	sb.UpdatedAt = time.Now()
	if err := tx.Save(&sb).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	newSB := models.Storyboard{
		ProjectID:         sb.ProjectID,
		ShotOrder:         sb.ShotOrder + 1,
		ShotNo:            sb.ShotNo + ".B",
		ShotSize:          sb.ShotSize,
		CameraMovement:    sb.CameraMovement,
		FrameContent:      secondContent,
		CharactersJSON:    sb.CharactersJSON,
		ScenesJSON:        sb.ScenesJSON,
		ElementsJSON:      sb.ElementsJSON,
		StylesJSON:        sb.StylesJSON,
		SoundDesign:       sb.SoundDesign,
		EstimatedDuration: sb.EstimatedDuration,
		DurationFine:      sb.DurationFine,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := tx.Create(&newSB).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	var project models.Project
	tx.First(&project, sb.ProjectID)
	ratio := project.AspectRatio
	if ratio == "" {
		ratio = "16:9"
	}

	baseTake := models.Take{
		ModelID:        "doubao-seedance-1-5-pro-251215",
		Ratio:          ratio,
		Duration:       normalizeDuration(sb.EstimatedDuration),
		ServiceTier:    "standard",
		GenerationMode: "standard",
		Status:         "Draft",
	}
	if len(sb.Takes) > 0 {
		baseTake = sb.Takes[len(sb.Takes)-1]
	}
	baseTake.ID = 0
	baseTake.StoryboardID = newSB.ID
	baseTake.TaskID = ""
	baseTake.Status = "Draft"
	baseTake.VideoURL = ""
	baseTake.LastFrameURL = ""
	baseTake.LocalVideoPath = ""
	baseTake.LocalLastFramePath = ""
	baseTake.DownloadStatus = ""
	baseTake.TokenUsage = 0
	baseTake.IsGood = false
	baseTake.Prompt = composeShotPrompt(
		newSB.FrameContent,
		parseEntityRefs(newSB.CharactersJSON),
		parseEntityRefs(newSB.ScenesJSON),
		parseEntityRefs(newSB.ElementsJSON),
		parseEntityRefs(newSB.StylesJSON),
		newSB.SoundDesign,
	)
	baseTake.CreatedAt = time.Now()
	if err := tx.Create(&baseTake).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := resequenceStoryboardsTx(tx, sb.ProjectID); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	return newSB.ID, nil
}

// ============================================================
// Asset Catalog
// ============================================================

func (a *App) UpdateAssetCatalog(params UpdateAssetCatalogParams) error {
	var catalog models.AssetCatalog
	if err := models.DB.First(&catalog, params.CatalogID).Error; err != nil {
		return fmt.Errorf("asset catalog not found")
	}
	catalog.Name = strings.TrimSpace(params.Name)
	catalog.Prompt = strings.TrimSpace(params.Prompt)
	return models.DB.Save(&catalog).Error
}

func (a *App) GenerateAssetImage(params GenerateAssetImageParams) (*AssetVersionResponse, error) {
	var catalog models.AssetCatalog
	if err := models.DB.First(&catalog, params.CatalogID).Error; err != nil {
		return nil, fmt.Errorf("asset catalog not found")
	}

	prompt := strings.TrimSpace(params.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(catalog.Prompt)
	}
	if prompt == "" {
		return nil, fmt.Errorf("asset prompt is empty")
	}
	modelID := strings.TrimSpace(params.ModelID)
	if modelID == "" {
		modelID = defaultImageModelID
	}

	imagePath, err := a.generateImageToLocal(modelID, prompt, params.InputImages)
	if err != nil {
		return nil, err
	}

	versionNo := nextAssetVersionNo(catalog.ID)
	version := models.AssetVersion{
		CatalogID:  catalog.ID,
		VersionNo:  versionNo,
		ImagePath:  imagePath,
		SourceType: "generated",
		ModelID:    modelID,
		Prompt:     prompt,
		Status:     "Succeeded",
		CreatedAt:  time.Now(),
	}
	if err := models.DB.Create(&version).Error; err != nil {
		return nil, err
	}
	resp := toAssetVersionResponse(version)
	return &resp, nil
}

func (a *App) UploadAssetImage(catalogID uint) (*AssetVersionResponse, error) {
	var catalog models.AssetCatalog
	if err := models.DB.First(&catalog, catalogID).Error; err != nil {
		return nil, fmt.Errorf("asset catalog not found")
	}

	path, err := a.SelectImageFile()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}

	versionNo := nextAssetVersionNo(catalog.ID)
	version := models.AssetVersion{
		CatalogID:  catalog.ID,
		VersionNo:  versionNo,
		ImagePath:  path,
		SourceType: "uploaded",
		ModelID:    "",
		Prompt:     catalog.Prompt,
		Status:     "Succeeded",
		CreatedAt:  time.Now(),
	}
	if err := models.DB.Create(&version).Error; err != nil {
		return nil, err
	}
	resp := toAssetVersionResponse(version)
	return &resp, nil
}

func (a *App) ToggleAssetVersionGood(versionID uint) (bool, error) {
	var version models.AssetVersion
	if err := models.DB.First(&version, versionID).Error; err != nil {
		return false, fmt.Errorf("asset version not found")
	}

	newState := !version.IsGood
	tx := models.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	if newState {
		if err := tx.Model(&models.AssetVersion{}).Where("catalog_id = ? AND id != ?", version.CatalogID, version.ID).Update("is_good", false).Error; err != nil {
			tx.Rollback()
			return false, err
		}
	}
	if err := tx.Model(&version).Update("is_good", newState).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	if err := tx.Commit().Error; err != nil {
		return false, err
	}
	return newState, nil
}

// ============================================================
// Shot Frames
// ============================================================

func (a *App) GenerateShotFrame(params GenerateShotFrameParams) (*ShotFrameVersionResponse, error) {
	frameType := normalizeFrameType(params.FrameType)
	if frameType == "" {
		return nil, fmt.Errorf("frame_type must be start or end")
	}

	var sb models.Storyboard
	if err := models.DB.First(&sb, params.StoryboardID).Error; err != nil {
		return nil, fmt.Errorf("shot not found")
	}

	prompt := strings.TrimSpace(params.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(sb.FrameContent)
	}
	if prompt == "" {
		return nil, fmt.Errorf("frame prompt is empty")
	}
	modelID := strings.TrimSpace(params.ModelID)
	if modelID == "" {
		modelID = defaultImageModelID
	}

	imagePath, err := a.generateImageToLocal(modelID, prompt, params.InputImages)
	if err != nil {
		return nil, err
	}

	versionNo := nextShotFrameVersionNo(sb.ID, frameType)
	frame := models.ShotFrameVersion{
		StoryboardID: sb.ID,
		FrameType:    frameType,
		VersionNo:    versionNo,
		ImagePath:    imagePath,
		SourceType:   "generated",
		ModelID:      modelID,
		Prompt:       prompt,
		Status:       "Succeeded",
		CreatedAt:    time.Now(),
	}
	if err := models.DB.Create(&frame).Error; err != nil {
		return nil, err
	}
	resp := toShotFrameVersionResponse(frame)
	return &resp, nil
}

func (a *App) UploadShotFrame(params UploadShotFrameParams) (*ShotFrameVersionResponse, error) {
	frameType := normalizeFrameType(params.FrameType)
	if frameType == "" {
		return nil, fmt.Errorf("frame_type must be start or end")
	}
	var sb models.Storyboard
	if err := models.DB.First(&sb, params.StoryboardID).Error; err != nil {
		return nil, fmt.Errorf("shot not found")
	}

	path, err := a.SelectImageFile()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, nil
	}

	versionNo := nextShotFrameVersionNo(sb.ID, frameType)
	frame := models.ShotFrameVersion{
		StoryboardID: sb.ID,
		FrameType:    frameType,
		VersionNo:    versionNo,
		ImagePath:    path,
		SourceType:   "uploaded",
		ModelID:      "",
		Prompt:       sb.FrameContent,
		Status:       "Succeeded",
		CreatedAt:    time.Now(),
	}
	if err := models.DB.Create(&frame).Error; err != nil {
		return nil, err
	}
	resp := toShotFrameVersionResponse(frame)
	return &resp, nil
}

func (a *App) ToggleShotFrameGood(versionID uint) (bool, error) {
	var frame models.ShotFrameVersion
	if err := models.DB.First(&frame, versionID).Error; err != nil {
		return false, fmt.Errorf("frame version not found")
	}

	newState := !frame.IsGood
	tx := models.DB.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}
	if newState {
		if err := tx.Model(&models.ShotFrameVersion{}).
			Where("storyboard_id = ? AND frame_type = ? AND id != ?", frame.StoryboardID, frame.FrameType, frame.ID).
			Update("is_good", false).Error; err != nil {
			tx.Rollback()
			return false, err
		}
	}
	if err := tx.Model(&frame).Update("is_good", newState).Error; err != nil {
		tx.Rollback()
		return false, err
	}
	if err := tx.Commit().Error; err != nil {
		return false, err
	}
	return newState, nil
}

// ============================================================
// Generation-time Helpers (used by app.go)
// ============================================================

func getActiveShotFramePath(storyboardID uint, frameType string) string {
	frameType = normalizeFrameType(frameType)
	if frameType == "" {
		return ""
	}
	var frames []models.ShotFrameVersion
	models.DB.Where("storyboard_id = ? AND frame_type = ?", storyboardID, frameType).Order("created_at asc").Find(&frames)
	if len(frames) == 0 {
		return ""
	}
	var active *models.ShotFrameVersion
	for i := len(frames) - 1; i >= 0; i-- {
		if frames[i].IsGood {
			active = &frames[i]
			break
		}
	}
	if active == nil {
		active = &frames[len(frames)-1]
	}
	return active.ImagePath
}

func composeTakePromptWithAssetRefs(sb models.Storyboard, basePrompt string) string {
	base := strings.TrimSpace(basePrompt)
	projectID := sb.ProjectID

	charRefs := parseEntityRefs(sb.CharactersJSON)
	sceneRefs := parseEntityRefs(sb.ScenesJSON)
	elementRefs := parseEntityRefs(sb.ElementsJSON)
	styleRefs := parseEntityRefs(sb.StylesJSON)

	catalogs := loadCatalogMap(projectID)
	var sections []string
	appendSection := func(title string, refs []EntityRef, assetType string) {
		lines := make([]string, 0, len(refs))
		for _, ref := range refs {
			key := assetType + ":" + ref.ID
			catalog, ok := catalogs[key]
			name := ref.Name
			prompt := ref.Prompt
			if ok {
				if catalog.Name != "" {
					name = catalog.Name
				}
				if catalog.Prompt != "" {
					prompt = catalog.Prompt
				}
				if active := chooseActiveCatalogVersionModel(catalog.Versions); active != nil && active.ImagePath != "" {
					prompt = strings.TrimSpace(prompt + "（参考图: /" + active.ImagePath + "）")
				}
			}
			if name == "" && prompt == "" {
				continue
			}
			if prompt == "" {
				lines = append(lines, "- "+name)
			} else {
				lines = append(lines, fmt.Sprintf("- %s: %s", name, prompt))
			}
		}
		if len(lines) > 0 {
			sections = append(sections, title+"\n"+strings.Join(lines, "\n"))
		}
	}

	appendSection("角色参考", charRefs, "character")
	appendSection("场景参考", sceneRefs, "scene")
	appendSection("特殊元素参考", elementRefs, "element")
	appendSection("风格参考", styleRefs, "style")

	if len(sections) == 0 {
		return base
	}
	return strings.TrimSpace(base + "\n\n" + strings.Join(sections, "\n\n"))
}

func getChainedFirstFramePath(storyboardID uint) string {
	var current models.Storyboard
	if err := models.DB.First(&current, storyboardID).Error; err != nil {
		return ""
	}
	var prev models.Storyboard
	if err := models.DB.Where("project_id = ? AND shot_order < ?", current.ProjectID, current.ShotOrder).Order("shot_order desc, id desc").First(&prev).Error; err != nil {
		return ""
	}
	if path := getActiveShotFramePath(prev.ID, "end"); path != "" {
		return path
	}

	var takes []models.Take
	models.DB.Where("storyboard_id = ?", prev.ID).Order("created_at asc").Find(&takes)
	if len(takes) == 0 {
		return ""
	}
	var active *models.Take
	for i := len(takes) - 1; i >= 0; i-- {
		if takes[i].IsGood {
			active = &takes[i]
			break
		}
	}
	if active == nil {
		active = &takes[len(takes)-1]
	}
	if active.LocalLastFramePath != "" {
		return active.LocalLastFramePath
	}
	return active.LastFramePath
}

// ============================================================
// Internal Helpers
// ============================================================

func getDefaultLLMModel() string {
	if v := strings.TrimSpace(os.Getenv("ARK_LLM_MODEL")); v != "" {
		return v
	}
	return defaultLLMModelID
}

func normalizeDuration(v int) int {
	if v >= 10 {
		return 10
	}
	return 5
}

func normalizeFrameType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "start", "first", "first_frame":
		return "start"
	case "end", "last", "last_frame":
		return "end"
	default:
		return ""
	}
}

func parseEntityRefs(raw string) []EntityRef {
	if strings.TrimSpace(raw) == "" {
		return []EntityRef{}
	}
	var refs []EntityRef
	if err := json.Unmarshal([]byte(raw), &refs); err != nil {
		return []EntityRef{}
	}
	for i := range refs {
		refs[i].ID = strings.TrimSpace(refs[i].ID)
		refs[i].Name = strings.TrimSpace(refs[i].Name)
		refs[i].Prompt = strings.TrimSpace(refs[i].Prompt)
	}
	return refs
}

func refsToJSON(refs []EntityRef) (string, error) {
	b, err := json.Marshal(refs)
	if err != nil {
		return "[]", err
	}
	return string(b), nil
}

func normalizeRefs(assetType string, refs []EntityRef, shotIndex int) []EntityRef {
	out := make([]EntityRef, 0, len(refs))
	for i, ref := range refs {
		item := EntityRef{
			ID:     strings.TrimSpace(ref.ID),
			Name:   strings.TrimSpace(ref.Name),
			Prompt: strings.TrimSpace(ref.Prompt),
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("%s_%02d_%02d", assetType, shotIndex, i+1)
		}
		if item.Name == "" {
			item.Name = item.ID
		}
		out = append(out, item)
	}
	return out
}

func mergeRefs(a, b []EntityRef) []EntityRef {
	seen := map[string]bool{}
	out := make([]EntityRef, 0, len(a)+len(b))
	add := func(r EntityRef) {
		key := strings.TrimSpace(r.ID)
		if key == "" {
			key = strings.TrimSpace(r.Name + "|" + r.Prompt)
		}
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, r)
	}
	for _, r := range a {
		add(r)
	}
	for _, r := range b {
		add(r)
	}
	return out
}

func composeShotPrompt(frameContent string, chars, scenes, elements, styles []EntityRef, sound string) string {
	lines := []string{
		"画面内容: " + strings.TrimSpace(frameContent),
	}
	appendRefs := func(name string, refs []EntityRef) {
		if len(refs) == 0 {
			return
		}
		parts := make([]string, 0, len(refs))
		for _, r := range refs {
			text := strings.TrimSpace(r.Name)
			if text == "" {
				text = r.ID
			}
			if strings.TrimSpace(r.Prompt) != "" {
				text = text + "（" + strings.TrimSpace(r.Prompt) + "）"
			}
			parts = append(parts, text)
		}
		lines = append(lines, name+": "+strings.Join(parts, "，"))
	}
	appendRefs("角色", chars)
	appendRefs("场景", scenes)
	appendRefs("元素", elements)
	appendRefs("风格", styles)
	if strings.TrimSpace(sound) != "" {
		lines = append(lines, "声音设计: "+strings.TrimSpace(sound))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func syncCatalogRefsTx(tx *gorm.DB, projectID uint, storyboardID uint, assetType string, refs []EntityRef) error {
	for _, ref := range refs {
		code := strings.TrimSpace(ref.ID)
		if code == "" {
			continue
		}
		name := strings.TrimSpace(ref.Name)
		if name == "" {
			name = code
		}
		prompt := strings.TrimSpace(ref.Prompt)
		var existing models.AssetCatalog
		err := tx.Where("project_id = ? AND asset_type = ? AND asset_code = ?", projectID, assetType, code).First(&existing).Error
		if err == nil {
			updates := map[string]interface{}{}
			if existing.Name == "" {
				updates["name"] = name
			}
			if existing.Prompt == "" {
				updates["prompt"] = prompt
			}
			if len(updates) > 0 {
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
			}
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		sourceID := storyboardID
		catalog := models.AssetCatalog{
			ProjectID:    projectID,
			AssetType:    assetType,
			AssetCode:    code,
			Name:         name,
			Prompt:       prompt,
			StoryboardID: &sourceID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := tx.Create(&catalog).Error; err != nil {
			return err
		}
	}
	return nil
}

func resequenceStoryboardsTx(tx *gorm.DB, projectID uint) error {
	var storyboards []models.Storyboard
	if err := tx.Where("project_id = ?", projectID).Order("shot_order asc, id asc").Find(&storyboards).Error; err != nil {
		return err
	}
	for i, sb := range storyboards {
		newOrder := i + 1
		if sb.ShotOrder == newOrder {
			continue
		}
		if err := tx.Model(&models.Storyboard{}).Where("id = ?", sb.ID).Update("shot_order", newOrder).Error; err != nil {
			return err
		}
	}
	return nil
}

func chooseActiveAssetVersion(versions []AssetVersionResponse) *AssetVersionResponse {
	if len(versions) == 0 {
		return nil
	}
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].IsGood {
			v := versions[i]
			return &v
		}
	}
	v := versions[len(versions)-1]
	return &v
}

func chooseActiveFrame(versions []ShotFrameVersionResponse) *ShotFrameVersionResponse {
	if len(versions) == 0 {
		return nil
	}
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].IsGood {
			v := versions[i]
			return &v
		}
	}
	v := versions[len(versions)-1]
	return &v
}

func toAssetVersionResponse(v models.AssetVersion) AssetVersionResponse {
	return AssetVersionResponse{
		ID:         v.ID,
		CatalogID:  v.CatalogID,
		VersionNo:  v.VersionNo,
		ImagePath:  v.ImagePath,
		SourceType: v.SourceType,
		ModelID:    v.ModelID,
		Prompt:     v.Prompt,
		Status:     v.Status,
		IsGood:     v.IsGood,
		CreatedAt:  v.CreatedAt,
	}
}

func toShotFrameVersionResponse(v models.ShotFrameVersion) ShotFrameVersionResponse {
	return ShotFrameVersionResponse{
		ID:           v.ID,
		StoryboardID: v.StoryboardID,
		FrameType:    v.FrameType,
		VersionNo:    v.VersionNo,
		ImagePath:    v.ImagePath,
		SourceType:   v.SourceType,
		ModelID:      v.ModelID,
		Prompt:       v.Prompt,
		Status:       v.Status,
		IsGood:       v.IsGood,
		CreatedAt:    v.CreatedAt,
	}
}

func toFrameVersionResponses(frames []models.ShotFrameVersion) []ShotFrameVersionResponse {
	out := make([]ShotFrameVersionResponse, 0, len(frames))
	for _, f := range frames {
		out = append(out, toShotFrameVersionResponse(f))
	}
	return out
}

func nextAssetVersionNo(catalogID uint) int {
	var maxNo int
	models.DB.Model(&models.AssetVersion{}).Where("catalog_id = ?", catalogID).Select("COALESCE(MAX(version_no),0)").Scan(&maxNo)
	return maxNo + 1
}

func nextShotFrameVersionNo(storyboardID uint, frameType string) int {
	var maxNo int
	models.DB.Model(&models.ShotFrameVersion{}).Where("storyboard_id = ? AND frame_type = ?", storyboardID, frameType).Select("COALESCE(MAX(version_no),0)").Scan(&maxNo)
	return maxNo + 1
}

func (a *App) generateImageToLocal(modelID string, prompt string, inputImages []string) (string, error) {
	var imageField interface{}
	trimmed := make([]string, 0, len(inputImages))
	for _, img := range inputImages {
		v := strings.TrimSpace(img)
		if v != "" {
			trimmed = append(trimmed, v)
		}
	}
	if len(trimmed) == 1 {
		imageField = trimmed[0]
	} else if len(trimmed) > 1 {
		imageField = trimmed
	}

	req := model.GenerateImagesRequest{
		Model:          modelID,
		Prompt:         prompt,
		Image:          imageField,
		Size:           volcengine.String("2K"),
		ResponseFormat: volcengine.String(model.GenerateImagesResponseFormatURL),
		Watermark:      volcengine.Bool(false),
	}

	resp, err := a.volcService.Client.GenerateImages(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Data) == 0 || resp.Data[0] == nil || resp.Data[0].Url == nil {
		return "", fmt.Errorf("image generation returned empty result")
	}

	local, err := services.DownloadAsset(*resp.Data[0].Url, ".png")
	if err != nil {
		return "", err
	}
	return local, nil
}

func (a *App) callStoryboardDecomposeLLM(provider string, apiKey string, baseURL string, modelID string, sourceText string) (*llmDecomposeResponse, error) {
	systemPrompt := `你是影视分镜结构化助手。将输入文案严格拆解为分镜JSON，确保字段完整并可用于后续视频生产流水线。要求：
1) 输出必须是JSON，且必须符合给定schema；
2) 每个分镜必须包含镜号、景别、运镜、画面内容、角色、场景、特殊元素、风格、声音设计、预估时长；
3) 角色/场景/元素/风格均需包含 id/name/prompt；
4) 同一场景的不同拍摄角度应视为不同场景，使用不同 id；
5) estimated_duration 仅允许 5 或 10。`

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"shots": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"shot_no":            map[string]interface{}{"type": "string"},
						"shot_size":          map[string]interface{}{"type": "string"},
						"camera_movement":    map[string]interface{}{"type": "string"},
						"frame_content":      map[string]interface{}{"type": "string"},
						"characters":         entityRefSchema(),
						"scenes":             entityRefSchema(),
						"special_elements":   entityRefSchema(),
						"visual_style":       entityRefSchema(),
						"sound_design":       map[string]interface{}{"type": "string"},
						"estimated_duration": map[string]interface{}{"type": "integer", "enum": []int{5, 10}},
					},
					"required": []string{
						"shot_no",
						"shot_size",
						"camera_movement",
						"frame_content",
						"characters",
						"scenes",
						"special_elements",
						"visual_style",
						"sound_design",
						"estimated_duration",
					},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"shots"},
		"additionalProperties": false,
	}

	userPrompt := "请将以下分镜需求转换为结构化数据：\n\n" + sourceText
	raw, err := a.requestDecomposeLLM(provider, apiKey, baseURL, modelID, systemPrompt, userPrompt, schema)
	if err != nil {
		return nil, err
	}
	raw = stripMarkdownJSONFence(raw)

	var decoded llmDecomposeResponse
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return nil, fmt.Errorf("invalid llm json output: %w", err)
	}
	return &decoded, nil
}

func entityRefSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "array",
		"items": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":     map[string]interface{}{"type": "string"},
				"name":   map[string]interface{}{"type": "string"},
				"prompt": map[string]interface{}{"type": "string"},
			},
			"required":             []string{"id", "name", "prompt"},
			"additionalProperties": false,
		},
	}
}

func stripMarkdownJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func (a *App) requestDecomposeLLM(provider string, apiKey string, baseURL string, modelID string, systemPrompt string, userPrompt string, schema map[string]interface{}) (string, error) {
	switch provider {
	case "ark_default":
		return requestWithArkClient(a.volcService.Client, modelID, systemPrompt, userPrompt, schema)
	case "ark_custom":
		key := strings.TrimSpace(apiKey)
		if key == "" {
			return "", fmt.Errorf("custom ark provider requires api_key")
		}
		url := strings.TrimSpace(baseURL)
		if url == "" {
			url = "https://ark.cn-beijing.volces.com/api/v3"
		}
		client := arkruntime.NewClientWithApiKey(
			key,
			arkruntime.WithBaseUrl(url),
		)
		return requestWithArkClient(client, modelID, systemPrompt, userPrompt, schema)
	case "openai_compatible":
		key := strings.TrimSpace(apiKey)
		url := strings.TrimSpace(baseURL)
		if key == "" {
			return "", fmt.Errorf("openai compatible provider requires api_key")
		}
		if url == "" {
			return "", fmt.Errorf("openai compatible provider requires base_url")
		}
		return requestWithOpenAICompatible(key, url, modelID, systemPrompt, userPrompt, schema)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func requestWithArkClient(client *arkruntime.Client, modelID string, systemPrompt string, userPrompt string, schema map[string]interface{}) (string, error) {
	req := model.CreateChatCompletionRequest{
		Model: modelID,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: "system",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(systemPrompt),
				},
			},
			{
				Role: "user",
				Content: &model.ChatCompletionMessageContent{
					StringValue: volcengine.String(userPrompt),
				},
			},
		},
		Temperature: volcengine.Float32(0.2),
		ResponseFormat: &model.ResponseFormat{
			Type: model.ResponseFormatJSONSchema,
			JSONSchema: &model.ResponseFormatJSONSchemaJSONSchemaParam{
				Name:        "storyboard_decompose",
				Description: "Storyboard decomposition JSON",
				Schema:      schema,
				Strict:      true,
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 || resp.Choices[0] == nil || resp.Choices[0].Message.Content == nil || resp.Choices[0].Message.Content.StringValue == nil {
		return "", fmt.Errorf("empty llm response")
	}
	return strings.TrimSpace(*resp.Choices[0].Message.Content.StringValue), nil
}

func requestWithOpenAICompatible(apiKey string, baseURL string, modelID string, systemPrompt string, userPrompt string, schema map[string]interface{}) (string, error) {
	type openAIMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type openAIResponseFormat struct {
		Type       string                 `json:"type"`
		JSONSchema map[string]interface{} `json:"json_schema,omitempty"`
	}
	payload := map[string]interface{}{
		"model": modelID,
		"messages": []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		"temperature": 0.2,
		"response_format": openAIResponseFormat{
			Type: "json_schema",
			JSONSchema: map[string]interface{}{
				"name":        "storyboard_decompose",
				"description": "Storyboard decomposition JSON",
				"schema":      schema,
				"strict":      true,
			},
		},
	}

	requestBody, _ := json.Marshal(payload)
	endpoint := strings.TrimRight(baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{Timeout: 120 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("invalid openai compatible response: %w", err)
	}

	if resp.StatusCode >= 300 {
		if errObj, ok := parsed["error"].(map[string]interface{}); ok {
			if msg, ok := errObj["message"].(string); ok && strings.TrimSpace(msg) != "" {
				return "", fmt.Errorf("provider error: %s", msg)
			}
		}
		return "", fmt.Errorf("provider error: http %d", resp.StatusCode)
	}

	choices, ok := parsed["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("empty llm response")
	}
	choice, _ := choices[0].(map[string]interface{})
	message, _ := choice["message"].(map[string]interface{})
	contentValue, exists := message["content"]
	if !exists {
		return "", fmt.Errorf("empty llm response")
	}

	if text, ok := contentValue.(string); ok {
		return strings.TrimSpace(text), nil
	}
	// Some OpenAI-compatible providers return content parts array.
	if parts, ok := contentValue.([]interface{}); ok {
		var out strings.Builder
		for _, p := range parts {
			partMap, _ := p.(map[string]interface{})
			if partMap == nil {
				continue
			}
			if t, _ := partMap["type"].(string); t == "text" {
				if txt, _ := partMap["text"].(string); txt != "" {
					out.WriteString(txt)
				}
			}
		}
		result := strings.TrimSpace(out.String())
		if result != "" {
			return result, nil
		}
	}

	return "", fmt.Errorf("empty llm response")
}

func loadCatalogMap(projectID uint) map[string]models.AssetCatalog {
	var catalogs []models.AssetCatalog
	models.DB.Preload("Versions", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at asc")
	}).Where("project_id = ?", projectID).Find(&catalogs)
	out := make(map[string]models.AssetCatalog, len(catalogs))
	for _, c := range catalogs {
		key := c.AssetType + ":" + c.AssetCode
		out[key] = c
	}
	return out
}

func chooseActiveCatalogVersionModel(versions []models.AssetVersion) *models.AssetVersion {
	if len(versions) == 0 {
		return nil
	}
	for i := len(versions) - 1; i >= 0; i-- {
		if versions[i].IsGood {
			v := versions[i]
			return &v
		}
	}
	v := versions[len(versions)-1]
	return &v
}

func autoSplitFirstPart(content string) string {
	parts := splitBySentence(content)
	if len(parts) <= 1 {
		runes := []rune(content)
		mid := len(runes) / 2
		return strings.TrimSpace(string(runes[:mid]))
	}
	mid := len(parts) / 2
	return strings.TrimSpace(strings.Join(parts[:mid], "。"))
}

func autoSplitSecondPart(content string) string {
	parts := splitBySentence(content)
	if len(parts) <= 1 {
		runes := []rune(content)
		mid := len(runes) / 2
		return strings.TrimSpace(string(runes[mid:]))
	}
	mid := len(parts) / 2
	return strings.TrimSpace(strings.Join(parts[mid:], "。"))
}

func splitBySentence(s string) []string {
	segments := strings.FieldsFunc(strings.TrimSpace(s), func(r rune) bool {
		return r == '。' || r == '！' || r == '？' || r == '\n'
	})
	out := make([]string, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func csvBytesToMarkdown(b []byte, comma rune) (string, error) {
	reader := csv.NewReader(bytes.NewReader(b))
	reader.Comma = comma
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}
	if len(records) == 0 {
		return "", nil
	}
	maxCols := 0
	for _, row := range records {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	for i := range records {
		for len(records[i]) < maxCols {
			records[i] = append(records[i], "")
		}
	}

	var sb strings.Builder
	writeRow := func(row []string) {
		sb.WriteString("|")
		for _, cell := range row {
			cell = strings.ReplaceAll(strings.TrimSpace(cell), "\n", " ")
			cell = strings.ReplaceAll(cell, "|", "\\|")
			sb.WriteString(" " + cell + " |")
		}
		sb.WriteString("\n")
	}

	writeRow(records[0])
	sb.WriteString("|")
	for i := 0; i < maxCols; i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")
	for i := 1; i < len(records); i++ {
		writeRow(records[i])
	}
	return sb.String(), nil
}

type xlsxWorkbook struct {
	Sheets []xlsxSheet `xml:"sheets>sheet"`
}

type xlsxSheet struct {
	Name string `xml:"name,attr"`
	ID   string `xml:"http://schemas.openxmlformats.org/officeDocument/2006/relationships id,attr"`
}

type xlsxRels struct {
	Relationships []xlsxRel `xml:"Relationship"`
}

type xlsxRel struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
}

type xlsxSharedStrings struct {
	Items []xlsxSI `xml:"si"`
}

type xlsxSI struct {
	T string    `xml:"t"`
	R []xlsxRun `xml:"r"`
}

type xlsxRun struct {
	T string `xml:"t"`
}

type xlsxWorksheet struct {
	Rows []xlsxRow `xml:"sheetData>row"`
}

type xlsxRow struct {
	Cells []xlsxCell `xml:"c"`
}

type xlsxCell struct {
	Ref string         `xml:"r,attr"`
	T   string         `xml:"t,attr"`
	V   string         `xml:"v"`
	IS  *xlsxInlineStr `xml:"is"`
}

type xlsxInlineStr struct {
	T string    `xml:"t"`
	R []xlsxRun `xml:"r"`
}

func xlsxToMarkdown(filePath string) (string, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()

	files := map[string]*zip.File{}
	for _, f := range zr.File {
		files[f.Name] = f
	}

	workbookData, err := readZipFile(files["xl/workbook.xml"])
	if err != nil {
		return "", fmt.Errorf("invalid xlsx workbook: %w", err)
	}
	var wb xlsxWorkbook
	if err := xml.Unmarshal(workbookData, &wb); err != nil {
		return "", err
	}
	if len(wb.Sheets) == 0 {
		return "", fmt.Errorf("xlsx has no sheet")
	}

	relsData, err := readZipFile(files["xl/_rels/workbook.xml.rels"])
	if err != nil {
		return "", err
	}
	var rels xlsxRels
	if err := xml.Unmarshal(relsData, &rels); err != nil {
		return "", err
	}
	relMap := map[string]string{}
	for _, rel := range rels.Relationships {
		relMap[rel.ID] = rel.Target
	}

	target := relMap[wb.Sheets[0].ID]
	if target == "" {
		return "", fmt.Errorf("cannot find first sheet")
	}
	sheetPath := path.Clean("xl/" + target)
	sheetData, err := readZipFile(files[sheetPath])
	if err != nil {
		return "", err
	}

	shared := []string{}
	if ssFile, ok := files["xl/sharedStrings.xml"]; ok {
		ssData, err := readZipFile(ssFile)
		if err == nil {
			var sst xlsxSharedStrings
			if xml.Unmarshal(ssData, &sst) == nil {
				for _, si := range sst.Items {
					if si.T != "" {
						shared = append(shared, si.T)
						continue
					}
					var b strings.Builder
					for _, run := range si.R {
						b.WriteString(run.T)
					}
					shared = append(shared, b.String())
				}
			}
		}
	}

	var ws xlsxWorksheet
	if err := xml.Unmarshal(sheetData, &ws); err != nil {
		return "", err
	}

	type rowData struct {
		cols map[int]string
	}
	rows := map[int]rowData{}
	maxRow := 0
	maxCol := 0

	for rowIdx, row := range ws.Rows {
		r := rowIdx + 1
		entry := rowData{cols: map[int]string{}}
		for _, cell := range row.Cells {
			col := excelColIndex(cell.Ref)
			if col < 0 {
				continue
			}
			val := strings.TrimSpace(cell.V)
			switch cell.T {
			case "s":
				idx, _ := strconv.Atoi(val)
				if idx >= 0 && idx < len(shared) {
					val = shared[idx]
				}
			case "inlineStr":
				if cell.IS != nil {
					if cell.IS.T != "" {
						val = cell.IS.T
					} else {
						var b strings.Builder
						for _, run := range cell.IS.R {
							b.WriteString(run.T)
						}
						val = b.String()
					}
				}
			}
			entry.cols[col] = strings.TrimSpace(val)
			if col > maxCol {
				maxCol = col
			}
		}
		rows[r] = entry
		if r > maxRow {
			maxRow = r
		}
	}
	if maxRow == 0 {
		return "", nil
	}

	table := make([][]string, 0, maxRow)
	for i := 1; i <= maxRow; i++ {
		row := make([]string, maxCol+1)
		if v, ok := rows[i]; ok {
			for c := 0; c <= maxCol; c++ {
				row[c] = strings.TrimSpace(v.cols[c])
			}
		}
		table = append(table, row)
	}

	var b strings.Builder
	writeRow := func(row []string) {
		b.WriteString("|")
		for _, cell := range row {
			cell = strings.ReplaceAll(strings.TrimSpace(cell), "\n", " ")
			cell = strings.ReplaceAll(cell, "|", "\\|")
			b.WriteString(" " + cell + " |")
		}
		b.WriteString("\n")
	}
	writeRow(table[0])
	b.WriteString("|")
	for i := 0; i <= maxCol; i++ {
		b.WriteString(" --- |")
	}
	b.WriteString("\n")
	for i := 1; i < len(table); i++ {
		writeRow(table[i])
	}
	return b.String(), nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	if f == nil {
		return nil, fmt.Errorf("file not found in zip")
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func excelColIndex(ref string) int {
	if ref == "" {
		return -1
	}
	letters := make([]rune, 0, len(ref))
	for _, r := range ref {
		if r >= 'A' && r <= 'Z' {
			letters = append(letters, r)
			continue
		}
		if r >= 'a' && r <= 'z' {
			letters = append(letters, r-32)
			continue
		}
		break
	}
	if len(letters) == 0 {
		return -1
	}
	col := 0
	for _, r := range letters {
		col = col*26 + int(r-'A'+1)
	}
	return col - 1
}
