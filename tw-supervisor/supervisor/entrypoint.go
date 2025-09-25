package supervisor

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/log"

	opservice "github.com/cpchain-network/cp-chain/cp-service"
	"github.com/cpchain-network/cp-chain/cp-service/cliapp"
	oplog "github.com/cpchain-network/cp-chain/cp-service/log"
	"github.com/cpchain-network/cp-chain/cp-supervisor/config"
	"github.com/cpchain-network/cp-chain/cp-supervisor/flags"
)

type MainFn func(ctx context.Context, cfg *config.Config, logger log.Logger) (cliapp.Lifecycle, error)

// Main is the entrypoint into the Supervisor.
// This method returns a cliapp.LifecycleAction, to create an cp-service CLI-lifecycle-managed supervisor with.
func Main(version string, fn MainFn) cliapp.LifecycleAction {
	return func(cliCtx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
		if err := flags.CheckRequired(cliCtx); err != nil {
			return nil, err
		}
		cfg := flags.ConfigFromCLI(cliCtx, version)
		if err := cfg.Check(); err != nil {
			return nil, fmt.Errorf("invalid CLI flags: %w", err)
		}

		l := oplog.NewLogger(oplog.AppOut(cliCtx), cfg.LogConfig)
		oplog.SetGlobalLogHandler(l.Handler())
		opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, l)

		l.Info("Initializing Supervisor")
		return fn(cliCtx.Context, cfg, l)
	}
}
