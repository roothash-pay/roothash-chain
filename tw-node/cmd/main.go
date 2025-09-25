package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/log"

	opnode "github.com/roothash-pay/theweb3-chain/tw-node"
	"github.com/roothash-pay/theweb3-chain/tw-node/chaincfg"
	"github.com/roothash-pay/theweb3-chain/tw-node/cmd/genesis"
	"github.com/roothash-pay/theweb3-chain/tw-node/cmd/interop"
	"github.com/roothash-pay/theweb3-chain/tw-node/cmd/networks"
	"github.com/roothash-pay/theweb3-chain/tw-node/cmd/p2p"
	"github.com/roothash-pay/theweb3-chain/tw-node/flags"
	"github.com/roothash-pay/theweb3-chain/tw-node/metrics"
	"github.com/roothash-pay/theweb3-chain/tw-node/node"
	"github.com/roothash-pay/theweb3-chain/tw-node/version"
	opservice "github.com/roothash-pay/theweb3-chain/tw-service"
	"github.com/roothash-pay/theweb3-chain/tw-service/cliapp"
	"github.com/roothash-pay/theweb3-chain/tw-service/ctxinterrupt"
	oplog "github.com/roothash-pay/theweb3-chain/tw-service/log"
	"github.com/roothash-pay/theweb3-chain/tw-service/metrics/doc"
)

var (
	GitCommit = ""
	GitDate   = ""
)

// VersionWithMeta holds the textual version string including the metadata.
var VersionWithMeta = opservice.FormatVersion(version.Version, GitCommit, GitDate, version.Meta)

func main() {
	// Set up logger with a default INFO level in case we fail to parse flags,
	// otherwise the final critical log won't show what the parsing error was.
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Version = VersionWithMeta
	app.Flags = cliapp.ProtectFlags(flags.Flags)
	app.Name = "tw-node"
	app.Usage = "Optimism Rollup Node"
	app.Description = "The Optimism Rollup Node derives core block inputs from L1 data and drives an external core Execution Engine to build a core chain."
	app.Action = cliapp.LifecycleCmd(RollupNodeMain)
	app.Commands = []*cli.Command{
		{
			Name:        "p2p",
			Subcommands: p2p.Subcommands,
		},
		{
			Name:        "genesis",
			Subcommands: genesis.Subcommands,
		},
		{
			Name:        "doc",
			Subcommands: doc.NewSubcommands(metrics.NewMetrics("default")),
		},
		{
			Name:        "networks",
			Subcommands: networks.Subcommands,
		},
		interop.InteropCmd,
	}

	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	err := app.RunContext(ctx, os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func RollupNodeMain(ctx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	logCfg := oplog.ReadCLIConfig(ctx)
	log := oplog.NewLogger(oplog.AppOut(ctx), logCfg)
	oplog.SetGlobalLogHandler(log.Handler())
	opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, log)
	opservice.WarnOnDeprecatedFlags(ctx, flags.DeprecatedFlags, log)
	m := metrics.NewMetrics("default")

	cfg, err := opnode.NewConfig(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("unable to create the rollup node config: %w", err)
	}
	cfg.Cancel = closeApp

	// Only pretty-print the banner if it is a terminal log. Other log it as key-value pairs.
	if logCfg.Format == "terminal" {
		log.Info("rollup config:\n" + cfg.Rollup.Description(chaincfg.L2ChainIDToNetworkDisplayName))
	} else {
		cfg.Rollup.LogDescription(log, chaincfg.L2ChainIDToNetworkDisplayName)
	}

	n, err := node.New(ctx.Context, cfg, log, VersionWithMeta, m)
	if err != nil {
		return nil, fmt.Errorf("unable to create the rollup node: %w", err)
	}

	return n, nil
}
