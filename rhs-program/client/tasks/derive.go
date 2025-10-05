package tasks

import (
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/roothash-pay/roothash-chain/rhs-node/rollup"
	cldr "github.com/roothash-pay/roothash-chain/rhs-program/client/driver"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l1"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l2"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l2/engineapi"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/mpt"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

type L2Source interface {
	L2OutputRoot(uint64) (common.Hash, eth.Bytes32, error)
}

type DerivationResult struct {
	Head       eth.L2BlockRef
	BlockHash  common.Hash
	OutputRoot eth.Bytes32
}

type DerivationOptions struct {
	// StoreBlockData controls whether block data, including intermediate trie nodes from transactions and receipts
	// of the derived block should be stored in the l2.KeyValueStore.
	StoreBlockData bool
}

// RunDerivation executes the core state transition, given a minimal interfaces to retrieve data.
// Returns the L2BlockRef of the safe head reached and the output root at l2ClaimBlockNum or
// the final safe head when l1Head is reached if l2ClaimBlockNum is not reached.
// Derivation may stop prior to l1Head if the l2ClaimBlockNum has already been reached though
// this is not guaranteed.
func RunDerivation(
	logger log.Logger,
	cfg *rollup.Config,
	l2Cfg *params.ChainConfig,
	l1Head common.Hash,
	l2OutputRoot common.Hash,
	l2ClaimBlockNum uint64,
	l1Oracle l1.Oracle,
	l2Oracle l2.Oracle,
	db l2.KeyValueStore,
	options DerivationOptions) (DerivationResult, error) {
	l1Source := l1.NewOracleL1Client(logger, l1Oracle, l1Head)
	l1BlobsSource := l1.NewBlobFetcher(logger, l1Oracle)
	engineBackend, err := l2.NewOracleBackedL2Chain(logger, l2Oracle, l1Oracle, l2Cfg, l2OutputRoot, db)
	if err != nil {
		return DerivationResult{}, fmt.Errorf("failed to create oracle-backed core chain: %w", err)
	}
	l2Source := l2.NewOracleEngine(cfg, logger, engineBackend, l2Oracle.Hinter())

	logger.Info("Starting derivation", "chainID", cfg.L2ChainID)
	d := cldr.NewDriver(logger, cfg, l1Source, l1BlobsSource, l2Source, l2ClaimBlockNum)
	result, err := d.RunComplete()
	if err != nil {
		return DerivationResult{}, fmt.Errorf("failed to run program to completion: %w", err)
	}
	logger.Info("Derivation complete", "head", result)

	if options.StoreBlockData {
		if err := storeBlockData(result.Hash, db, engineBackend); err != nil {
			return DerivationResult{}, fmt.Errorf("failed to write trie nodes: %w", err)
		}
		logger.Info("Trie nodes written")
	}
	return loadOutputRoot(l2ClaimBlockNum, result, l2Source)
}

func loadOutputRoot(l2ClaimBlockNum uint64, head eth.L2BlockRef, src L2Source) (DerivationResult, error) {
	blockHash, outputRoot, err := src.L2OutputRoot(min(l2ClaimBlockNum, head.Number))
	if err != nil {
		return DerivationResult{}, fmt.Errorf("calculate core output root: %w", err)
	}
	return DerivationResult{
		Head:       head,
		BlockHash:  blockHash,
		OutputRoot: outputRoot,
	}, nil
}

func storeBlockData(derivedBlockHash common.Hash, db l2.KeyValueStore, backend engineapi.CachingEngineBackend) error {
	block := backend.GetBlockByHash(derivedBlockHash)
	if block == nil {
		return fmt.Errorf("%w: derived block %v is missing", ethereum.NotFound, derivedBlockHash)
	}
	headerRLP, err := rlp.EncodeToBytes(block.Header())
	if err != nil {
		return fmt.Errorf("failed to encode block header: %w", err)
	}
	var blockHashKey []byte
	if err := db.Put(blockHashKey[:], headerRLP); err != nil {
		return fmt.Errorf("failed to store block header: %w", err)
	}

	opaqueTxs, err := eth.EncodeTransactions(block.Transactions())
	if err != nil {
		return err
	}
	if err := storeTrieNodes(opaqueTxs, db); err != nil {
		return err
	}
	receipts := backend.GetReceiptsByBlockHash(block.Hash())
	if receipts == nil {
		return fmt.Errorf("%w: receipts for block %v are missing", ethereum.NotFound, block.Hash())
	}
	opaqueReceipts, err := eth.EncodeReceipts(receipts)
	if err != nil {
		return err
	}
	return storeTrieNodes(opaqueReceipts, db)
}

func storeTrieNodes(values []hexutil.Bytes, db l2.KeyValueStore) error {
	_, nodes := mpt.WriteTrie(values)
	for _, node := range nodes {
		var key []byte
		if err := db.Put(key[:], node); err != nil {
			return fmt.Errorf("failed to store node: %w", err)
		}
	}
	return nil
}
