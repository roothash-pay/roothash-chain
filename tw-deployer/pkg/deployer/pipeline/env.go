package pipeline

import (
	"context"
	"fmt"
	"path"

	"github.com/cpchain-network/cp-chain/common/genesis"
	"github.com/cpchain-network/cp-chain/common/script"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/broadcaster"
	"github.com/cpchain-network/cp-chain/cp-node/rollup"

	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/state"

	"github.com/cpchain-network/cp-chain/common/foundry"

	"github.com/cpchain-network/cp-chain/cp-service/jsonutil"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

type Env struct {
	StateWriter  StateWriter
	L1ScriptHost *script.Host
	L1Client     *ethclient.Client
	Broadcaster  broadcaster.Broadcaster
	Deployer     common.Address
	Logger       log.Logger
}

type StateWriter interface {
	WriteState(st *state.State) error
}

type stateWriterFunc func(st *state.State) error

func (f stateWriterFunc) WriteState(st *state.State) error {
	return f(st)
}

func WorkdirStateWriter(workdir string) StateWriter {
	return stateWriterFunc(func(st *state.State) error {
		return WriteState(workdir, st)
	})
}

func NoopStateWriter() StateWriter {
	return stateWriterFunc(func(st *state.State) error {
		return nil
	})
}

func ReadIntent(workdir string) (*state.Intent, error) {
	intentPath := path.Join(workdir, "intent.toml")
	intent, err := jsonutil.LoadTOML[state.Intent](intentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read intent file: %w", err)
	}
	return intent, nil
}

func ReadState(workdir string) (*state.State, error) {
	statePath := path.Join(workdir, "state.json")
	st, err := jsonutil.LoadJSON[state.State](statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}
	return st, nil
}

func WriteState(workdir string, st *state.State) error {
	statePath := path.Join(workdir, "state.json")
	return st.WriteToFile(statePath)
}

type ArtifactsBundle struct {
	L1 foundry.StatDirFs
	L2 foundry.StatDirFs
}

type Stage func(ctx context.Context, env *Env, bundle ArtifactsBundle, intent *state.Intent, st *state.State) error

func RenderGenesisAndRollup(globalState *state.State, chainID common.Hash, useGlobalIntent *state.Intent) (*core.Genesis, *rollup.Config, error) {
	if useGlobalIntent == nil && globalState.AppliedIntent == nil {
		return nil, nil, fmt.Errorf("chain state is not applied - run cp-deployer apply")
	}

	globalIntent := useGlobalIntent
	if globalIntent == nil {
		globalIntent = globalState.AppliedIntent
	}

	chainIntent, err := globalIntent.Chain(chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get applied chain intent: %w", err)
	}

	chainState, err := globalState.Chain(chainID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get chain ID %s: %w", chainID.String(), err)
	}

	l2Allocs := chainState.Allocs.Data
	config, err := state.CombineDeployConfig(
		globalIntent,
		chainIntent,
		globalState,
		chainState,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to combine core init config: %w", err)
	}

	l2GenesisBuilt, err := genesis.BuildL2Genesis(&config, l2Allocs, chainState.StartBlock.ToBlockRef())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build core genesis: %w", err)
	}
	l2GenesisBlock := l2GenesisBuilt.ToBlock()

	rollupConfig, err := config.RollupConfig(
		chainState.StartBlock.ToBlockRef(),
		l2GenesisBlock.Hash(),
		l2GenesisBlock.Number().Uint64(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build rollup config: %w", err)
	}

	if err := rollupConfig.Check(); err != nil {
		return nil, nil, fmt.Errorf("generated rollup config does not pass validation: %w", err)
	}

	return l2GenesisBuilt, rollupConfig, nil
}
