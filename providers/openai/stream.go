package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Ars-Ludus/providertron/capability"
)

type streamChunkResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

func (o *OpenAI) Stream(ctx context.Context, req capability.StreamRequest) (<-chan capability.StreamChunk, error) {
	model := o.modelName(req.Model)
	params := o.cfg.chatParams(req.Temperature, req.TopP, req.MaxTokens)

	body := chatRequest{
		Model:       model,
		Temperature: params["temperature"].(float64),
		TopP:        params["top_p"].(float64),
		MaxTokens:   params["max_tokens"].(int),
		Stop:        req.Stop,
		Stream:      true,
	}

	for _, m := range req.Messages {
		body.Messages = append(body.Messages, chatMessage{Role: m.Role, Content: m.Content})
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai.Stream: marshal request: %w", err)
	}

	url := o.cfg.BaseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("openai.Stream: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(o.cfg.HTTPTimeout*10) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai.Stream: http: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("openai.Stream: api error %d", resp.StatusCode)
	}

	slog.Default().With("pkg", "openai").Info("stream started", "model", model)

	ch := make(chan capability.StreamChunk, 16)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- capability.StreamChunk{Done: true}
				return
			}

			var chunk streamChunkResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("openai.Stream: decode chunk: %w", err)}
				return
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					select {
					case ch <- capability.StreamChunk{Delta: delta}:
					case <-ctx.Done():
						ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("openai.Stream: %w", ctx.Err())}
						return
					}
				}
				if chunk.Choices[0].FinishReason != nil {
					ch <- capability.StreamChunk{Done: true}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("openai.Stream: read: %w", err)}
			return
		}
		ch <- capability.StreamChunk{Done: true}
	}()

	return ch, nil
}
