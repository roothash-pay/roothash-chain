package apis

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/cpchain-network/cp-chain/cp-service/eth"
	"github.com/cpchain-network/cp-chain/cp-supervisor/supervisor/types"
)

type SupervisorAPI interface {
	SupervisorAdminAPI
	SupervisorQueryAPI
}

type SupervisorAdminAPI interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	AddL2RPC(ctx context.Context, rpc string, jwtSecret eth.Bytes32) error
}

type SupervisorQueryAPI interface {
	CheckAccessList(ctx context.Context, inboxEntries []common.Hash,
		minSafety types.SafetyLevel, executingDescriptor types.ExecutingDescriptor) error
	CrossDerivedToSource(ctx context.Context, chainID eth.ChainID, derived eth.BlockID) (derivedFrom eth.BlockRef, err error)
	LocalUnsafe(ctx context.Context, chainID eth.ChainID) (eth.BlockID, error)
	CrossSafe(ctx context.Context, chainID eth.ChainID) (types.DerivedIDPair, error)
	Finalized(ctx context.Context, chainID eth.ChainID) (eth.BlockID, error)
	FinalizedL1(ctx context.Context) (eth.BlockRef, error)
	SuperRootAtTimestamp(ctx context.Context, timestamp hexutil.Uint64) (eth.SuperRootResponse, error)
	SyncStatus(ctx context.Context) (eth.SupervisorSyncStatus, error)
	AllSafeDerivedAt(ctx context.Context, derivedFrom eth.BlockID) (derived map[eth.ChainID]eth.BlockID, err error)
}
