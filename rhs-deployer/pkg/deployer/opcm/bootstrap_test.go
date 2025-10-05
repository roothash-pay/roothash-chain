package opcm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/roothash-chain/common/foundry"
	"github.com/roothash-pay/roothash-chain/common/script"
	"github.com/roothash-pay/roothash-chain/rhs-service/testlog"
	"github.com/stretchr/testify/require"
)

// createTestHost is a helper function for testing deploy script wrappers
func createTestHost(t *testing.T) *script.Host {
	t.Helper()

	// Create a logger
	logger, _ := testlog.CaptureLogger(t, log.LevelInfo)

	// Create an artifact filesystem pointing to the bedrock contracts artifact directory
	af := foundry.OpenArtifactsDir("../../../../packages/contracts-theweb3Chain/forge-artifacts")

	// Now put a host together
	host := script.NewHost(logger, af, nil, script.DefaultContext, script.WithCreate2Deployer())
	host.SetTxOrigin(common.BigToAddress(big.NewInt(6)))

	// And enable cheats
	err := host.EnableCheats()
	require.NoError(t, err)

	return host
}
