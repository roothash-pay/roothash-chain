package derive

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/cpchain-network/cp-chain/cp-node/rollup"
	"github.com/cpchain-network/cp-chain/cp-service/eth"
	"github.com/cpchain-network/cp-chain/cp-service/predeploys"
)

// L1ReceiptsFetcher fetches L1 header info and receipts for the payload attributes derivation (the info tx and deposits)
type L1ReceiptsFetcher interface {
	InfoByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, error)
	FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error)
}

type SystemConfigL2Fetcher interface {
	SystemConfigByL2Hash(ctx context.Context, hash common.Hash) (eth.SystemConfig, error)
}

// FetchingAttributesBuilder fetches inputs for the building of core payload attributes on the fly.
type FetchingAttributesBuilder struct {
	rollupCfg *rollup.Config
	l2        SystemConfigL2Fetcher
	// whether to skip the L1 origin timestamp check - only for testing purposes
	testSkipL1OriginCheck bool
}

func NewFetchingAttributesBuilder(rollupCfg *rollup.Config, l2 SystemConfigL2Fetcher) *FetchingAttributesBuilder {
	return &FetchingAttributesBuilder{
		rollupCfg: rollupCfg,
		l2:        l2,
	}
}

// TestSkipL1OriginCheck skips the L1 origin timestamp check for testing purposes.
// Must not be used in production!
func (ba *FetchingAttributesBuilder) TestSkipL1OriginCheck() {
	ba.testSkipL1OriginCheck = true
}

// PreparePayloadAttributes prepares a PayloadAttributes template that is ready to build a core block with deposits only, on top of the given l2Parent, with the given epoch as L1 origin.
// The template defaults to NoTxPool=true, and no sequencer transactions: the caller has to modify the template to add transactions,
// by setting NoTxPool=false as sequencer, or by appending batch transactions as verifier.
// The severity of the error is returned; a crit=false error means there was a temporary issue, like a failed RPC or time-out.
// A crit=true error means the input arguments are inconsistent or invalid.
func (ba *FetchingAttributesBuilder) PreparePayloadAttributes(ctx context.Context, l2Parent eth.L2BlockRef) (attrs *eth.PayloadAttributes, err error) {
	sysConfig, err := ba.l2.SystemConfigByL2Hash(ctx, l2Parent.Hash)
	if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("failed to retrieve core parent block: %w", err))
	}

	nextL2Time := l2Parent.Time + ba.rollupCfg.BlockTime
	// Sanity check the L1 origin was correctly selected to maintain the time invariant between L1 and core

	var upgradeTxs []hexutil.Bytes
	if ba.rollupCfg.IsEcotoneActivationBlock(nextL2Time) {
		upgradeTxs, err = EcotoneNetworkUpgradeTransactions()
		if err != nil {
			return nil, NewCriticalError(fmt.Errorf("failed to build ecotone network upgrade txs: %w", err))
		}
	}

	if ba.rollupCfg.IsFjordActivationBlock(nextL2Time) {
		fjord, err := FjordNetworkUpgradeTransactions()
		if err != nil {
			return nil, NewCriticalError(fmt.Errorf("failed to build fjord network upgrade txs: %w", err))
		}
		upgradeTxs = append(upgradeTxs, fjord...)
	}

	if ba.rollupCfg.IsIsthmusActivationBlock(nextL2Time) {
		isthmus, err := IsthmusNetworkUpgradeTransactions()
		if err != nil {
			return nil, NewCriticalError(fmt.Errorf("failed to build isthmus network upgrade txs: %w", err))
		}
		upgradeTxs = append(upgradeTxs, isthmus...)
	}

	var afterForceIncludeTxs []hexutil.Bytes

	txs := make([]hexutil.Bytes, 0, 1+len(afterForceIncludeTxs)+len(upgradeTxs))
	txs = append(txs, afterForceIncludeTxs...)
	txs = append(txs, upgradeTxs...)

	var withdrawals *types.Withdrawals
	if ba.rollupCfg.IsCanyon(nextL2Time) {
		withdrawals = &types.Withdrawals{}
	}

	if nextL2Time < uint64(time.Now().Unix())-60 {
		nextL2Time = uint64(time.Now().Unix())
	}

	r := &eth.PayloadAttributes{
		Timestamp:             hexutil.Uint64(nextL2Time),
		SuggestedFeeRecipient: predeploys.SequencerFeeVaultAddr,
		Transactions:          txs,
		NoTxPool:              true,
		GasLimit:              (*eth.Uint64Quantity)(&sysConfig.GasLimit),
		Withdrawals:           withdrawals,
	}
	if ba.rollupCfg.IsHolocene(nextL2Time) {
		r.EIP1559Params = new(eth.Bytes8)
		*r.EIP1559Params = sysConfig.EIP1559Params
	}

	return r, nil
}
