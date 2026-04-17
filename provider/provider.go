package provider

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Ars-Ludus/providertron/capability"
)

type Config interface {
	ProviderName() string
	Validate() error
}

// If you get an error trying to reach a capability you know the provider has, you need to modify the config inside the providers/providername sub-dir.
type Backend interface {
	capability.Generator
	capability.Streamer
	capability.Embedder
	capability.Batcher
	capability.Voicer
	capability.Discoverer
}

type Provider struct {
	name    string
	backend Backend
}

func New(cfg Config, backend Backend) (*Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("provider: config must not be nil")
	}
	if backend == nil {
		return nil, fmt.Errorf("provider: backend must not be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("provider: %w", err)
	}
	return &Provider{
		name:    cfg.ProviderName(),
		backend: backend,
	}, nil
}

func (p *Provider) Name() string { return p.name }

func (p *Provider) Generate(ctx context.Context, req capability.GenerateRequest) (capability.GenerateResponse, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("generate")
	resp, err := p.backend.Generate(ctx, req)
	if err != nil {
		return capability.GenerateResponse{}, fmt.Errorf("provider(%s).Generate: %w", p.name, err)
	}
	return resp, nil
}

func (p *Provider) Stream(ctx context.Context, req capability.StreamRequest) (<-chan capability.StreamChunk, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("stream")
	ch, err := p.backend.Stream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("provider(%s).Stream: %w", p.name, err)
	}
	return ch, nil
}

func (p *Provider) Embed(ctx context.Context, req capability.EmbedRequest) (capability.EmbedResponse, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("embed")
	resp, err := p.backend.Embed(ctx, req)
	if err != nil {
		return capability.EmbedResponse{}, fmt.Errorf("provider(%s).Embed: %w", p.name, err)
	}
	return resp, nil
}

func (p *Provider) Batch(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("batch")
	resp, err := p.backend.Batch(ctx, req)
	if err != nil {
		return capability.BatchResponse{}, fmt.Errorf("provider(%s).Batch: %w", p.name, err)
	}
	return resp, nil
}

func (p *Provider) Voice(ctx context.Context, req capability.VoiceRequest) (capability.VoiceResponse, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("voice")
	resp, err := p.backend.Voice(ctx, req)
	if err != nil {
		return capability.VoiceResponse{}, fmt.Errorf("provider(%s).Voice: %w", p.name, err)
	}
	return resp, nil
}

func (p *Provider) FetchModels(ctx context.Context) ([]capability.ModelInfo, error) {
	slog.Default().With("pkg", "provider", "provider", p.name).Debug("fetch models")
	models, err := p.backend.FetchModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider(%s).FetchModels: %w", p.name, err)
	}
	return models, nil
}
