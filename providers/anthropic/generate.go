package anthropic

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type messagesRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	TopK        int                `json:"top_k,omitempty"`
	StopSeqs    []string           `json:"stop_sequences,omitempty"`
}

type messagesResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (a *Anthropic) Generate(ctx context.Context, req capability.GenerateRequest) (capability.GenerateResponse, error) {
	model := a.modelName(req.Model)
	params := a.cfg.messageParams(req.Temperature, req.TopP, req.TopK, req.MaxTokens)

	body := messagesRequest{
		Model:       model,
		MaxTokens:   params["max_tokens"].(int),
		Temperature: params["temperature"].(float64),
		TopP:        params["top_p"].(float64),
		StopSeqs:    req.Stop,
	}

	if topK, ok := params["top_k"]; ok {
		body.TopK = topK.(int)
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			body.System = m.Content
			continue
		}
		body.Messages = append(body.Messages, anthropicMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	slog.Default().With("pkg", "anthropic").Info("generate", "model", model)

	var raw messagesResponse
	if err := a.doRequest(ctx, "/v1/messages", body, &raw); err != nil {
		return capability.GenerateResponse{}, fmt.Errorf("anthropic.Generate: %w", err)
	}

	if len(raw.Content) == 0 {
		return capability.GenerateResponse{}, fmt.Errorf("anthropic.Generate: empty response")
	}

	var text string
	for _, block := range raw.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}

	return capability.GenerateResponse{
		ID:      raw.ID,
		Content: text,
		Model:   raw.Model,
		Usage: capability.UsageInfo{
			InputTokens:  raw.Usage.InputTokens,
			OutputTokens: raw.Usage.OutputTokens,
		},
	}, nil
}
