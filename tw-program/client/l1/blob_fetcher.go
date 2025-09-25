package l1

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/theweb3-chain/tw-node/rollup/derive"
	"github.com/roothash-pay/theweb3-chain/tw-service/eth"
)

var InvalidHashesLengthError = errors.New("invalid hashes length")

type BlobFetcher struct {
	logger log.Logger
	oracle Oracle
}

var _ derive.L1BlobsFetcher = (*BlobFetcher)(nil)

func NewBlobFetcher(logger log.Logger, oracle Oracle) *BlobFetcher {
	return &BlobFetcher{
		logger: logger,
		oracle: oracle,
	}
}

// GetBlobs fetches blobs that were confirmed in the given L1 block with the given indexed blob hashes.
func (b *BlobFetcher) GetBlobs(ctx context.Context, ref eth.L1BlockRef, hashes []eth.IndexedBlobHash) ([]*eth.Blob, error) {
	blobs := make([]*eth.Blob, len(hashes))
	for i := 0; i < len(hashes); i++ {
		b.logger.Info("Fetching blob", "l1_ref", ref.Hash, "blob_versioned_hash", hashes[i].Hash, "index", hashes[i].Index)
		blobs[i] = b.oracle.GetBlob(ref, hashes[i])
	}
	return blobs, nil
}
