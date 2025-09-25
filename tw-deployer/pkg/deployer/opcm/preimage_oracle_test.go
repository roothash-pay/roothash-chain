package opcm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/deployer/broadcaster"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/deployer/testutil"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/env"
	"github.com/roothash-pay/theweb3-chain/tw-service/testlog"
	"github.com/stretchr/testify/require"
)

func TestDeployPreimageOracle(t *testing.T) {
	t.Parallel()

	_, artifacts := testutil.LocalArtifacts(t)

	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		testlog.Logger(t, log.LevelInfo),
		common.Address{'D'},
		artifacts,
	)
	require.NoError(t, err)

	input := DeployPreimageOracleInput{
		MinProposalSize: big.NewInt(123),
		ChallengePeriod: big.NewInt(456),
	}

	output, err := DeployPreimageOracle(host, input)
	require.NoError(t, err)

	require.NotEmpty(t, output.PreimageOracle)
}
