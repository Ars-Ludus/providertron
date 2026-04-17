package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type anthropicModelsResponse struct {
	Data []struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"data"`
}

func (a *Anthropic) FetchModels(ctx context.Context) ([]capability.ModelInfo, error) {
	slog.Default().With("pkg", "anthropic").Info("fetching models")

	var raw anthropicModelsResponse
	if err := a.doGet(ctx, "/v1/models", &raw); err != nil {
		return nil, fmt.Errorf("anthropic.FetchModels: %w", err)
	}

	models := make([]capability.ModelInfo, 0, len(raw.Data))
	for _, m := range raw.Data {
		models = append(models, capability.ModelInfo{
			ID:           m.ID,
			Provider:     "anthropic",
			DisplayName:  m.DisplayName,
			Capabilities: inferAnthropicCapabilities(m.ID),
		})
	}

	slog.Default().With("pkg", "anthropic").Info("fetched models", "count", len(models))
	return models, nil
}

func inferAnthropicCapabilities(id string) []capability.CapabilityType {
	return []capability.CapabilityType{
		capability.CapabilityGenerate,
		capability.CapabilityStream,
		capability.CapabilityBatch,
	}
}
