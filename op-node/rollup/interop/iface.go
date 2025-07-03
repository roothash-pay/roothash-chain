package interop

import (
	"context"

	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/cp-chain/op-node/rollup"
	"github.com/cpchain-network/cp-chain/op-node/rollup/event"
	"github.com/cpchain-network/cp-chain/op-node/rollup/interop/managed"
	"github.com/cpchain-network/cp-chain/op-node/rollup/interop/standard"
	opmetrics "github.com/cpchain-network/cp-chain/op-service/metrics"
)

type SubSystem interface {
	event.Deriver
	event.AttachEmitter
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

var _ SubSystem = (*managed.ManagedMode)(nil)
var _ SubSystem = (*standard.StandardMode)(nil)

type L1Source interface {
	managed.L1Source
}

type L2Source interface {
	managed.L2Source
}

type Setup interface {
	Setup(ctx context.Context, logger log.Logger, rollupCfg *rollup.Config, l2 L2Source, m opmetrics.RPCMetricer) (SubSystem, error)
	Check() error
}
