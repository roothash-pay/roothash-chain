package inspect

import (
	"fmt"

	"github.com/cpchain-network/cp-chain/common/genesis"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/pipeline"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/state"
	"github.com/cpchain-network/cp-chain/cp-service/ioutil"
	"github.com/cpchain-network/cp-chain/cp-service/jsonutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
)

func DeployConfigCLI(cliCtx *cli.Context) error {
	cliCfg, err := readConfig(cliCtx)
	if err != nil {
		return err
	}

	globalState, err := pipeline.ReadState(cliCfg.Workdir)
	if err != nil {
		return fmt.Errorf("failed to read globalState: %w", err)
	}

	config, err := DeployConfig(globalState, cliCfg.ChainID)
	if err != nil {
		return fmt.Errorf("failed to generate deploy config: %w", err)
	}

	if err := jsonutil.WriteJSON(config, ioutil.ToStdOutOrFileOrNoop(cliCfg.Outfile, 0o666)); err != nil {
		return fmt.Errorf("failed to write deploy config: %w", err)
	}

	return nil
}

func DeployConfig(globalState *state.State, chainID common.Hash) (*genesis.DeployConfig, error) {
	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chain state: %w", err)
	}

	intent := globalState.AppliedIntent
	if intent == nil {
		return nil, fmt.Errorf("can only run this command following a full apply")
	}
	chainIntent, err := intent.Chain(chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to find chain intent: %w", err)
	}

	config, err := state.CombineDeployConfig(intent, chainIntent, globalState, chainState)
	if err != nil {
		return nil, fmt.Errorf("failed to generate deploy config: %w", err)
	}

	return &config, nil
}
