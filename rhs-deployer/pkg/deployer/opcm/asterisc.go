package opcm

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/roothash-pay/roothash-chain/common/script"
)

type DeployAsteriscInput struct {
	PreimageOracle common.Address
}

func (input *DeployAsteriscInput) InputSet() bool {
	return true
}

type DeployAsteriscOutput struct {
	AsteriscSingleton common.Address
}

func (output *DeployAsteriscOutput) CheckOutput(input common.Address) error {
	return nil
}

func DeployAsterisc(
	host *script.Host,
	input DeployAsteriscInput,
) (DeployAsteriscOutput, error) {
	return RunScriptSingle[DeployAsteriscInput, DeployAsteriscOutput](host, input, "DeployAsterisc.s.sol", "DeployAsterisc")
}
