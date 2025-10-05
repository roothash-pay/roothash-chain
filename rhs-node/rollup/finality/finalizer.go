package finality

import (
	"context"
	"github.com/ethereum/go-ethereum/log"
	"sync"

	"github.com/roothash-pay/roothash-chain/rhs-node/rollup"
	"github.com/roothash-pay/roothash-chain/rhs-node/rollup/derive"
	"github.com/roothash-pay/roothash-chain/rhs-node/rollup/engine"
	"github.com/roothash-pay/roothash-chain/rhs-node/rollup/event"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

// defaultFinalityLookback defines the amount of L1<>core relations to track for finalization purposes, one per L1 block.
//
// When L1 finalizes blocks, it finalizes finalityLookback blocks behind the L1 head.
// Non-finality may take longer, but when it does finalize again, it is within this range of the L1 head.
// Thus we only need to retain the L1<>core derivation relation data of this many L1 blocks.
//
// In the event of older finalization signals, misconfiguration, or insufficient L1<>core derivation relation data,
// then we may miss the opportunity to finalize more core blocks.
// This does not cause any divergence, it just causes lagging finalization status.
//
// The beacon chain on mainnet has 32 slots per epoch,
// and new finalization events happen at most 4 epochs behind the head.
// And then we add 1 to make pruning easier by leaving room for a new item without pruning the 32*4.
const defaultFinalityLookback = 4*32 + 1

// finalityDelay is the number of L1 blocks to traverse before trying to finalize core blocks again.
// We do not want to do this too often, since it requires fetching a L1 block by number, so no cache data.
const finalityDelay = 64

// calcFinalityLookback calculates the default finality lookback based on DA challenge window if altDA
// mode is activated or L1 finality lookback.
func calcFinalityLookback(cfg *rollup.Config) uint64 {
	return defaultFinalityLookback
}

type FinalityData struct {
	// The last core block that was fully derived and inserted into the core engine while processing this L1 block.
	L2Block eth.L2BlockRef
}

type FinalizerEngine interface {
	Finalized() eth.L2BlockRef
	SetFinalizedHead(eth.L2BlockRef)
}

type FinalizerL1Interface interface {
	L1BlockRefByNumber(context.Context, uint64) (eth.L1BlockRef, error)
}

type Finalizer struct {
	mu sync.Mutex

	log log.Logger

	ctx context.Context

	cfg *rollup.Config

	emitter event.Emitter

	// finalizedL1 is the currently perceived finalized L1 block.
	// This may be ahead of the current traversed origin when syncing.
	finalizedL1 eth.L1BlockRef

	// lastFinalizedL2 maintains how far we finalized, so we don't have to emit re-attempts.
	lastFinalizedL2 eth.L2BlockRef

	// triedFinalizeAt tracks at which L1 block number we last tried to finalize during sync.
	triedFinalizeAt uint64

	// Tracks which core blocks where last derived from which L1 block. At most finalityLookback large.
	finalityData []FinalityData

	// Maximum amount of core blocks to store in finalityData.
	finalityLookback uint64
}

func NewFinalizer(ctx context.Context, log log.Logger, cfg *rollup.Config) *Finalizer {
	lookback := calcFinalityLookback(cfg)
	return &Finalizer{
		ctx:              ctx,
		cfg:              cfg,
		log:              log,
		finalizedL1:      eth.L1BlockRef{},
		triedFinalizeAt:  0,
		finalityData:     make([]FinalityData, 0, lookback),
		finalityLookback: lookback,
	}
}

func (fi *Finalizer) AttachEmitter(em event.Emitter) {
	fi.emitter = em
}

// FinalizedL1 identifies the L1 chain (incl.) that included and/or produced all the finalized core blocks.
// This may return a zeroed ID if no finalization signals have been seen yet.
func (fi *Finalizer) FinalizedL1() (out eth.L1BlockRef) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	out = fi.finalizedL1
	return
}

type FinalizeL1Event struct {
	FinalizedL1 eth.L1BlockRef
}

func (ev FinalizeL1Event) String() string {
	return "finalized-l1"
}

type TryFinalizeEvent struct{}

func (ev TryFinalizeEvent) String() string {
	return "try-finalize"
}

func (fi *Finalizer) OnEvent(ev event.Event) bool {
	switch x := ev.(type) {
	case FinalizeL1Event:
		fi.onL1Finalized()
	case engine.SafeDerivedEvent:
		fi.onDerivedSafeBlock(x.Safe)
	case derive.DeriverIdleEvent:
		fi.onDerivationIdle(x.Origin)
	case rollup.ResetEvent:
		fi.onReset()
	case TryFinalizeEvent:
		fi.tryFinalize()
	case engine.ForkchoiceUpdateEvent:
		fi.lastFinalizedL2 = x.FinalizedL2Head
	default:
		return false
	}
	return true
}

// onL1Finalized applies a L1 finality signal
func (fi *Finalizer) onL1Finalized() {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	// when the L1 change we can suggest to try to finalize, as the pre-condition for core finality has now changed
	fi.emitter.Emit(TryFinalizeEvent{})
}

// onDerivationIdle is called when the pipeline is exhausted of new data (i.e. no more core blocks to derive from).
//
// Since finality applies to all core blocks fully derived from the same block,
// it optimal to only check after the derivation from the L1 block has been exhausted.
//
// This will look at what has been buffered so far,
// sanity-check we are on the finalizing L1 chain,
// and finalize any core blocks that were fully derived from known finalized L1 blocks.
func (fi *Finalizer) onDerivationIdle(derivedFrom eth.L1BlockRef) {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	if fi.finalizedL1 == (eth.L1BlockRef{}) {
		return // if no L1 information is finalized yet, then skip this
	}
	// If we recently tried finalizing, then don't try again just yet, but traverse more of L1 first.
	if fi.triedFinalizeAt != 0 && derivedFrom.Number <= fi.triedFinalizeAt+finalityDelay {
		return
	}
	fi.log.Debug("processing L1 finality information", "l1_finalized", fi.finalizedL1, "derived_from", derivedFrom, "previous", fi.triedFinalizeAt)
	fi.triedFinalizeAt = derivedFrom.Number
	fi.emitter.Emit(TryFinalizeEvent{})
}

func (fi *Finalizer) tryFinalize() {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	// overwritten if we finalize
	finalizedL2 := fi.lastFinalizedL2 // may be zeroed if nothing was finalized since startup.
	// go through the latest inclusion data, and find the last core block that was derived from a finalized L1 block
	for _, fd := range fi.finalityData {
		if fd.L2Block.Number > finalizedL2.Number {
			finalizedL2 = fd.L2Block
			// keep iterating, there may be later core blocks that can also be finalized
		}
	}
	fi.emitter.Emit(engine.PromoteFinalizedEvent{Ref: finalizedL2})
}

// onDerivedSafeBlock buffers the L1 block the safe head was fully derived from,
// to finalize it once the derived-from L1 block, or a later L1 block, finalizes.
func (fi *Finalizer) onDerivedSafeBlock(l2Safe eth.L2BlockRef) {
	fi.mu.Lock()
	defer fi.mu.Unlock()

	// Stop registering blocks after interop.
	// Finality in interop is determined by the superchain backend,
	// i.e. the rhs-supervisor RPC identifies which core block may be finalized.
	if fi.cfg.IsInterop(l2Safe.Time) {
		return
	}

	// remember the last core block that we fully derived from the given finality data
	if len(fi.finalityData) == 0 {
		// prune finality data if necessary, before appending any data.
		if uint64(len(fi.finalityData)) >= fi.finalityLookback {
			fi.finalityData = append(fi.finalityData[:0], fi.finalityData[1:fi.finalityLookback]...)
		}
		// append entry for new L1 block
		fi.finalityData = append(fi.finalityData, FinalityData{
			L2Block: l2Safe,
		})
		last := &fi.finalityData[len(fi.finalityData)-1]
		fi.log.Debug("extended finality-data", "last_l2", last.L2Block)
	} else {
		// if it's a new core block that was derived from the same latest L1 block, then just update the entry
		last := &fi.finalityData[len(fi.finalityData)-1]
		if last.L2Block != l2Safe { // avoid logging if there are no changes
			last.L2Block = l2Safe
			fi.log.Debug("updated finality-data", "last_l2", last.L2Block)
		}
	}
}

// onReset clears the recent history of safe-core blocks used for finalization,
// to avoid finalizing any reorged-out core blocks.
func (fi *Finalizer) onReset() {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	fi.finalityData = fi.finalityData[:0]
	fi.triedFinalizeAt = 0
	// no need to reset finalizedL1, it's finalized after all
}
