package services

import (
	"context"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

type VolcEngineService struct {
	Client *arkruntime.Client
}

func NewVolcEngineService() *VolcEngineService {
	apiKey := os.Getenv("ARK_API_KEY")
	client := arkruntime.NewClientWithApiKey(
		apiKey,
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
	return &VolcEngineService{Client: client}
}

func (s *VolcEngineService) SetAPIKey(apiKey string) {
	s.Client = arkruntime.NewClientWithApiKey(
		apiKey,
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)
}

func (s *VolcEngineService) CreateVideoTask(modelID string, prompt string, firstFrameURL string, lastFrameURL string, ratio string, duration int, generateAudio bool, serviceTier string, expiresAfter int64) (string, error) {
	ctx := context.Background()

	contentItems := []*model.CreateContentGenerationContentItem{
		{
			Type: model.ContentGenerationContentItemTypeText,
			Text: volcengine.String(prompt),
		},
	}

	if firstFrameURL != "" {
		contentItems = append(contentItems, &model.CreateContentGenerationContentItem{
			Type: model.ContentGenerationContentItemTypeImage,
			ImageURL: &model.ImageURL{
				URL: firstFrameURL,
			},
			Role: volcengine.String("first_frame"),
		})
	}

	if lastFrameURL != "" {
		contentItems = append(contentItems, &model.CreateContentGenerationContentItem{
			Type: model.ContentGenerationContentItemTypeImage,
			ImageURL: &model.ImageURL{
				URL: lastFrameURL,
			},
			Role: volcengine.String("last_frame"),
		})
	}

	req := model.CreateContentGenerationTaskRequest{
		Model:           modelID,
		Content:         contentItems,
		Watermark:       volcengine.Bool(false),
		ReturnLastFrame: volcengine.Bool(true),
		GenerateAudio:   volcengine.Bool(generateAudio),
	}

	// Handle Ratio
	if ratio != "" {
		req.Ratio = volcengine.String(ratio)
	} else {
		req.Ratio = volcengine.String("adaptive")
	}

	// Handle Duration
	if duration > 0 {
		req.Duration = volcengine.Int64(int64(duration))
	}

	// Handle Service Tier
	// Only set ServiceTier when explicitly "flex", otherwise let API use default behavior
	if serviceTier == "flex" {
		req.ServiceTier = volcengine.String("flex")
		// Use provided expiresAfter or default to 86400 (24h)
		if expiresAfter > 0 {
			req.ExecutionExpiresAfter = volcengine.Int64(expiresAfter)
		} else {
			req.ExecutionExpiresAfter = volcengine.Int64(86400)
		}
	}
	// When serviceTier is "standard" or empty, do NOT set ServiceTier field

	// Helper for 1.5 Pro audio generation if needed, but defaults are usually fine.
	// Documentation says 'Seedance 1.5 pro' supports GenerateAudio.
	// Let's assume default behavior for now or add it if requested.

	resp, err := s.Client.CreateContentGenerationTask(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (s *VolcEngineService) GetTaskStatus(taskID string) (*model.GetContentGenerationTaskResponse, error) {
	ctx := context.Background()
	req := model.GetContentGenerationTaskRequest{
		ID: taskID,
	}
	// Note: The SDK might have a different method signature or return type,
	// but based on typical usage it should be GetContentGenerationTask.
	// If SDK is very new, it might match the Create one.
	resp, err := s.Client.GetContentGenerationTask(ctx, req)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
