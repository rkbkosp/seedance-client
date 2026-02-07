package config

import (
	"embed"
	"encoding/json"
)

//go:embed models.json
var modelsFS embed.FS

// ModelPricing holds pricing information for a model (price per million tokens)
type ModelPricing struct {
	Standard      float64 `json:"standard"`
	StandardAudio float64 `json:"standard_audio,omitempty"`
	Flex          float64 `json:"flex"`
	FlexAudio     float64 `json:"flex_audio,omitempty"`
	PlatformPrice float64 `json:"platform_price"` // For savings calculation
}

// Model represents a video generation model configuration
type Model struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	SupportsAudio bool         `json:"supports_audio"`
	Default       bool         `json:"default"`
	Pricing       ModelPricing `json:"pricing"`
}

// ModelsConfig holds the list of available models
type ModelsConfig struct {
	Models []Model `json:"models"`
}

var loadedModels []Model
var modelMap map[string]*Model

func init() {
	LoadModels()
}

// LoadModels loads model configurations from embedded JSON
func LoadModels() error {
	data, err := modelsFS.ReadFile("models.json")
	if err != nil {
		return err
	}

	var config ModelsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	loadedModels = config.Models

	// Build lookup map
	modelMap = make(map[string]*Model)
	for i := range loadedModels {
		modelMap[loadedModels[i].ID] = &loadedModels[i]
	}

	return nil
}

// GetModels returns all available models
func GetModels() []Model {
	return loadedModels
}

// GetModelByID returns a model by its ID
func GetModelByID(id string) *Model {
	return modelMap[id]
}

// GetDefaultModel returns the default model
func GetDefaultModel() *Model {
	for i := range loadedModels {
		if loadedModels[i].Default {
			return &loadedModels[i]
		}
	}
	if len(loadedModels) > 0 {
		return &loadedModels[0]
	}
	return nil
}

// GetAudioSupportedModelIDs returns IDs of models that support audio generation
func GetAudioSupportedModelIDs() []string {
	var ids []string
	for _, m := range loadedModels {
		if m.SupportsAudio {
			ids = append(ids, m.ID)
		}
	}
	return ids
}

// GetPricePerMillion returns the price per million tokens for a take
func GetPricePerMillion(modelID string, serviceTier string, generateAudio bool) float64 {
	model := GetModelByID(modelID)
	if model == nil {
		return 0
	}

	if serviceTier == "flex" {
		if generateAudio && model.Pricing.FlexAudio > 0 {
			return model.Pricing.FlexAudio
		}
		return model.Pricing.Flex
	}

	// Standard tier
	if generateAudio && model.Pricing.StandardAudio > 0 {
		return model.Pricing.StandardAudio
	}
	return model.Pricing.Standard
}

// GetPlatformPrice returns the platform comparison price for a model
func GetPlatformPrice(modelID string) float64 {
	model := GetModelByID(modelID)
	if model == nil {
		return 0
	}
	return model.Pricing.PlatformPrice
}
