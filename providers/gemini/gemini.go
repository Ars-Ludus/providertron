package gemini

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

type Gemini struct {
	cfg    *Config
	client *http.Client
}

func New(cfg *Config) (*Gemini, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gemini: config must not be nil")
	}
	cfg.applyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Gemini{
		cfg: cfg,
		client: &http.Client{
			Timeout: time.Duration(cfg.HTTPTimeout) * time.Second,
		},
	}, nil
}

func (g *Gemini) buildURL(path string) string {
	return g.cfg.BaseURL + path + "?key=" + g.cfg.APIKey
}

func (g *Gemini) doRequest(ctx context.Context, url string, body any, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("gemini: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("gemini: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	slog.Default().With("pkg", "gemini").Debug("sending request", "url", url)

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gemini: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gemini: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gemini: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("gemini: decode response: %w", err)
	}
	return nil
}

func (g *Gemini) doGet(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("gemini: build get request: %w", err)
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("gemini: http get: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gemini: read get response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gemini: api error %d: %s", resp.StatusCode, string(raw))
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return fmt.Errorf("gemini: decode get response: %w", err)
	}
	return nil
}

func (g *Gemini) modelName(override string) string {
	if override != "" {
		return override
	}
	return g.cfg.Model
}
