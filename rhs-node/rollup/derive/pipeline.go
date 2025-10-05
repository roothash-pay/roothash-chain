package derive

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/roothash-pay/roothash-chain/rhs-node/rollup"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

var ErrEngineResetReq = errors.New("cannot continue derivation until Engine has been reset")

type Metrics interface {
	RecordL1Ref(name string, ref eth.L1BlockRef)
	RecordL2Ref(name string, ref eth.L2BlockRef)
	RecordChannelInputBytes(inputCompressedBytes int)
	RecordHeadChannelOpened()
	RecordChannelTimedOut()
	RecordFrame()
	RecordDerivedBatches(batchType string)
	SetDerivationIdle(idle bool)
	RecordPipelineReset()
}

type L1Fetcher interface {
	L1BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L1BlockRef, error)
	L1BlockRefByNumberFetcher
	L1BlockRefByHashFetcher
	L1ReceiptsFetcher
	L1TransactionFetcher
}

type ResettableStage interface {
	// Reset resets a pull stage. `base` refers to the L1 Block Reference to reset to, with corresponding configuration.
	Reset(ctx context.Context, baseCfg eth.SystemConfig) error
}

// A ChannelFlusher flushes all internal state related to the current channel and then
// calls FlushChannel on the stage it owns. Note that this is in contrast to Reset, which
// is called by the owning Pipeline in a loop over all stages.
type ChannelFlusher interface {
	FlushChannel()
}

type ForkTransformer interface {
	Transform(rollup.ForkName)
}

type L2Source interface {
	PayloadByHash(context.Context, common.Hash) (*eth.ExecutionPayloadEnvelope, error)
	PayloadByNumber(context.Context, uint64) (*eth.ExecutionPayloadEnvelope, error)
	L2BlockRefByLabel(ctx context.Context, label eth.BlockLabel) (eth.L2BlockRef, error)
	L2BlockRefByHash(ctx context.Context, l2Hash common.Hash) (eth.L2BlockRef, error)
	L2BlockRefByNumber(ctx context.Context, num uint64) (eth.L2BlockRef, error)
	SystemConfigL2Fetcher
}

type l1TraversalStage interface {
	NextBlockProvider
	ResettableStage
	AdvanceL1Block(ctx context.Context) error
}

// DerivationPipeline is updated with new L1 data, and the Step() function can be iterated on to generate attributes
type DerivationPipeline struct {
	log       log.Logger
	rollupCfg *rollup.Config

	l2 L2Source

	// Index of the stage that is currently being reset.
	// >= len(stages) if no additional resetting is required
	resetting int
	stages    []ResettableStage

	attrib *AttributesQueue

	// L1 block that the next returned attributes are derived from, i.e. at the core-end of the pipeline.
	origin         eth.L1BlockRef
	resetL2Safe    eth.L2BlockRef
	resetSysConfig eth.SystemConfig
	engineIsReset  bool

	metrics Metrics
}

// NewDerivationPipeline creates a DerivationPipeline, to turn L1 data into core block-inputs.
func NewDerivationPipeline(log log.Logger, rollupCfg *rollup.Config, l2Source L2Source, metrics Metrics, managedMode bool,
) *DerivationPipeline {
	attrBuilder := NewFetchingAttributesBuilder(rollupCfg, l2Source)
	attributesQueue := NewAttributesQueue(log, rollupCfg, attrBuilder)

	// Reset from ResetEngine then up from L1 Traversal. The stages do not talk to each other during
	// the ResetEngine, but after the ResetEngine, this is the order in which the stages could talk to each other.
	// Note: The ResetEngine is the only reset that can fail.
	stages := []ResettableStage{attributesQueue}

	return &DerivationPipeline{
		log:       log,
		rollupCfg: rollupCfg,
		resetting: 0,
		stages:    stages,
		metrics:   metrics,
		attrib:    attributesQueue,
		l2:        l2Source,
	}
}

// DerivationReady returns true if the derivation pipeline is ready to be used.
// When it's being reset its state is inconsistent, and should not be used externally.
func (dp *DerivationPipeline) DerivationReady() bool {
	return dp.engineIsReset && dp.resetting > 0
}

func (dp *DerivationPipeline) Reset() {
	dp.resetting = 0
	dp.resetSysConfig = eth.SystemConfig{}
	dp.resetL2Safe = eth.L2BlockRef{}
	dp.engineIsReset = false
}

func (dp *DerivationPipeline) DepositsOnlyAttributes(parent eth.BlockID, derivedFrom eth.L1BlockRef) (*AttributesWithParent, error) {
	return dp.attrib.DepositsOnlyAttributes(parent, derivedFrom)
}

// Origin is the L1 block of the inner-most stage of the derivation pipeline,
// i.e. the L1 chain up to and including this point included and/or produced all the safe core blocks.
func (dp *DerivationPipeline) Origin() eth.L1BlockRef {
	return dp.origin
}

// Step tries to progress the buffer.
// An EOF is returned if the pipeline is blocked by waiting for new L1 data.
// If ctx errors no error is returned, but the step may exit early in a state that can still be continued.
// Any other error is critical and the derivation pipeline should be reset.
// An error is expected when the underlying source closes.
// When Step returns nil, it should be called again, to continue the derivation process.
func (dp *DerivationPipeline) Step(ctx context.Context, pendingSafeHead eth.L2BlockRef) (outAttrib *AttributesWithParent, outErr error) {
	defer dp.metrics.RecordL1Ref("l1_derived", dp.Origin())

	dp.metrics.SetDerivationIdle(false)
	defer func() {
		if outErr == io.EOF || errors.Is(outErr, EngineELSyncing) {
			dp.metrics.SetDerivationIdle(true)
		}
	}()

	// if any stages need to be reset, do that first.
	if dp.resetting < len(dp.stages) {
		if !dp.engineIsReset {
			return nil, NewResetError(ErrEngineResetReq)
		}

		// After the Engine has been reset to ensure it is derived from the canonical L1 chain,
		// we still need to internally rewind the L1 traversal further,
		// so we can read all the core data necessary for constructing the next batches that come after the safe head.
		if pendingSafeHead != dp.resetL2Safe {
			if err := dp.initialReset(ctx, pendingSafeHead); err != nil {
				return nil, fmt.Errorf("failed initial reset work: %w", err)
			}
		}

		if err := dp.stages[dp.resetting].Reset(ctx, dp.resetSysConfig); err == io.EOF {
			dp.log.Debug("reset of stage completed", "stage", dp.resetting)
			dp.resetting += 1
			return nil, nil
		} else if err != nil {
			return nil, fmt.Errorf("stage %d failed resetting: %w", dp.resetting, err)
		} else {
			return nil, nil
		}
	}

	if attrib, err := dp.attrib.NextAttributes(ctx, pendingSafeHead); err == nil {
		return attrib, nil
	} else if err == io.EOF {
		// If every stage has returned io.EOF, try to advance the L1 Origin
		return nil, err
	} else if errors.Is(err, EngineELSyncing) {
		return nil, err
	} else {
		return nil, fmt.Errorf("derivation failed: %w", err)
	}
}

// initialReset does the initial reset work of finding the L1 point to rewind back to
func (dp *DerivationPipeline) initialReset(ctx context.Context, resetL2Safe eth.L2BlockRef) error {
	dp.log.Info("Rewinding derivation-pipeline L1 traversal to handle reset")

	dp.metrics.RecordPipelineReset()

	// Walk back core chain to find the L1 origin that is old enough to start buffering channel data from.
	pipelineL2 := resetL2Safe

	parent, err := dp.l2.L2BlockRefByHash(ctx, pipelineL2.ParentHash)
	if err != nil {
		return NewResetError(fmt.Errorf("failed to fetch core parent block %s", pipelineL2.ParentID()))
	}
	pipelineL2 = parent

	sysCfg, err := dp.l2.SystemConfigByL2Hash(ctx, pipelineL2.Hash)
	if err != nil {
		return NewTemporaryError(fmt.Errorf("failed to fetch L1 config of core block %s: %w", pipelineL2.ID(), err))
	}

	dp.resetSysConfig = sysCfg
	dp.resetL2Safe = resetL2Safe
	return nil
}

func (db *DerivationPipeline) transformStages(oldOrigin, newOrigin eth.L1BlockRef) {
	fork := db.rollupCfg.IsActivationBlock(oldOrigin.Time, newOrigin.Time)
	if fork == "" {
		return
	}

	db.log.Info("Transforming stages", "fork", fork)
	for _, stage := range db.stages {
		if tf, ok := stage.(ForkTransformer); ok {
			tf.Transform(fork)
		}
	}
}

func (dp *DerivationPipeline) ConfirmEngineReset() {
	dp.engineIsReset = true
}
