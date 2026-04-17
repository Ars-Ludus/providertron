# Connecting providertron to Your Application

providertron is a Go library. You import it, wire up a provider, and talk to `*provider.Provider`. Your application never touches provider-specific packages after that.

---

## Installation

```bash
go get github.com/Ars-Ludus/providertron
```

---

## The Three-Step Pattern

Every integration follows the same three steps:

1. **Import** the `provider` package and whichever specific provider(s) you need
2. **Wire** — construct the backend and hand it to `provider.New()`
3. **Call** — use `*provider.Provider` methods anywhere in your app

```go
import (
    "github.com/Ars-Ludus/providertron/provider"
    "github.com/Ars-Ludus/providertron/providers/gemini"
)

// 1. Configure
cfg := &gemini.Config{
    APIKey: os.Getenv("GEMINI_API_KEY"),
    // All other fields have provider defaults — override only what you need
}

// 2. Wire
backend, err := gemini.New(cfg)
if err != nil { /* validation error — check APIKey, etc. */ }

p, err := provider.New(cfg, backend)
if err != nil { /* shouldn't happen if backend.New() succeeded */ }

// 3. Call
resp, err := p.Generate(ctx, capability.GenerateRequest{
    Messages: []capability.Message{
        {Role: "user", Content: "Hello"},
    },
})
```

After step 2, `p` is your application's only dependency. Switch providers by changing the wire step — nothing else changes.

---

## Request Types

All request types live in `github.com/Ars-Ludus/providertron/capability`.

### Generate (single-turn or multi-turn)

```go
import "github.com/Ars-Ludus/providertron/capability"

resp, err := p.Generate(ctx, capability.GenerateRequest{
    Messages: []capability.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user",   Content: "Explain Go interfaces in one paragraph."},
    },
    Model:       "gemini-2.5-pro",  // omit to use provider default
    Temperature: 0.7,               // omit to use provider default
    MaxTokens:   512,
})

fmt.Println(resp.Content)
fmt.Printf("tokens: %d in / %d out\n", resp.Usage.InputTokens, resp.Usage.OutputTokens)
```

### Stream

```go
ch, err := p.Stream(ctx, capability.StreamRequest{
    GenerateRequest: capability.GenerateRequest{
        Messages: []capability.Message{{Role: "user", Content: "Write a haiku"}},
    },
})
if err != nil { ... }

for chunk := range ch {
    if chunk.Err != nil { /* terminal error */ break }
    if chunk.Done      { break }
    fmt.Print(chunk.Delta)
}
fmt.Println()
```

### Embed

```go
resp, err := p.Embed(ctx, capability.EmbedRequest{
    Input: "The quick brown fox",
    // TaskType: "RETRIEVAL_DOCUMENT" // Gemini-specific hint; ignored by others
})

fmt.Printf("vector dimensions: %d\n", len(resp.Vector))
```

### Batch

Batch accepts a slice of `BatchItem` and a `CapType` to specify what operation to run on each item.

```go
resp, err := p.Batch(ctx, capability.BatchRequest{
    CapType: capability.CapabilityEmbed, // or capability.CapabilityGenerate
    Items: []capability.BatchItem{
        {ID: "a", Input: "First document"},
        {ID: "b", Input: "Second document"},
    },
})

for _, item := range resp.Items {
    if item.Err != "" {
        fmt.Printf("%s failed: %s\n", item.ID, item.Err)
        continue
    }
    embedResp := item.Result.(capability.EmbedResponse)
    fmt.Printf("%s: %d dims\n", item.ID, len(embedResp.Vector))
}
```

### Voice (TTS / STT)

```go
// Text-to-speech
resp, err := p.Voice(ctx, capability.VoiceRequest{
    Direction: "tts",
    Input:     "Hello from providertron",
    Voice:     "alloy", // OpenAI; "Kore" for Gemini
})
os.WriteFile("output.mp3", resp.Audio, 0644)

// Speech-to-text (OpenAI only)
audio, _ := os.ReadFile("recording.mp3")
resp, err := p.Voice(ctx, capability.VoiceRequest{
    Direction: "stt",
    Audio:     audio,
})
fmt.Println(resp.Text)
```

**Checking for unsupported capabilities:** When a provider doesn't offer a capability (e.g. Anthropic has no voice or embedding API), the error wraps `capability.ErrCapabilityUnavailable`. Check for it explicitly to distinguish misconfiguration from network errors:

```go
resp, err := p.Voice(ctx, req)
if errors.Is(err, capability.ErrCapabilityUnavailable) {
    // Provider doesn't offer this — switch providers or skip
    log.Warn("voice not available", "provider", p.Name())
    return
}
if err != nil { /* real error */ }
```

### JSON Dispatch

If your application receives raw JSON (e.g. from an HTTP handler or message queue), use `Dispatch`. It reads the `"type"` field and routes automatically:

```go
// The JSON payload decides which capability runs
raw := json.RawMessage(`{
    "type": "embedding",
    "input": "text to embed"
}`)

result, err := p.Dispatch(ctx, raw)

// result is one of: GenerateResponse, <-chan StreamChunk, EmbedResponse,
// BatchResponse, or VoiceResponse depending on "type"
embedResp := result.(capability.EmbedResponse)
```

Valid `"type"` values: `"generate"`, `"stream"`, `"embedding"`, `"batch"`, `"voice"`.

---

## Model Discovery

providertron is self-healing: it queries the provider's API for available models and writes a local cache. No code changes are needed when a provider releases or deprecates a model.

```go
import "github.com/Ars-Ludus/providertron/models"

store := &models.Store{Path: "./models.json"}

// Discover and cache — call this at startup or on a schedule
err := models.RunDiscovery(ctx, []capability.Discoverer{p}, store)

// Load the cache to inspect available models
file, err := store.Load()
for key, model := range file.Models {
    fmt.Printf("%s: caps=%v\n", key, model.Capabilities)
}
```

To discover from multiple providers at once, pass all of them:

```go
models.RunDiscovery(ctx, []capability.Discoverer{geminiProvider, openaiProvider}, store)
```

`RunDiscovery` fans out concurrently, merges results, and saves atomically. A failed provider is logged but does not block the others.

---

## Switching Providers

Only the wire step changes. Everything that calls `p.Generate()`, `p.Embed()`, etc. is untouched:

```go
// Before: Gemini
backend, _ := gemini.New(&gemini.Config{APIKey: os.Getenv("GEMINI_API_KEY")})
p, _ := provider.New(&gemini.Config{APIKey: os.Getenv("GEMINI_API_KEY")}, backend)

// After: OpenAI — rest of application unchanged
backend, _ := openai.New(&openai.Config{APIKey: os.Getenv("OPENAI_API_KEY")})
p, _ := provider.New(&openai.Config{APIKey: os.Getenv("OPENAI_API_KEY")}, backend)
```

---

## Provider Configuration Reference

Each provider's `Config` struct has the same shape. All fields except `APIKey` are optional — unset fields use the provider's default.

### Gemini (`providers/gemini`)

| Field | Default | Notes |
|---|---|---|
| `APIKey` | — | **Required** |
| `Model` | `gemini-3.1-flash-preview` | Generation model |
| `EmbedModel` | `text-embedding-001` | Embedding model |
| `VoiceModel` | `gemini-2.5-flash-preview-tts` | TTS model |
| `BaseURL` | `https://generativelanguage.googleapis.com` | Override for proxies |
| `Temperature` | `0.7` | |
| `TopK` | `40` | |
| `TopP` | `0.95` | |
| `MaxTokens` | `8192` | |
| `HTTPTimeout` | `30` (seconds) | |

### OpenAI (`providers/openai`)

| Field | Default | Notes |
|---|---|---|
| `APIKey` | — | **Required** |
| `Model` | `gpt-4o` | |
| `EmbedModel` | `text-embedding-3-small` | |
| `VoiceModel` | `tts-1` | TTS: `tts-1` or `tts-1-hd` |
| `STTModel` | `whisper-1` | |
| `BaseURL` | `https://api.openai.com` | Override for proxies or compatible APIs |
| `Temperature` | `0.7` | |
| `TopP` | `1.0` | |
| `MaxTokens` | `4096` | |
| `HTTPTimeout` | `30` (seconds) | |

### Anthropic (`providers/anthropic`)

| Field | Default | Notes |
|---|---|---|
| `APIKey` | — | **Required** |
| `Model` | `claude-sonnet-4-6` | |
| `BaseURL` | `https://api.anthropic.com` | |
| `Temperature` | `0.7` | |
| `TopP` | `0.95` | |
| `TopK` | `0` (disabled) | |
| `MaxTokens` | `8096` | Anthropic requires this field |
| `HTTPTimeout` | `30` (seconds) | |

**Anthropic capability notes:** Embed and Voice return `ErrCapabilityUnavailable` — Anthropic offers neither API. Batch uses the Anthropic Message Batches API (generate only).

---

## Capability Support Matrix

| Capability | Gemini | OpenAI | Anthropic |
|---|:---:|:---:|:---:|
| Generate | ✓ | ✓ | ✓ |
| Stream | ✓ | ✓ | ✓ |
| Embed | ✓ | ✓ | — |
| Batch | ✓ | ✓ | ✓ (generate only) |
| Voice TTS | ✓ | ✓ | — |
| Voice STT | — | ✓ | — |
| Discover | ✓ | ✓ | ✓ |

---

## Adding a New Provider

Each provider is a Go package under `providers/`. To add one:

1. Create `providers/myprovider/` with the same file layout:
   - `config.go` — `Config` struct implementing `provider.Config` (`ProviderName()`, `Validate()`, `applyDefaults()`)
   - `myprovider.go` — concrete struct + `New()` + shared HTTP helpers
   - `generate.go`, `stream.go`, `embed.go`, `batch.go`, `voice.go` — one file per capability
   - `discover.go` — `FetchModels()` implementation
2. For unsupported capabilities, return `ErrCapabilityUnavailable`:
   ```go
   func (m *MyProvider) Embed(_ context.Context, _ capability.EmbedRequest) (capability.EmbedResponse, error) {
       return capability.EmbedResponse{}, fmt.Errorf(
           "myprovider.Embed: provider does not offer embeddings: %w",
           capability.ErrCapabilityUnavailable,
       )
   }
   ```
3. Register it in `cmd/providertron/wire.go` under the switch statement.

The `capability` and `provider` packages never change — only the new provider package and `wire.go`.

---

## Logging

providertron uses `log/slog`. All packages call `slog.Default()` lazily, so setting the default logger in your application before the first call takes effect globally:

```go
slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug, // "debug" shows every outbound request URL
})))
```

Each log record includes a `pkg` attribute (`"gemini"`, `"openai"`, `"anthropic"`, `"provider"`, `"models"`) for filtering.

---

## Testing the CLI

The `cmd/providertron` binary lets you exercise every capability without writing application code:

```bash
go build -o ptron ./cmd/providertron

# Discover and cache models
./ptron -provider gemini -api-key $GEMINI_KEY -command discover

# Generate
./ptron -provider gemini -api-key $GEMINI_KEY -command generate -prompt "Say hello"

# Stream
./ptron -provider openai -api-key $OPENAI_KEY -command stream -prompt "Count to 5"

# Embed
./ptron -provider openai -api-key $OPENAI_KEY -command embed -prompt "Hello world"

# Batch (requires a JSON file)
echo '[{"id":"1","input":"foo"},{"id":"2","input":"bar"}]' > items.json
./ptron -provider openai -api-key $OPENAI_KEY -command batch -input items.json

# TTS
./ptron -provider gemini -api-key $GEMINI_KEY -command voice -voice-dir tts \
    -prompt "Hello from providertron" -audio-out out.wav

# Confirm unavailable capability error
./ptron -provider anthropic -api-key $ANTHROPIC_KEY -command voice -voice-dir tts -prompt "test"
# → time=... level=ERROR msg="command failed" command=voice
#   err="provider(anthropic).Voice: anthropic.Voice: Anthropic does not provide a voice API; ... capability unavailable for this provider"

# JSON dispatch
echo '{"type":"generate","messages":[{"role":"user","content":"ping"}]}' | \
    ./ptron -provider gemini -api-key $GEMINI_KEY -command dispatch -prompt "$(cat)"

# Raw JSON output
./ptron -provider gemini -api-key $GEMINI_KEY -command generate -prompt "ping" -output-json

# Debug logging
./ptron -provider gemini -api-key $GEMINI_KEY -command generate -prompt "ping" -log-level debug
```
