package gemini

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

func (g *Gemini) Stream(ctx context.Context, req capability.StreamRequest) (<-chan capability.StreamChunk, error) {
	model := g.modelName(req.Model)
	url := g.buildURL("/v1beta/models/" + model + ":streamGenerateContent") + "&alt=sse"

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

	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini.Stream: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("gemini.Stream: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: time.Duration(g.cfg.HTTPTimeout*10) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini.Stream: http: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("gemini.Stream: api error %d", resp.StatusCode)
	}

	slog.Default().With("pkg", "gemini").Info("stream started", "model", model)

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

			var raw generateResponse
			if err := json.Unmarshal([]byte(data), &raw); err != nil {
				ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("gemini.Stream: decode chunk: %w", err)}
				return
			}

			if len(raw.Candidates) > 0 && len(raw.Candidates[0].Content.Parts) > 0 {
				delta := raw.Candidates[0].Content.Parts[0].Text
				select {
				case ch <- capability.StreamChunk{Delta: delta}:
				case <-ctx.Done():
					ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("gemini.Stream: %w", ctx.Err())}
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- capability.StreamChunk{Done: true, Err: fmt.Errorf("gemini.Stream: read: %w", err)}
			return
		}
		ch <- capability.StreamChunk{Done: true}
	}()

	return ch, nil
}
