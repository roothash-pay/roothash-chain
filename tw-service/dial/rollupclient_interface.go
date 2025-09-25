package dial

import (
	"context"

	"github.com/cpchain-network/cp-chain/cp-node/rollup"
	"github.com/cpchain-network/cp-chain/cp-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

// RollupClientInterface is an interfaces for providing a RollupClient
// It does not describe all of the functions a RollupClient has, only the ones used by the core Providers and their callers
type RollupClientInterface interface {
	SyncStatusProvider
	OutputAtBlock(ctx context.Context, blockNum uint64) (*eth.OutputResponse, error)
	RollupConfig(ctx context.Context) (*rollup.Config, error)
	StartSequencer(ctx context.Context, unsafeHead common.Hash) error
	SequencerActive(ctx context.Context) (bool, error)
	Close()
}

// SyncStatusProvider is the interfaces of a rollup client from which its sync status
// can be queried.
type SyncStatusProvider interface {
	SyncStatus(ctx context.Context) (*eth.SyncStatus, error)
}
