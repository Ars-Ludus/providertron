package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/arsmorientem/providertron/capability"
)

type batchEmbedRequest struct {
	Requests []embedRequestBody `json:"requests"`
}

type batchEmbedResponse struct {
	Embeddings []struct {
		Values []float64 `json:"values"`
	} `json:"embeddings"`
}

func (g *Gemini) Batch(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	switch req.CapType {
	case capability.CapabilityEmbed:
		return g.batchEmbed(ctx, req)
	case capability.CapabilityGenerate:
		return g.batchGenerate(ctx, req)
	default:
		return capability.BatchResponse{}, fmt.Errorf("gemini.Batch: unsupported cap_type %q", req.CapType)
	}
}

func (g *Gemini) batchEmbed(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	model := req.Model
	if model == "" {
		model = g.cfg.EmbedModel
	}
	url := g.buildURL("/v1beta/models/" + model + ":batchEmbedContents")

	batchReq := batchEmbedRequest{}
	for _, item := range req.Items {
		batchReq.Requests = append(batchReq.Requests, embedRequestBody{
			Model:   "models/" + model,
			Content: geminiContent{Parts: []geminiPart{{Text: item.Input}}},
		})
	}

	slog.Default().With("pkg", "gemini").Info("batch embed", "model", model, "count", len(req.Items))

	var raw batchEmbedResponse
	if err := g.doRequest(ctx, url, batchReq, &raw); err != nil {
		return capability.BatchResponse{}, fmt.Errorf("gemini.Batch embed: %w", err)
	}

	results := make([]capability.BatchResultItem, len(req.Items))
	for i, item := range req.Items {
		if i >= len(raw.Embeddings) {
			results[i] = capability.BatchResultItem{ID: item.ID, Err: "no embedding returned"}
			continue
		}
		results[i] = capability.BatchResultItem{
			ID:     item.ID,
			Result: capability.EmbedResponse{Vector: raw.Embeddings[i].Values, Model: model},
		}
	}

	return capability.BatchResponse{Items: results, Model: model}, nil
}

func (g *Gemini) batchGenerate(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	model := g.modelName(req.Model)

	slog.Default().With("pkg", "gemini").Info("batch generate (fan-out)", "model", model, "count", len(req.Items))

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
			resp, err := g.Generate(ctx, genReq)
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
