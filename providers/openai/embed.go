package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type embedRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
}

func (o *OpenAI) Embed(ctx context.Context, req capability.EmbedRequest) (capability.EmbedResponse, error) {
	model := req.Model
	if model == "" {
		model = o.cfg.EmbedModel
	}

	slog.Default().With("pkg", "openai").Info("embed", "model", model)

	body := embedRequest{Input: req.Input, Model: model}

	var raw embedResponse
	if err := o.doRequest(ctx, "/v1/embeddings", body, &raw); err != nil {
		return capability.EmbedResponse{}, fmt.Errorf("openai.Embed: %w", err)
	}

	if len(raw.Data) == 0 {
		return capability.EmbedResponse{}, fmt.Errorf("openai.Embed: empty response")
	}

	return capability.EmbedResponse{
		Vector: raw.Data[0].Embedding,
		Model:  raw.Model,
	}, nil
}
