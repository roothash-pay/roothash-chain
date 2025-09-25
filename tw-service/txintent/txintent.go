package txintent

import (
	"context"

	"github.com/cpchain-network/cp-chain/cp-service/eth"
	"github.com/cpchain-network/cp-chain/cp-service/plan"
	"github.com/cpchain-network/cp-chain/cp-service/txplan"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type Call interface {
	To() (*common.Address, error)
	Data() ([]byte, error)
	AccessList() (types.AccessList, error)
}

type Result interface {
	FromReceipt(ctx context.Context, rec *types.Receipt, includedIn eth.BlockRef, chainID eth.ChainID) error
	Init() Result
}

type IntentTx[V Call, R Result] struct {
	PlannedTx *txplan.PlannedTx
	Content   plan.Lazy[V]
	Result    plan.Lazy[R]
}

func NewIntent[V Call, R Result](opts ...txplan.Option) *IntentTx[V, R] {
	v := &IntentTx[V, R]{
		PlannedTx: txplan.NewPlannedTx(opts...),
	}
	v.PlannedTx.To.DependOn(&v.Content)
	v.PlannedTx.To.Fn(func(ctx context.Context) (*common.Address, error) {
		return v.Content.Value().To()
	})
	v.PlannedTx.Data.DependOn(&v.Content)
	v.PlannedTx.Data.Fn(func(ctx context.Context) (hexutil.Bytes, error) {
		return v.Content.Value().Data()
	})
	v.PlannedTx.AccessList.DependOn(&v.Content)
	v.PlannedTx.AccessList.Fn(func(ctx context.Context) (types.AccessList, error) {
		return v.Content.Value().AccessList()
	})
	v.Result.DependOn(&v.PlannedTx.Included, &v.PlannedTx.IncludedBlock, &v.PlannedTx.ChainID)
	v.Result.Fn(func(ctx context.Context) (R, error) {
		r := (*new(R)).Init().(R)
		err := r.FromReceipt(ctx, v.PlannedTx.Included.Value(), v.PlannedTx.IncludedBlock.Value(), v.PlannedTx.ChainID.Value())
		return r, err
	})
	return v
}
