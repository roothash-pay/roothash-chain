package client

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/boot"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/claim"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l1"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l2"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/tasks"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

func RunPreInteropProgram(
	logger log.Logger,
	bootInfo *boot.BootInfo,
	l1PreimageOracle *l1.CachingOracle,
	l2PreimageOracle *l2.CachingOracle,
	db l2.KeyValueStore,
	opts tasks.DerivationOptions,
) error {
	logger.Info("Program Bootstrapped", "bootInfo", bootInfo)
	result, err := tasks.RunDerivation(
		logger,
		bootInfo.RollupConfig,
		bootInfo.L2ChainConfig,
		bootInfo.L1Head,
		bootInfo.L2OutputRoot,
		bootInfo.L2ClaimBlockNumber,
		l1PreimageOracle,
		l2PreimageOracle,
		db,
		opts,
	)
	if err != nil {
		return err
	}
	return claim.ValidateClaim(logger, eth.Bytes32(bootInfo.L2Claim), result.OutputRoot)
}
