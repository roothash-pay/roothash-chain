package engine

import "github.com/roothash-pay/theweb3-chain/tw-service/eth"

type PayloadInvalidEvent struct {
	Envelope *eth.ExecutionPayloadEnvelope
	Err      error
}

func (ev PayloadInvalidEvent) String() string {
	return "payload-invalid"
}

func (eq *EngDeriver) onPayloadInvalid(ev PayloadInvalidEvent) {
	eq.log.Warn("Payload was invalid", "block", ev.Envelope.ExecutionPayload.ID(),
		"err", ev.Err, "timestamp", uint64(ev.Envelope.ExecutionPayload.Timestamp))
}
