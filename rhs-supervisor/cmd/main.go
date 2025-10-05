package main

import (
	"context"
	"os"

	"github.com/roothash-pay/roothash-chain/rhs-supervisor/config"
	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/log"

	opservice "github.com/roothash-pay/roothash-chain/rhs-service"
	"github.com/roothash-pay/roothash-chain/rhs-service/cliapp"
	"github.com/roothash-pay/roothash-chain/rhs-service/ctxinterrupt"
	oplog "github.com/roothash-pay/roothash-chain/rhs-service/log"
	"github.com/roothash-pay/roothash-chain/rhs-service/metrics/doc"
	"github.com/roothash-pay/roothash-chain/rhs-supervisor/flags"
	"github.com/roothash-pay/roothash-chain/rhs-supervisor/metrics"
	"github.com/roothash-pay/roothash-chain/rhs-supervisor/supervisor"
)

var (
	Version   = "v0.0.0"
	GitCommit = ""
	GitDate   = ""
)

func main() {
	ctx := ctxinterrupt.WithSignalWaiterMain(context.Background())
	err := run(ctx, os.Args, fromConfig)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func run(ctx context.Context, args []string, fn supervisor.MainFn) error {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = cliapp.ProtectFlags(flags.Flags)
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "rhs-supervisor"
	app.Usage = "rhs-supervisor monitors cross-core interop messaging"
	app.Description = "The rhs-supervisor monitors cross-core interop messaging by pre-fetching events and then resolving the cross-core dependencies to answer safety queries."
	app.Action = cliapp.LifecycleCmd(supervisor.Main(app.Version, fn))
	app.Commands = []*cli.Command{
		{
			Name:        "doc",
			Subcommands: doc.NewSubcommands(metrics.NewMetrics("default")),
		},
	}
	return app.RunContext(ctx, args)
}

func fromConfig(ctx context.Context, cfg *config.Config, logger log.Logger) (cliapp.Lifecycle, error) {
	return supervisor.SupervisorFromConfig(ctx, cfg, logger)
}
