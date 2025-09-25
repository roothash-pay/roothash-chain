package engine

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/cp-chain/cp-node/rollup"
	"github.com/cpchain-network/cp-chain/cp-node/rollup/event"
	"github.com/cpchain-network/cp-chain/cp-node/rollup/sync"
)

// ResetEngineRequestEvent requests the EngineResetDeriver to walk
// the core chain backwards until it finds a plausible unsafe head,
// and find an core safe block that is guaranteed to still be from the L1 chain.
// This event is not used in interop.
type ResetEngineRequestEvent struct{}

func (ev ResetEngineRequestEvent) String() string {
	return "reset-engine-request"
}

type EngineResetDeriver struct {
	ctx     context.Context
	log     log.Logger
	cfg     *rollup.Config
	l2      sync.L2Chain
	syncCfg *sync.Config

	emitter event.Emitter
}

func NewEngineResetDeriver(ctx context.Context, log log.Logger, cfg *rollup.Config,
	l2 sync.L2Chain, syncCfg *sync.Config) *EngineResetDeriver {
	return &EngineResetDeriver{
		ctx:     ctx,
		log:     log,
		cfg:     cfg,
		l2:      l2,
		syncCfg: syncCfg,
	}
}

func (d *EngineResetDeriver) AttachEmitter(em event.Emitter) {
	d.emitter = em
}

func (d *EngineResetDeriver) OnEvent(ev event.Event) bool {
	switch ev.(type) {
	case ResetEngineRequestEvent:
		result, err := sync.FindL2Heads(d.ctx, d.cfg, d.l2, d.log, d.syncCfg)
		if err != nil {
			d.emitter.Emit(rollup.ResetEvent{Err: fmt.Errorf("failed to find the core Heads to start from: %w", err)})
			return true
		}
		d.emitter.Emit(rollup.ForceResetEvent{
			LocalUnsafe: result.Unsafe,
			CrossUnsafe: result.Unsafe,
			LocalSafe:   result.Safe,
			CrossSafe:   result.Safe,
			Finalized:   result.Finalized,
		})
	default:
		return false
	}
	return true
}
