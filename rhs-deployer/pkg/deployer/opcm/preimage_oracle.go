package opcm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/roothash-pay/roothash-chain/common/script"
)

type DeployPreimageOracleInput struct {
	MinProposalSize *big.Int
	ChallengePeriod *big.Int
}

type DeployPreimageOracleOutput struct {
	PreimageOracle common.Address
}

func DeployPreimageOracle(
	host *script.Host,
	input DeployPreimageOracleInput,
) (DeployPreimageOracleOutput, error) {
	return RunScriptSingle[DeployPreimageOracleInput, DeployPreimageOracleOutput](
		host,
		input,
		"DeployPreimageOracle.s.sol",
		"DeployPreimageOracle",
	)
}
