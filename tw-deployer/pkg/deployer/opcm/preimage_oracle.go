package opcm

import (
	"math/big"

	"github.com/cpchain-network/cp-chain/common/script"
	"github.com/ethereum/go-ethereum/common"
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
