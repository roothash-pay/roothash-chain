package networks

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	opnode "github.com/roothash-pay/theweb3-chain/tw-node"
	"github.com/roothash-pay/theweb3-chain/tw-node/flags"
	opflags "github.com/roothash-pay/theweb3-chain/tw-service/flags"
	oplog "github.com/roothash-pay/theweb3-chain/tw-service/log"
)

var Subcommands = []*cli.Command{
	{
		Name:  "dump-rollup-config",
		Usage: "Dumps network configs",
		Flags: []cli.Flag{
			opflags.CLINetworkFlag(flags.EnvVarPrefix, ""),
		},
		Action: func(ctx *cli.Context) error {
			logCfg := oplog.ReadCLIConfig(ctx)
			logger := oplog.NewLogger(oplog.AppOut(ctx), logCfg)

			network := ctx.String(opflags.NetworkFlagName)
			if network == "" {
				return errors.New("must specify a network name")
			}

			rCfg, err := opnode.NewRollupConfigFromCLI(logger, ctx)
			if err != nil {
				return err
			}

			out, err := json.MarshalIndent(rCfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	},
}
