package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arsmorientem/providertron/capability"
)

// Dispatch is the entry. JSON.
// unmarshals BaseRequest to read the routing token then unmarshals full typed req in matched branch.
func (p *Provider) Dispatch(ctx context.Context, raw json.RawMessage) (any, error) {
	var base capability.BaseRequest
	if err := json.Unmarshal(raw, &base); err != nil {
		return nil, fmt.Errorf("provider(%s).Dispatch: decode type field: %w", p.name, err)
	}

	switch base.Type {
	case capability.CapabilityGenerate:
		var req capability.GenerateRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("provider(%s).Dispatch: decode generate request: %w", p.name, err)
		}
		return p.Generate(ctx, req)

	case capability.CapabilityStream:
		var req capability.StreamRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("provider(%s).Dispatch: decode stream request: %w", p.name, err)
		}
		return p.Stream(ctx, req)

	case capability.CapabilityEmbed:
		var req capability.EmbedRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("provider(%s).Dispatch: decode embed request: %w", p.name, err)
		}
		return p.Embed(ctx, req)

	case capability.CapabilityBatch:
		var req capability.BatchRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("provider(%s).Dispatch: decode batch request: %w", p.name, err)
		}
		return p.Batch(ctx, req)

	case capability.CapabilityVoice:
		var req capability.VoiceRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("provider(%s).Dispatch: decode voice request: %w", p.name, err)
		}
		return p.Voice(ctx, req)

	default:
		return nil, fmt.Errorf("provider(%s).Dispatch: unknown capability type %q", p.name, base.Type)
	}
}
