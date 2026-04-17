package openai

import (
	"fmt"
)

const (
	defaultBaseURL     = "https://api.openai.com"
	defaultModel       = "gpt-4o"
	defaultEmbedModel  = "text-embedding-3-small"
	defaultVoiceModel  = "tts-1"
	defaultSTTModel    = "whisper-1"
	defaultTemperature = 0.7
	defaultTopP        = 1.0
	defaultMaxTokens   = 4096
	defaultHTTPTimeout = 30
)

type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Temperature float64
	TopP        float64
	MaxTokens   int

	EmbedModel  string
	VoiceModel  string
	STTModel    string

	ModelsPath  string
	HTTPTimeout int // in seconds
}

func (c *Config) ProviderName() string { return "openai" }

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("openai: api_key is required")
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
	if c.EmbedModel == "" {
		c.EmbedModel = defaultEmbedModel
	}
	if c.VoiceModel == "" {
		c.VoiceModel = defaultVoiceModel
	}
	if c.STTModel == "" {
		c.STTModel = defaultSTTModel
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

func (c *Config) chatParams(temp float64, topP float64, maxTokens int) map[string]any {
	t := c.Temperature
	if temp != 0 {
		t = temp
	}
	p := c.TopP
	if topP != 0 {
		p = topP
	}
	m := c.MaxTokens
	if maxTokens != 0 {
		m = maxTokens
	}
	return map[string]any{
		"temperature": t,
		"top_p":       p,
		"max_tokens":  m,
	}
}
