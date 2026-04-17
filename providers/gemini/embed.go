package gemini

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type embedRequestBody struct {
	Model   string        `json:"model"`
	Content geminiContent `json:"content"`
	TaskType string       `json:"taskType,omitempty"`
}

type embedResponseBody struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

func (g *Gemini) Embed(ctx context.Context, req capability.EmbedRequest) (capability.EmbedResponse, error) {
	model := req.Model
	if model == "" {
		model = g.cfg.EmbedModel
	}
	url := g.buildURL("/v1beta/models/" + model + ":embedContent")

	body := embedRequestBody{
		Model:    "models/" + model,
		Content:  geminiContent{Parts: []geminiPart{{Text: req.Input}}},
		TaskType: req.TaskType,
	}

	slog.Default().With("pkg", "gemini").Info("embed", "model", model)

	var raw embedResponseBody
	if err := g.doRequest(ctx, url, body, &raw); err != nil {
		return capability.EmbedResponse{}, fmt.Errorf("gemini.Embed: %w", err)
	}

	return capability.EmbedResponse{
		Vector: raw.Embedding.Values,
		Model:  model,
	}, nil
}
