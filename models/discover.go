package models

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/arsmorientem/providertron/capability"
)

// Here be the voodoo to drive model updates without updating the application. Yay less compiling.
func RunDiscovery(ctx context.Context, discoverers []capability.Discoverer, store *Store) error {
	if len(discoverers) == 0 {
		return fmt.Errorf("models.RunDiscovery: no discoverers provided")
	}

	type result struct {
		models []capability.ModelInfo
		err    error
	}

	results := make([]result, len(discoverers))
	var wg sync.WaitGroup

	for i, d := range discoverers {
		wg.Add(1)
		go func(idx int, disc capability.Discoverer) {
			defer wg.Done()
			models, err := disc.FetchModels(ctx)
			results[idx] = result{models: models, err: err}
		}(i, d)
	}

	wg.Wait()

	existing, err := store.Load()
	if err != nil {
		return fmt.Errorf("models.RunDiscovery: load store: %w", err)
	}

	var failCount int
	for _, r := range results {
		if r.err != nil {
			slog.Default().With("pkg", "models").Warn("provider discovery failed", "err", r.err)
			failCount++
			continue
		}

		incoming := capability.ModelsFile{
			Models: make(map[string]capability.ModelInfo, len(r.models)),
		}
		for _, m := range r.models {
			key := m.Provider + ":" + m.ID
			incoming.Models[key] = m
		}
		existing = Merge(existing, incoming)
	}

	if failCount == len(discoverers) {
		return fmt.Errorf("models.RunDiscovery: all providers failed")
	}

	if err := store.Save(existing); err != nil {
		return fmt.Errorf("models.RunDiscovery: save: %w", err)
	}

	slog.Default().With("pkg", "models").Info("discovery complete",
		"total_models", len(existing.Models),
		"failed_providers", failCount,
	)
	return nil
}
