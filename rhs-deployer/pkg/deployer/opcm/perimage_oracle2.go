package opcm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/roothash-pay/roothash-chain/common/script"
)

type DeployPreimageOracle2Input struct {
	MinProposalSize *big.Int
	ChallengePeriod *big.Int
}

type DeployPreimageOracle2Output struct {
	PreimageOracle common.Address
}

type DeployPreimageOracleScript script.DeployScriptWithOutput[DeployPreimageOracle2Input, DeployPreimageOracle2Output]

// NewDeployPreimageOracleScript loads and validates the DeployPreimageOracle2 script contract
func NewDeployPreimageOracleScript(host *script.Host) (DeployPreimageOracleScript, error) {
	return script.NewDeployScriptWithOutputFromFile[DeployPreimageOracle2Input, DeployPreimageOracle2Output](host, "DeployPreimageOracle2.s.sol", "DeployPreimageOracle2")
}
