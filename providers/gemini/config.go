package gemini

import (
	"fmt"
)

const (
	defaultBaseURL      = "https://generativelanguage.googleapis.com"
	defaultModel        = "gemini-3.1-flash-preview"
	defaultEmbedModel   = "text-embedding-001"
	defaultTemperature  = 0.7
	defaultTopK         = 40
	defaultTopP         = 0.95
	defaultMaxTokens    = 8192
	defaultHTTPTimeout  = 30
)

type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	Temperature float64
	TopK        int
	TopP        float64
	MaxTokens   int

	EmbedModel  string
	VoiceModel  string 

	ModelsPath  string 
	HTTPTimeout int    // in seconds seconds
}

func (c *Config) ProviderName() string { return "gemini" }

func (c *Config) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("gemini: api_key is required")
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
	if c.Temperature == 0 {
		c.Temperature = defaultTemperature
	}
	if c.TopK == 0 {
		c.TopK = defaultTopK
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

func (c *Config) generationConfig(temp float64, topK int, topP float64, maxTokens int) map[string]any {
	t := c.Temperature
	if temp != 0 {
		t = temp
	}
	k := c.TopK
	if topK != 0 {
		k = topK
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
		"temperature":     t,
		"topK":            k,
		"topP":            p,
		"maxOutputTokens": m,
	}
}
