package derive

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/cpchain-network/cp-chain/cp-node/rollup"
	"github.com/cpchain-network/cp-chain/cp-service/eth"
)

// L2BlockRefSource is a source for the generation of a L2BlockRef. E.g. a
// *types.Block is a L2BlockRefSource.
//
// L2BlockToBlockRef extracts L2BlockRef from a L2BlockRefSource. The first
// transaction of a source must be a Deposit transaction.
type L2BlockRefSource interface {
	Hash() common.Hash
	ParentHash() common.Hash
	NumberU64() uint64
	Time() uint64
	Transactions() types.Transactions
}

// L2BlockToBlockRef extracts the essential L2BlockRef information from an core
// block ref source, falling back to genesis information if necessary.
func L2BlockToBlockRef(rollupCfg *rollup.Config, block L2BlockRefSource) (eth.L2BlockRef, error) {
	hash, number := block.Hash(), block.NumberU64()

	var sequenceNumber uint64
	genesis := &rollupCfg.Genesis
	if number == genesis.L2.Number {
		if hash != genesis.L2.Hash {
			return eth.L2BlockRef{}, fmt.Errorf("expected core genesis hash to match core block at genesis block number %d: %s <> %s", genesis.L2.Number, hash, genesis.L2.Hash)
		}
		sequenceNumber = 0
	} else {
		txs := block.Transactions()
		if txs.Len() == 0 {
			return eth.L2BlockRef{}, fmt.Errorf("l2 block is missing L1 info deposit tx, block hash: %s", hash)
		}
		tx := txs[0]
		if tx.Type() != types.DepositTxType {
			return eth.L2BlockRef{}, fmt.Errorf("first payload tx has unexpected tx type: %d", tx.Type())
		}
		info, err := L1BlockInfoFromBytes(rollupCfg, block.Time(), tx.Data())
		if err != nil {
			return eth.L2BlockRef{}, fmt.Errorf("failed to parse L1 info deposit tx from core block: %w", err)
		}
		sequenceNumber = info.SequenceNumber
	}

	return eth.L2BlockRef{
		Hash:           hash,
		Number:         number,
		ParentHash:     block.ParentHash(),
		Time:           block.Time(),
		SequenceNumber: sequenceNumber,
	}, nil
}
