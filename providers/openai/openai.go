package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type OpenAI struct {
	cfg    *Config
	client *http.Client
}

func New(cfg *Config) (*OpenAI, error) {
	if cfg == nil {
		return nil, fmt.Errorf("openai: config must not be nil")
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &OpenAI{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.HTTPTimeout) * time.Second,
		},
	}, nil
}

func (o *OpenAI) doRequest(ctx context.Context, path string, body any, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("openai: marshal request: %w", err)
	}

	url := o.cfg.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("openai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	slog.Default().With("pkg", "openai").Debug("sending request", "path", path)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("openai: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("openai: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("openai: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("openai: decode response: %w", err)
	}
	return nil
}

func (o *OpenAI) doGet(ctx context.Context, path string, dst any) error {
	url := o.cfg.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("openai: build get request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+o.cfg.APIKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("openai: http get: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("openai: read get response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("openai: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("openai: decode get response: %w", err)
	}
	return nil
}

func (o *OpenAI) modelName(override string) string {
	if override != "" {
		return override
	}
	return o.cfg.Model
}
