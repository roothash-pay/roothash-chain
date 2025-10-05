package engine

import (
	"github.com/roothash-pay/roothash-chain/rhs-node/rollup/derive"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

// EngineState provides a read-only interfaces of the forkchoice state properties of the core Engine.
type EngineState interface {
	Finalized() eth.L2BlockRef
	UnsafeL2Head() eth.L2BlockRef
	SafeL2Head() eth.L2BlockRef
}

type Engine interface {
	ExecEngine
	derive.L2Source
}

type LocalEngineState interface {
	EngineState

	PendingSafeL2Head() eth.L2BlockRef
	BackupUnsafeL2Head() eth.L2BlockRef
}

type LocalEngineControl interface {
	LocalEngineState
	ResetEngineControl
}

var _ LocalEngineControl = (*EngineController)(nil)
