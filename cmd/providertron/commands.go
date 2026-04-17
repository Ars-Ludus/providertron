package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Ars-Ludus/providertron/capability"
	"github.com/Ars-Ludus/providertron/models"
	"github.com/Ars-Ludus/providertron/provider"
)

func runGenerate(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	req, err := buildGenerateRequest(cfg)
	if err != nil {
		return err
	}

	resp, err := p.Generate(ctx, req)
	if err != nil {
		return err
	}

	return printResult(cfg.OutputJSON, resp, resp.Content)
}

func runStream(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	req, err := buildStreamRequest(cfg)
	if err != nil {
		return err
	}

	ch, err := p.Stream(ctx, req)
	if err != nil {
		return err
	}

	for chunk := range ch {
		if chunk.Err != nil {
			return chunk.Err
		}
		if chunk.Done {
			break
		}
		if cfg.OutputJSON {
			b, _ := json.Marshal(chunk)
			fmt.Println(string(b))
		} else {
			fmt.Print(chunk.Delta)
		}
	}

	if !cfg.OutputJSON {
		fmt.Println()
	}
	return nil
}

func runEmbed(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	input, err := resolveTextInput(cfg)
	if err != nil {
		return err
	}

	req := capability.EmbedRequest{
		BaseRequest: capability.BaseRequest{Type: capability.CapabilityEmbed},
		Input:       input,
		Model:       cfg.Model,
	}

	resp, err := p.Embed(ctx, req)
	if err != nil {
		return err
	}

	if cfg.OutputJSON {
		b, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(b))
	} else {
		fmt.Printf("model: %s\n", resp.Model)
		fmt.Printf("dimensions: %d\n", len(resp.Vector))
		if len(resp.Vector) > 0 {
			preview := resp.Vector
			if len(preview) > 5 {
				preview = preview[:5]
			}
			fmt.Printf("vector (first %d): %v...\n", len(preview), preview)
		}
	}
	return nil
}

func runBatch(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	if cfg.InputFile == "" {
		return fmt.Errorf("batch command requires -input <file> with a JSON array of batch items")
	}

	data, err := os.ReadFile(cfg.InputFile)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	var items []capability.BatchItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("decode batch items: %w", err)
	}

	capType := capability.CapabilityGenerate
	if cfg.Command == "batch-embed" {
		capType = capability.CapabilityEmbed
	}

	req := capability.BatchRequest{
		BaseRequest: capability.BaseRequest{Type: capability.CapabilityBatch},
		Items:       items,
		Model:       cfg.Model,
		CapType:     capType,
	}

	resp, err := p.Batch(ctx, req)
	if err != nil {
		return err
	}

	return printResult(cfg.OutputJSON, resp, fmt.Sprintf("batch complete: %d items", len(resp.Items)))
}

func runVoice(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	req := capability.VoiceRequest{
		BaseRequest: capability.BaseRequest{Type: capability.CapabilityVoice},
		Direction:   cfg.VoiceDir,
		Model:       cfg.Model,
	}

	switch cfg.VoiceDir {
	case "tts":
		input, err := resolveTextInput(cfg)
		if err != nil {
			return err
		}
		req.Input = input

	case "stt":
		if cfg.AudioFile == "" {
			return fmt.Errorf("stt requires -audio-file <path>")
		}
		audio, err := os.ReadFile(cfg.AudioFile)
		if err != nil {
			return fmt.Errorf("read audio file: %w", err)
		}
		req.Audio = audio

	default:
		return fmt.Errorf("unknown -voice-dir %q, expected tts or stt", cfg.VoiceDir)
	}

	resp, err := p.Voice(ctx, req)
	if err != nil {
		if errors.Is(err, capability.ErrCapabilityUnavailable) {
			return fmt.Errorf("voice capability not available: %w", err)
		}
		return err
	}

	if cfg.VoiceDir == "tts" {
		if cfg.AudioOut != "" {
			if err := os.WriteFile(cfg.AudioOut, resp.Audio, 0644); err != nil {
				return fmt.Errorf("write audio output: %w", err)
			}
			fmt.Printf("audio written to %s (%d bytes)\n", cfg.AudioOut, len(resp.Audio))
		} else {
			os.Stdout.Write(resp.Audio)
		}
	} else {
		if cfg.OutputJSON {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
		} else {
			fmt.Println(resp.Text)
		}
	}
	return nil
}

func runDiscover(ctx context.Context, p *provider.Provider, store *models.Store) error {
	fmt.Printf("discovering models for provider: %s\n", p.Name())

	discoverers := []capability.Discoverer{p}
	if err := models.RunDiscovery(ctx, discoverers, store); err != nil {
		return err
	}

	f, err := store.Load()
	if err != nil {
		return fmt.Errorf("load updated store: %w", err)
	}

	fmt.Printf("models cache updated: %s\n", store.Path)
	fmt.Printf("total models: %d\n", len(f.Models))
	return nil
}

func runDispatch(ctx context.Context, p *provider.Provider, cfg *CLIConfig) error {
	var raw json.RawMessage

	if cfg.InputFile != "" {
		data, err := os.ReadFile(cfg.InputFile)
		if err != nil {
			return fmt.Errorf("read input file: %w", err)
		}
		raw = json.RawMessage(data)
	} else if cfg.Prompt != "" {
		raw = json.RawMessage(cfg.Prompt)
	} else {
		return fmt.Errorf("dispatch requires -input <file> or -prompt <raw-json>")
	}

	result, err := p.Dispatch(ctx, raw)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode result: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func buildGenerateRequest(cfg *CLIConfig) (capability.GenerateRequest, error) {
	req := capability.GenerateRequest{
		BaseRequest: capability.BaseRequest{Type: capability.CapabilityGenerate},
		Model:       cfg.Model,
		Temperature: cfg.Temperature,
		TopK:        cfg.TopK,
		TopP:        cfg.TopP,
		MaxTokens:   cfg.MaxTokens,
	}

	if cfg.InputFile != "" {
		data, err := os.ReadFile(cfg.InputFile)
		if err != nil {
			return req, fmt.Errorf("read input file: %w", err)
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return req, fmt.Errorf("decode generate request: %w", err)
		}
		return req, nil
	}

	if cfg.Prompt == "" {
		return req, fmt.Errorf("generate requires -prompt <text> or -input <file>")
	}

	req.Messages = []capability.Message{{Role: "user", Content: cfg.Prompt}}
	return req, nil
}

func buildStreamRequest(cfg *CLIConfig) (capability.StreamRequest, error) {
	genReq, err := buildGenerateRequest(cfg)
	if err != nil {
		return capability.StreamRequest{}, err
	}
	genReq.Type = capability.CapabilityStream
	return capability.StreamRequest{GenerateRequest: genReq}, nil
}

func resolveTextInput(cfg *CLIConfig) (string, error) {
	if cfg.InputFile != "" {
		data, err := os.ReadFile(cfg.InputFile)
		if err != nil {
			return "", fmt.Errorf("read input file: %w", err)
		}
		return string(data), nil
	}
	if cfg.Prompt != "" {
		return cfg.Prompt, nil
	}
	return "", fmt.Errorf("input required: use -prompt <text> or -input <file>")
}

func printResult(asJSON bool, v any, human string) error {
	if asJSON {
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("encode result: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}
	fmt.Println(human)
	return nil
}
