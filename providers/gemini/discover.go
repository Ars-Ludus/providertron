package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ars-Ludus/providertron/capability"
)

type modelsListResponse struct {
	Models []geminiModelInfo `json:"models"`
}

type geminiModelInfo struct {
	Name                       string   `json:"name"`
	DisplayName                string   `json:"displayName"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	InputTokenLimit            int      `json:"inputTokenLimit"`
	OutputTokenLimit           int      `json:"outputTokenLimit"`
}

func (g *Gemini) FetchModels(ctx context.Context) ([]capability.ModelInfo, error) {
	url := g.cfg.BaseURL + "/v1beta/models?key=" + g.cfg.APIKey

	slog.Default().With("pkg", "gemini").Info("fetching models")

	var raw modelsListResponse
	if err := g.doGet(ctx, url, &raw); err != nil {
		return nil, fmt.Errorf("gemini.FetchModels: %w", err)
	}

	models := make([]capability.ModelInfo, 0, len(raw.Models))
	for _, m := range raw.Models {
		id := stripPrefix(m.Name, "models/")
		models = append(models, capability.ModelInfo{
			ID:           id,
			Provider:     "gemini",
			DisplayName:  m.DisplayName,
			Capabilities: methodsToCapabilities(m.SupportedGenerationMethods),
			InputTokens:  m.InputTokenLimit,
			OutputTokens: m.OutputTokenLimit,
		})
	}

	slog.Default().With("pkg", "gemini").Info("fetched models", "count", len(models))
	return models, nil
}

func stripPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

func methodsToCapabilities(methods []string) []capability.CapabilityType {
	var caps []capability.CapabilityType
	for _, m := range methods {
		switch m {
		case "generateContent":
			caps = append(caps, capability.CapabilityGenerate, capability.CapabilityStream)
		case "embedContent", "batchEmbedContents":
			caps = appendUnique(caps, capability.CapabilityEmbed)
		}
	}
	return caps
}

func appendUnique(caps []capability.CapabilityType, c capability.CapabilityType) []capability.CapabilityType {
	for _, existing := range caps {
		if existing == c {
			return caps
		}
	}
	return append(caps, c)
}
