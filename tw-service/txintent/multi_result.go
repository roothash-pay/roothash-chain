package txintent

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/roothash-pay/theweb3-chain/tw-service/eth"
)

var _ Result = (*MulticallOutput)(nil)

type MulticallOutput struct {
	receipt    *types.Receipt
	includedIn eth.BlockRef
	chainID    eth.ChainID
}

func (m *MulticallOutput) Init() Result {
	return &MulticallOutput{}
}

// FromReceipt stores all gained info
func (m *MulticallOutput) FromReceipt(ctx context.Context, rec *types.Receipt, includedIn eth.BlockRef, chainID eth.ChainID) error {
	m.receipt = rec
	m.includedIn = includedIn
	m.chainID = chainID
	return nil
}
