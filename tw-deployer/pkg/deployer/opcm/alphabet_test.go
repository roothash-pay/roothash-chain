package opcm

import (
	"testing"

	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/broadcaster"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/deployer/testutil"
	"github.com/cpchain-network/cp-chain/cp-deployer/pkg/env"
	"github.com/cpchain-network/cp-chain/cp-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

func TestDeployAlphabetVM(t *testing.T) {
	t.Parallel()

	_, artifacts := testutil.LocalArtifacts(t)

	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		testlog.Logger(t, log.LevelInfo),
		common.Address{'D'},
		artifacts,
	)
	require.NoError(t, err)

	input := DeployAlphabetVMInput{
		AbsolutePrestate: common.Hash{'A'},
		PreimageOracle:   common.Address{'O'},
	}

	output, err := DeployAlphabetVM(host, input)
	require.NoError(t, err)

	require.NotEmpty(t, output.AlphabetVM)
}
