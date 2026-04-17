package anthropic

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

type streamRequest struct {
	messagesRequest
	Stream bool `json:"stream"`
}

type sseEvent struct {
	Type  string          `json:"type"`
	Delta *anthropicDelta `json:"delta,omitempty"`
}

type anthropicDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (a *Anthropic) Stream(ctx context.Context, req capability.StreamRequest) (<-chan capability.StreamChunk, error) {
	model := a.modelName(req.Model)
	params := a.cfg.messageParams(req.Temperature, req.TopP, req.TopK, req.MaxTokens)

	body := streamRequest{
		messagesRequest: messagesRequest{
			Model:       model,
			MaxTokens:   params["max_tokens"].(int),
			Temperature: params["temperature"].(float64),
			TopP:        params["top_p"].(float64),
			StopSeqs:    req.Stop,
		},
		Stream: true,
	}

	if topK, ok := params["top_k"]; ok {
		body.messagesRequest.TopK = topK.(int)
	}

	for _, m := range req.Messages {
		if m.Role == "system" {
			body.messagesRequest.System = m.Content
			continue
		}
		body.messagesRequest.Messages = append(body.messagesRequest.Messages, anthropicMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic.Stream: marshal request: %w", err)
	}

	url := a.cfg.BaseURL + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("anthropic.Stream: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.cfg.APIKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(a.cfg.HTTPTimeout*10) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic.Stream: http: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic.Stream: api error %d", resp.StatusCode)
	}

	slog.Default().With("pkg", "anthropic").Info("stream started", "model", model)

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

			var event sseEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			switch event.Type {
			case "content_block_delta":
				if event.Delta != nil && event.Delta.Type == "text_delta" {
					select {
					case ch <- capability.StreamChunk{Delta: event.Delta.Text}:
					case <-ctx.Done():
						ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("anthropic.Stream: %w", ctx.Err())}
						return
					}
				}
			case "message_stop":
				ch <- capability.StreamChunk{Done: true}
				return
			case "error":
				ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("anthropic.Stream: server error event")}
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("anthropic.Stream: read: %w", err)}
			return
		}
		ch <- capability.StreamChunk{Done: true}
	}()

	return ch, nil
}
