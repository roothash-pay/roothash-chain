package genesis

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/roothash-pay/roothash-chain/rhs-service/sources/batching"
	"github.com/roothash-pay/roothash-chain/rhs-service/sources/batching/rpcblock"
)

var (
	methodStartBlock = "startBlock"
)

type SystemConfigContract struct {
	caller   *batching.MultiCaller
	contract *batching.BoundContract
}

func NewSystemConfigContract(caller *batching.MultiCaller, addr common.Address) *SystemConfigContract {
	return &SystemConfigContract{
		caller:   caller,
		contract: nil,
	}
}

func (c *SystemConfigContract) StartBlock(ctx context.Context) (*big.Int, error) {
	result, err := c.caller.SingleCall(ctx, rpcblock.Latest, c.contract.Call(methodStartBlock))
	if err != nil {
		return nil, fmt.Errorf("failed to call startBlock: %w", err)
	}
	return result.GetBigInt(0), nil
}
