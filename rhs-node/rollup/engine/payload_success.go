package engine

import (
	"time"

	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

type PayloadSuccessEvent struct {
	// if payload should be promoted to (local) safe (must also be pending safe, see DerivedFrom)
	Concluding bool
	// payload is promoted to pending-safe if non-zero
	BuildStarted  time.Time
	InsertStarted time.Time

	Envelope *eth.ExecutionPayloadEnvelope
	Ref      eth.L2BlockRef
}

func (ev PayloadSuccessEvent) String() string {
	return "payload-success"
}

func (eq *EngDeriver) onPayloadSuccess(ev PayloadSuccessEvent) {
	eq.emitter.Emit(PromoteUnsafeEvent{Ref: ev.Ref})

	eq.emitter.Emit(PromotePendingSafeEvent{
		Ref:        ev.Ref,
		Concluding: ev.Concluding,
	})

	eq.emitter.Emit(TryUpdateEngineEvent{
		BuildStarted:  ev.BuildStarted,
		InsertStarted: ev.InsertStarted,
		Envelope:      ev.Envelope,
	})
}
