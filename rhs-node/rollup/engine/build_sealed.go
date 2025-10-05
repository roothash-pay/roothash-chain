package engine

import (
	"time"

	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

// BuildSealedEvent is emitted by the engine when a payload finished building,
// but is not locally inserted as canonical block yet
type BuildSealedEvent struct {
	// if payload should be promoted to (local) safe (must also be pending safe, see DerivedFrom)
	Concluding bool
	// payload is promoted to pending-safe if non-zero
	BuildStarted time.Time

	Info     eth.PayloadInfo
	Envelope *eth.ExecutionPayloadEnvelope
	Ref      eth.L2BlockRef
}

func (ev BuildSealedEvent) String() string {
	return "build-sealed"
}

func (eq *EngDeriver) onBuildSealed(ev BuildSealedEvent) {
	// If a (pending) safe block, immediately process the block
	eq.emitter.Emit(PayloadProcessEvent{
		Concluding:   ev.Concluding,
		Envelope:     ev.Envelope,
		Ref:          ev.Ref,
		BuildStarted: ev.BuildStarted,
	})

}
