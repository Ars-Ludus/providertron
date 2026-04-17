package anthropic

import (
	"fmt"
)

const (
	defaultBaseURL        = "https://api.anthropic.com"
	defaultModel          = "claude-sonnet-4-6"
	defaultTemperature    = 0.7
	defaultTopP           = 0.95
	defaultMaxTokens      = 8096
	defaultHTTPTimeout    = 30
	anthropicVersion      = "2023-06-01"
	anthropicBeta         = "max-tokens-3-5-sonnet-2024-07-15"
)

type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Temperature float64
	TopP        float64
	TopK        int
	MaxTokens   int // Why do you require this Anthropic T_T no one else does!

	ModelsPath  string
	HTTPTimeout int // in seconds
}

func (c *Config) ProviderName() string { return "anthropic" }

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("anthropic: api_key is required")
	}
	return nil
}

func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL
	}
	if c.Model == "" {
		c.Model = defaultModel
	}
	if c.Temperature == 0 {
		c.Temperature = defaultTemperature
	}
	if c.TopP == 0 {
		c.TopP = defaultTopP
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = defaultMaxTokens
	}
	if c.HTTPTimeout == 0 {
		c.HTTPTimeout = defaultHTTPTimeout
	}
}

func (c *Config) messageParams(temp float64, topP float64, topK int, maxTokens int) map[string]any {
	t := c.Temperature
	if temp != 0 {
		t = temp
	}
	p := c.TopP
	if topP != 0 {
		p = topP
	}
	k := c.TopK
	if topK != 0 {
		k = topK
	}
	m := c.MaxTokens
	if maxTokens != 0 {
		m = maxTokens
	}
	params := map[string]any{
		"temperature": t,
		"top_p":       p,
		"max_tokens":  m,
	}
	if k > 0 {
		params["top_k"] = k
	}
	return params
}
