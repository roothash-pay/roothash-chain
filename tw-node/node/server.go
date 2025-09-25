package node

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/roothash-pay/theweb3-chain/tw-node/rollup"
	opmetrics "github.com/roothash-pay/theweb3-chain/tw-service/metrics"
	oprpc "github.com/roothash-pay/theweb3-chain/tw-service/rpc"
)

func newRPCServer(rpcCfg *RPCConfig, rollupCfg *rollup.Config, l2Client l2EthClient, dr driverClient,
	safeDB SafeDBReader, log log.Logger, metrics opmetrics.RPCMetricer, appVersion string) *oprpc.Server {
	server := oprpc.NewServer(rpcCfg.ListenAddr, rpcCfg.ListenPort, appVersion,
		oprpc.WithLogger(log),
		oprpc.WithCORSHosts([]string{"*"}), // CORS is not important on tw-node, but we used to do this on the old tw-node RPC server, so kept for compatibility.
		oprpc.WithRPCRecorder(metrics.NewRecorder("main")),
	)
	api := NewNodeAPI(rollupCfg, l2Client, dr, safeDB, log)
	server.AddAPI(rpc.API{
		Namespace: "optimism",
		Service:   api,
	})
	return server
}
