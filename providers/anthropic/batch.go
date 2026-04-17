package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/arsmorientem/providertron/capability"
)

type batchRequestItem struct {
	CustomID string          `json:"custom_id"`
	Params   messagesRequest `json:"params"`
}

type createBatchRequest struct {
	Requests []batchRequestItem `json:"requests"`
}

type batchCreateResponse struct {
	ID              string `json:"id"`
	ProcessingStatus string `json:"processing_status"`
}

type batchResultsResponse struct {
	Results []struct {
		CustomID string `json:"custom_id"`
		Result   struct {
			Type    string          `json:"type"`
			Message json.RawMessage `json:"message"`
			Error   *struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			} `json:"error,omitempty"`
		} `json:"result"`
	} `json:"results"`
}

func (a *Anthropic) Batch(ctx context.Context, req capability.BatchRequest) (capability.BatchResponse, error) {
	if req.CapType != capability.CapabilityGenerate {
		return capability.BatchResponse{}, fmt.Errorf(
			"anthropic.Batch: only \"generate\" cap_type is supported; "+
				"Anthropic has no embedding batch API: %w",
			capability.ErrCapabilityUnavailable,
		)
	}

	model := a.modelName(req.Model)
	params := a.cfg.messageParams(0, 0, 0, 0)

	var items []batchRequestItem
	for _, item := range req.Items {
		items = append(items, batchRequestItem{
			CustomID: item.ID,
			Params: messagesRequest{
				Model:       model,
				MaxTokens:   params["max_tokens"].(int),
				Temperature: params["temperature"].(float64),
				Messages:    []anthropicMessage{{Role: "user", Content: item.Input}},
			},
		})
	}

	slog.Default().With("pkg", "anthropic").Info("batch submit", "model", model, "count", len(items))

	var createResp batchCreateResponse
	if err := a.doRequest(ctx, "/v1/messages/batches", createBatchRequest{Requests: items}, &createResp); err != nil {
		return capability.BatchResponse{}, fmt.Errorf("anthropic.Batch: submit: %w", err)
	}

	slog.Default().With("pkg", "anthropic").Info("batch submitted", "id", createResp.ID, "status", createResp.ProcessingStatus)

	batchID := createResp.ID
	for {
		select {
		case <-ctx.Done():
			return capability.BatchResponse{}, fmt.Errorf("anthropic.Batch: %w", ctx.Err())
		default:
		}

		var status struct {
			ProcessingStatus string `json:"processing_status"`
			ResultsURL       string `json:"results_url,omitempty"`
		}
		if err := a.doGet(ctx, "/v1/messages/batches/"+batchID, &status); err != nil {
			return capability.BatchResponse{}, fmt.Errorf("anthropic.Batch: poll: %w", err)
		}

		if status.ProcessingStatus == "ended" {
			break
		}

		slog.Default().With("pkg", "anthropic").Debug("batch polling", "id", batchID, "status", status.ProcessingStatus)
		select {
		case <-ctx.Done():
			return capability.BatchResponse{}, fmt.Errorf("anthropic.Batch: %w", ctx.Err())
		default:
		}
	}

	var resultsResp batchResultsResponse
	if err := a.doGet(ctx, "/v1/messages/batches/"+batchID+"/results", &resultsResp); err != nil {
		return capability.BatchResponse{}, fmt.Errorf("anthropic.Batch: fetch results: %w", err)
	}

	resultsByID := make(map[string]capability.BatchResultItem, len(resultsResp.Results))
	for _, r := range resultsResp.Results {
		item := capability.BatchResultItem{ID: r.CustomID}
		if r.Result.Error != nil {
			item.Err = r.Result.Error.Message
		} else {
			var msg messagesResponse
			if err := json.Unmarshal(r.Result.Message, &msg); err != nil {
				item.Err = fmt.Sprintf("decode result: %v", err)
			} else {
				var text string
				for _, block := range msg.Content {
					if block.Type == "text" {
						text += block.Text
					}
				}
				item.Result = capability.GenerateResponse{
					ID:      msg.ID,
					Content: text,
					Model:   msg.Model,
					Usage: capability.UsageInfo{
						InputTokens:  msg.Usage.InputTokens,
						OutputTokens: msg.Usage.OutputTokens,
					},
				}
			}
		}
		resultsByID[r.CustomID] = item
	}

	results := make([]capability.BatchResultItem, len(req.Items))
	for i, item := range req.Items {
		if r, ok := resultsByID[item.ID]; ok {
			results[i] = r
		} else {
			results[i] = capability.BatchResultItem{ID: item.ID, Err: "no result returned"}
		}
	}

	return capability.BatchResponse{Items: results, Model: model}, nil
}
