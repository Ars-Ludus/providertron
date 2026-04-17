package main

import (
	"flag"
	"fmt"
	"os"
)

type CLIConfig struct {
	Provider 	string
	APIKey   	string
	Model    	string
	BaseURL  	string
	Temperature float64
	MaxTokens   int
	TopK        int
	TopP        float64
	Command 	string
	Prompt   	string
	InputFile 	string
	VoiceDir  	string
	AudioFile 	string
	OutputJSON 	bool
	AudioOut   	string
	ModelsPath 	string
	LogLevel 	string
}

func parseFlags() *CLIConfig {
	cfg := &CLIConfig{}

	flag.StringVar(&cfg.Provider, "provider", "", "AI provider: gemini|openai|anthropic (required)")
	flag.StringVar(&cfg.APIKey, "api-key", "", "Provider API key (required)")
	flag.StringVar(&cfg.Model, "model", "", "Model name (uses provider default if omitted)")
	flag.StringVar(&cfg.BaseURL, "base-url", "", "Override provider base URL")
	flag.Float64Var(&cfg.Temperature, "temperature", 0, "Sampling temperature (0 = provider default)")
	flag.IntVar(&cfg.MaxTokens, "max-tokens", 0, "Max output tokens (0 = provider default)")
	flag.IntVar(&cfg.TopK, "top-k", 0, "Top-K sampling (0 = provider default)")
	flag.Float64Var(&cfg.TopP, "top-p", 0, "Top-P nucleus sampling (0 = provider default)")
	flag.StringVar(&cfg.Command, "command", "", "Command: generate|stream|embed|batch|voice|discover|dispatch (required)")
	flag.StringVar(&cfg.Prompt, "prompt", "", "Inline text input")
	flag.StringVar(&cfg.InputFile, "input", "", "Path to JSON request file (overrides -prompt)")
	flag.StringVar(&cfg.VoiceDir, "voice-dir", "tts", "Voice direction: tts|stt")
	flag.StringVar(&cfg.AudioFile, "audio-file", "", "Path to audio file for STT input")
	flag.BoolVar(&cfg.OutputJSON, "output-json", false, "Print raw JSON response")
	flag.StringVar(&cfg.AudioOut, "audio-out", "", "Path to write TTS audio output (default: stdout bytes)")
	flag.StringVar(&cfg.ModelsPath, "models-path", "models.json", "Path to models cache file")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level: debug|info|warn|error")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: providertron -provider <name> -api-key <key> -command <cmd> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  generate   Single-turn text generation\n")
		fmt.Fprintf(os.Stderr, "  stream     Streaming text generation (prints chunks as they arrive)\n")
		fmt.Fprintf(os.Stderr, "  embed      Vector embedding for a single input\n")
		fmt.Fprintf(os.Stderr, "  batch      Batch generation or embedding from a JSON items file\n")
		fmt.Fprintf(os.Stderr, "  voice      Text-to-speech (tts) or speech-to-text (stt)\n")
		fmt.Fprintf(os.Stderr, "  discover   Query provider for available models and update models cache\n")
		fmt.Fprintf(os.Stderr, "  dispatch   Route a raw JSON request via the 'type' field\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	return cfg
}
