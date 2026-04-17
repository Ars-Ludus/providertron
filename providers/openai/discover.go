package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type modelsListResponse struct {
	Data []openAIModelInfo `json:"data"`
}

type openAIModelInfo struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

func (o *OpenAI) FetchModels(ctx context.Context) ([]capability.ModelInfo, error) {
	slog.Default().With("pkg", "openai").Info("fetching models")

	var raw modelsListResponse
	if err := o.doGet(ctx, "/v1/models", &raw); err != nil {
		return nil, fmt.Errorf("openai.FetchModels: %w", err)
	}

	models := make([]capability.ModelInfo, 0, len(raw.Data))
	for _, m := range raw.Data {
		models = append(models, capability.ModelInfo{
			ID:           m.ID,
			Provider:     "openai",
			Capabilities: inferOpenAICapabilities(m.ID),
		})
	}

	slog.Default().With("pkg", "openai").Info("fetched models", "count", len(models))
	return models, nil
}

func inferOpenAICapabilities(id string) []capability.CapabilityType {
	switch {
	case hasPrefix(id, "text-embedding"):
		return []capability.CapabilityType{capability.CapabilityEmbed}
	case hasPrefix(id, "whisper"):
		return []capability.CapabilityType{capability.CapabilityVoice}
	case hasPrefix(id, "tts"):
		return []capability.CapabilityType{capability.CapabilityVoice}
	case hasPrefix(id, "gpt"), hasPrefix(id, "o1"), hasPrefix(id, "o3"), hasPrefix(id, "o4"):
		return []capability.CapabilityType{capability.CapabilityGenerate, capability.CapabilityStream}
	default:
		return []capability.CapabilityType{capability.CapabilityGenerate}
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
