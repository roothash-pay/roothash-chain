package opcm_test

import (
	"math/big"
	"testing"

	"github.com/cpchain-network/cp-chain/common/foundry"
	"github.com/cpchain-network/cp-chain/common/script"
	"github.com/cpchain-network/cp-chain/cp-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// createTestHost is a helper function for testing deploy script wrappers
func createTestHost(t *testing.T) *script.Host {
	t.Helper()

	// Create a logger
	logger, _ := testlog.CaptureLogger(t, log.LevelInfo)

	// Create an artifact filesystem pointing to the bedrock contracts artifact directory
	af := foundry.OpenArtifactsDir("../../../../packages/contracts-cpchain/forge-artifacts")

	// Now put a host together
	host := script.NewHost(logger, af, nil, script.DefaultContext, script.WithCreate2Deployer())
	host.SetTxOrigin(common.BigToAddress(big.NewInt(6)))

	// And enable cheats
	err := host.EnableCheats()
	require.NoError(t, err)

	return host
}
