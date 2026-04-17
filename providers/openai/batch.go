package openai

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/arsmorientem/providertron/capability"
)

func (o *OpenAI) Batch(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	switch req.CapType {
	case capability.CapabilityEmbed:
		return o.batchEmbed(ctx, req)
	case capability.CapabilityGenerate:
		return o.batchGenerate(ctx, req)
	default:
		return capability.BatchResponse{}, fmt.Errorf("openai.Batch: unsupported cap_type %q", req.CapType)
	}
}

func (o *OpenAI) batchEmbed(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	model := req.Model
	if model == "" {
		model = o.cfg.EmbedModel
	}

	slog.Default().With("pkg", "openai").Info("batch embed (fan-out)", "model", model, "count", len(req.Items))

	results := make([]capability.BatchResultItem, len(req.Items))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, item := range req.Items {
		wg.Add(1)
		go func(idx int, it capability.BatchItem) {
			defer wg.Done()
			resp, err := o.Embed(ctx, capability.EmbedRequest{Input: it.Input, Model: model})
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[idx] = capability.BatchResultItem{ID: it.ID, Err: err.Error()}
			} else {
				results[idx] = capability.BatchResultItem{ID: it.ID, Result: resp}
			}
		}(i, item)
	}

	wg.Wait()
	return capability.BatchResponse{Items: results, Model: model}, nil
}

func (o *OpenAI) batchGenerate(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	model := o.modelName(req.Model)

	slog.Default().With("pkg", "openai").Info("batch generate (fan-out)", "model", model, "count", len(req.Items))

	results := make([]capability.BatchResultItem, len(req.Items))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i, item := range req.Items {
		wg.Add(1)
		go func(idx int, it capability.BatchItem) {
			defer wg.Done()
			genReq := capability.GenerateRequest{
				Messages: []capability.Message{{Role: "user", Content: it.Input}},
				Model:    model,
			}
			resp, err := o.Generate(ctx, genReq)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results[idx] = capability.BatchResultItem{ID: it.ID, Err: err.Error()}
			} else {
				results[idx] = capability.BatchResultItem{ID: it.ID, Result: resp}
			}
		}(i, item)
	}

	wg.Wait()
	return capability.BatchResponse{Items: results, Model: model}, nil
}
