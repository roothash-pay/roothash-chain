package pipeline

import (
	"fmt"

	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/common"

	"github.com/roothash-pay/roothash-chain/rhs-deployer/pkg/deployer/state"
)

// PrefundL2DevGenesis pre-funds accounts in the core dev genesis for testing purposes
func PrefundL2DevGenesis(env *Env, intent *state.Intent, st *state.State, chainID common.Hash) error {
	lgr := env.Logger.New("stage", "prefund-l2-dev-genesis")
	lgr.Info("Prefunding accounts in core dev genesis")

	thisIntent, err := intent.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain intent: %w", err)
	}

	thisChainState, err := st.Chain(chainID)
	if err != nil {
		return fmt.Errorf("failed to get chain state: %w", err)
	}

	if thisIntent.L2DevGenesisParams == nil {
		lgr.Warn("No core dev params, will not prefund any accounts")
		return nil
	}
	prefundMap := thisIntent.L2DevGenesisParams.Prefund
	if len(prefundMap) == 0 {
		lgr.Warn("Not prefunding any core dev accounts. core dev genesis may not be usable.")
		return nil
	}

	for addr, amount := range prefundMap {
		acc := thisChainState.Allocs.Data.Accounts[addr]
		acc.Balance = (*uint256.Int)(amount).ToBig()
		thisChainState.Allocs.Data.Accounts[addr] = acc
	}
	lgr.Info("Prefunded dev accounts on core", "accounts", len(prefundMap))
	return nil
}
