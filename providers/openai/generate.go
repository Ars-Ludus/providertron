package openai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ars-Ludus/providertron/capability"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stop        []string      `json:"stop,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (o *OpenAI) Generate(ctx context.Context, req capability.GenerateRequest) (capability.GenerateResponse, error) {
	model := o.modelName(req.Model)
	params := o.cfg.chatParams(req.Temperature, req.TopP, req.MaxTokens)

	body := chatRequest{
		Model:       model,
		Temperature: params["temperature"].(float64),
		TopP:        params["top_p"].(float64),
		MaxTokens:   params["max_tokens"].(int),
		Stop:        req.Stop,
	}

	for _, m := range req.Messages {
		body.Messages = append(body.Messages, chatMessage{Role: m.Role, Content: m.Content})
	}

	slog.Default().With("pkg", "openai").Info("generate", "model", model)

	var raw chatResponse
	if err := o.doRequest(ctx, "/v1/chat/completions", body, &raw); err != nil {
		return capability.GenerateResponse{}, fmt.Errorf("openai.Generate: %w", err)
	}

	if len(raw.Choices) == 0 {
		return capability.GenerateResponse{}, fmt.Errorf("openai.Generate: empty response")
	}

	return capability.GenerateResponse{
		ID:      raw.ID,
		Content: raw.Choices[0].Message.Content,
		Model:   raw.Model,
		Usage: capability.UsageInfo{
			InputTokens:  raw.Usage.PromptTokens,
			OutputTokens: raw.Usage.CompletionTokens,
		},
	}, nil
}
