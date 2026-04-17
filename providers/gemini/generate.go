package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type generateRequestBody struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig  map[string]any        `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string        `json:"role,omitempty"`
	Parts []geminiPart  `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type generateResponse struct {
	Candidates []struct {
		Content geminiContent `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
	ModelVersion string `json:"modelVersion"`
}

func (g *Gemini) Generate(ctx context.Context, req capability.GenerateRequest) (capability.GenerateResponse, error) {
	model := g.modelName(req.Model)
	url := g.buildURL("/v1beta/models/" + model + ":generateContent")

	body := generateRequestBody{
		GenerationConfig: g.cfg.generationConfig(req.Temperature, req.TopK, req.TopP, req.MaxTokens),
	}

	for _, m := range req.Messages {
		switch m.Role {
		case "system":
			body.SystemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: m.Content}},
			}
		default:
			role := m.Role
			if role == "assistant" {
				role = "model"
			}
			body.Contents = append(body.Contents, geminiContent{
				Role:  role,
				Parts: []geminiPart{{Text: m.Content}},
			})
		}
	}

	slog.Default().With("pkg", "gemini").Info("generate", "model", model)

	var raw generateResponse
	if err := g.doRequest(ctx, url, body, &raw); err != nil {
		return capability.GenerateResponse{}, fmt.Errorf("gemini.Generate: %w", err)
	}

	if len(raw.Candidates) == 0 || len(raw.Candidates[0].Content.Parts) == 0 {
		return capability.GenerateResponse{}, fmt.Errorf("gemini.Generate: empty response")
	}

	return capability.GenerateResponse{
		Content: raw.Candidates[0].Content.Parts[0].Text,
		Model:   raw.ModelVersion,
		Usage: capability.UsageInfo{
			InputTokens:  raw.UsageMetadata.PromptTokenCount,
			OutputTokens: raw.UsageMetadata.CandidatesTokenCount,
		},
	}, nil
}
