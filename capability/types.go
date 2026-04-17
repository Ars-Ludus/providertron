package capability

import "errors"

type CapabilityType string

const (
	CapabilityGenerate CapabilityType = "generate"
	CapabilityStream   CapabilityType = "stream"
	CapabilityEmbed    CapabilityType = "embedding"
	CapabilityBatch    CapabilityType = "batch"
	CapabilityVoice    CapabilityType = "voice"
)

var ErrCapabilityUnavailable = errors.New("capability unavailable for this provider")
