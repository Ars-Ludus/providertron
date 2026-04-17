package main

import (
	"fmt"

	"github.com/Ars-Ludus/providertron/provider"
	"github.com/Ars-Ludus/providertron/providers/anthropic"
	"github.com/Ars-Ludus/providertron/providers/gemini"
	"github.com/Ars-Ludus/providertron/providers/openai"
)

func buildProvider(cfg *CLIConfig) (*provider.Provider, error) {
	switch cfg.Provider {
	case "gemini":
		gcfg := &gemini.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			BaseURL:     cfg.BaseURL,
			Temperature: cfg.Temperature,
			TopK:        cfg.TopK,
			TopP:        cfg.TopP,
			MaxTokens:   cfg.MaxTokens,
			ModelsPath:  cfg.ModelsPath,
		}
		backend, err := gemini.New(gcfg)
		if err != nil {
			return nil, fmt.Errorf("wire: gemini: %w", err)
		}
		return provider.New(gcfg, backend)

	case "openai":
		ocfg := &openai.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			BaseURL:     cfg.BaseURL,
			Temperature: cfg.Temperature,
			TopP:        cfg.TopP,
			MaxTokens:   cfg.MaxTokens,
			ModelsPath:  cfg.ModelsPath,
		}
		backend, err := openai.New(ocfg)
		if err != nil {
			return nil, fmt.Errorf("wire: openai: %w", err)
		}
		return provider.New(ocfg, backend)

	case "anthropic":
		acfg := &anthropic.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			BaseURL:     cfg.BaseURL,
			Temperature: cfg.Temperature,
			TopK:        cfg.TopK,
			TopP:        cfg.TopP,
			MaxTokens:   cfg.MaxTokens,
			ModelsPath:  cfg.ModelsPath,
		}
		backend, err := anthropic.New(acfg)
		if err != nil {
			return nil, fmt.Errorf("wire: anthropic: %w", err)
		}
		return provider.New(acfg, backend)

	default:
		return nil, fmt.Errorf("wire: unknown provider %q (supported: gemini, openai, anthropic)", cfg.Provider)
	}
}
