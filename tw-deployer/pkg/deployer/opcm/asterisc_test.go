package opcm

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/deployer/broadcaster"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/deployer/testutil"
	"github.com/roothash-pay/theweb3-chain/tw-deployer/pkg/env"
	"github.com/roothash-pay/theweb3-chain/tw-service/testlog"
	"github.com/stretchr/testify/require"
)

func TestDeployAsterisc(t *testing.T) {
	t.Parallel()

	_, artifacts := testutil.LocalArtifacts(t)

	host, err := env.DefaultScriptHost(
		broadcaster.NoopBroadcaster(),
		testlog.Logger(t, log.LevelInfo),
		common.Address{'D'},
		artifacts,
	)
	require.NoError(t, err)

	input := DeployAsteriscInput{
		PreimageOracle: common.Address{0xab},
	}

	output, err := DeployAsterisc(host, input)
	require.NoError(t, err)

	require.NotEmpty(t, output.AsteriscSingleton)
}
