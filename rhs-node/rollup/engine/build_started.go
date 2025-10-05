package engine

import (
	"time"

	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

type BuildStartedEvent struct {
	Info eth.PayloadInfo

	BuildStarted time.Time

	Parent eth.L2BlockRef

	// if payload should be promoted to (local) safe (must also be pending safe, see DerivedFrom)
	Concluding bool
	// payload is promoted to pending-safe if non-zero
}

func (ev BuildStartedEvent) String() string {
	return "build-started"
}

func (eq *EngDeriver) onBuildStarted(ev BuildStartedEvent) {
	// If a (pending) safe block, immediately seal the block
	//eq.emitter.Emit(BuildSealEvent{
	//	Info:         ev.Info,
	//	BuildStarted: ev.BuildStarted,
	//	Concluding:   ev.Concluding,
	//})
}
