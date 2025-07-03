package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/cpchain-network/cp-chain/cp-service/eth"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/cpchain-network/cp-chain/cp-service/dial"
	oplog "github.com/cpchain-network/cp-chain/cp-service/log"
	opmetrics "github.com/cpchain-network/cp-chain/cp-service/metrics"
	"github.com/cpchain-network/cp-chain/cp-service/oppprof"
	oprpc "github.com/cpchain-network/cp-chain/cp-service/rpc"
	"github.com/cpchain-network/cp-chain/cp-service/testlog"
	"github.com/cpchain-network/cp-chain/cp-supervisor/config"
	"github.com/cpchain-network/cp-chain/cp-supervisor/supervisor/backend/depset"
	"github.com/cpchain-network/cp-chain/cp-supervisor/supervisor/types"
)

func TestSupervisorService(t *testing.T) {
	depSet, err := depset.NewStaticConfigDependencySet(make(map[eth.ChainID]*depset.StaticConfigDependency))
	require.NoError(t, err)

	cfg := &config.Config{
		Version: "",
		LogConfig: oplog.CLIConfig{
			Level:  log.LevelError,
			Color:  false,
			Format: oplog.FormatLogFmt,
		},
		MetricsConfig: opmetrics.CLIConfig{
			Enabled:    true,
			ListenAddr: "127.0.0.1",
			ListenPort: 0, // pick a port automatically
		},
		PprofConfig: oppprof.CLIConfig{
			ListenEnabled:   true,
			ListenAddr:      "127.0.0.1",
			ListenPort:      0, // pick a port automatically
			ProfileType:     "",
			ProfileDir:      "",
			ProfileFilename: "",
		},
		RPC: oprpc.CLIConfig{
			ListenAddr:  "127.0.0.1",
			ListenPort:  0, // pick a port automatically
			EnableAdmin: true,
		},
		DependencySetSource: depSet,
		MockRun:             true,
	}
	logger := testlog.Logger(t, log.LevelError)
	supervisor, err := SupervisorFromConfig(context.Background(), cfg, logger)
	require.NoError(t, err)
	require.NoError(t, supervisor.Start(context.Background()), "start service")
	// run some RPC tests against the service with the mock backend
	{
		endpoint := "http://" + supervisor.rpcServer.Endpoint()
		t.Logf("dialing %s", endpoint)
		cl, err := dial.DialRPCClientWithTimeout(context.Background(), time.Second*5, logger, endpoint)
		require.NoError(t, err)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		err = cl.CallContext(ctx, nil, "supervisor_checkAccessList",
			[]common.Hash{}, types.CrossUnsafe, types.ExecutingDescriptor{Timestamp: 1234568})
		cancel()
		require.NoError(t, err)
		cl.Close()
	}
	require.NoError(t, supervisor.Stop(context.Background()), "stop service")
}
