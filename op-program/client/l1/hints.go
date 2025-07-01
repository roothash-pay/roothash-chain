package l1

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	HintL1BlockHeader  = "l1-block-header"
	HintL1Transactions = "l1-transactions"
	HintL1Receipts     = "l1-receipts"
	HintL1Blob         = "l1-blob"
	HintL1Precompile   = "l1-precompile"
	HintL1PrecompileV2 = "l1-precompile-v2"
)

type BlockHeaderHint common.Hash

func (l BlockHeaderHint) Hint() string {
	return HintL1BlockHeader + " " + (common.Hash)(l).String()
}

type TransactionsHint common.Hash

func (l TransactionsHint) Hint() string {
	return HintL1Transactions + " " + (common.Hash)(l).String()
}

type ReceiptsHint common.Hash

func (l ReceiptsHint) Hint() string {
	return HintL1Receipts + " " + (common.Hash)(l).String()
}

type BlobHint []byte

func (l BlobHint) Hint() string {
	return HintL1Blob + " " + hexutil.Encode(l)
}

type PrecompileHint []byte

func (l PrecompileHint) Hint() string {
	return HintL1Precompile + " " + hexutil.Encode(l)
}

type PrecompileHintV2 []byte

func (l PrecompileHintV2) Hint() string {
	return HintL1PrecompileV2 + " " + hexutil.Encode(l)
}
