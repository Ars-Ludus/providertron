package anthropic

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

type Anthropic struct {
	cfg    *Config
	client *http.Client
}

func New(cfg *Config) (*Anthropic, error) {
	if cfg == nil {
		return nil, fmt.Errorf("anthropic: config must not be nil")
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Anthropic{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.HTTPTimeout) * time.Second,
		},
	}, nil
}

func (a *Anthropic) doRequest(ctx context.Context, path string, body any, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("anthropic: marshal request: %w", err)
	}

	url := a.cfg.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("anthropic: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.cfg.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	slog.Default().With("pkg", "anthropic").Debug("sending request", "path", path)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("anthropic: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("anthropic: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("anthropic: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("anthropic: decode response: %w", err)
	}
	return nil
}

func (a *Anthropic) doGet(ctx context.Context, path string, dst any) error {
	url := a.cfg.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("anthropic: build get request: %w", err)
	}
	req.Header.Set("x-api-key", a.cfg.APIKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("anthropic-beta", "models-2025-02-19")

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("anthropic: http get: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("anthropic: read get response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("anthropic: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("anthropic: decode get response: %w", err)
	}
	return nil
}

func (a *Anthropic) modelName(override string) string {
	if override != "" {
		return override
	}
	return a.cfg.Model
}
