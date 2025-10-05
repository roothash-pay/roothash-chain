package client

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/log"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/boot"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/claim"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/interop"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l1"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l2"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/tasks"
	oplog "github.com/roothash-pay/roothash-chain/rhs-service/log"
)

var errInvalidConfig = errors.New("invalid config")

type Config struct {
	SkipValidation bool
	InteropEnabled bool
	DB             l2.KeyValueStore
	StoreBlockData bool
}

// Main executes the client program in a detached context and exits the current process.
// The client runtime environment must be preset before calling this function.
func Main(useInterop bool) {
	// Default to a machine parsable but relatively human friendly log format.
	// Don't do anything fancy to detect if color output is supported.
	logger := oplog.NewLogger(os.Stdout, oplog.CLIConfig{
		Level:  log.LevelInfo,
		Format: oplog.FormatLogFmt,
		Color:  false,
	})
	oplog.SetGlobalLogHandler(logger.Handler())

	logger.Info("Starting fault proof program client", "useInterop", useInterop)
	var preimageOracle io.ReadWriter
	var preimageHinter io.ReadWriter
	config := Config{
		InteropEnabled: useInterop,
		DB:             memorydb.New(),
	}
	if err := RunProgram(logger, preimageOracle, preimageHinter, config); errors.Is(err, claim.ErrClaimNotValid) {
		log.Error("Claim is invalid", "err", err)
		os.Exit(1)
	} else if err != nil {
		log.Error("Program failed", "err", err)
		os.Exit(2)
	} else {
		log.Info("Claim successfully verified")
		os.Exit(0)
	}
}

// RunProgram executes the Program, while attached to an IO based pre-image oracle, to be served by a host.
func RunProgram(logger log.Logger, preimageOracle io.ReadWriter, preimageHinter io.ReadWriter, cfg Config) error {

	l1PreimageOracle := l1.NewCachingOracle(l1.NewPreimageOracle())
	l2PreimageOracle := l2.NewCachingOracle(l2.NewPreimageOracle(nil, nil, cfg.InteropEnabled))

	if cfg.InteropEnabled {
		bootInfo := boot.BootstrapInterop(nil)
		return interop.RunInteropProgram(logger, bootInfo, l1PreimageOracle, l2PreimageOracle, !cfg.SkipValidation)
	}
	if cfg.DB == nil {
		return fmt.Errorf("%w: db config is required", errInvalidConfig)
	}
	bootInfo := boot.NewBootstrapClient(nil).BootInfo()
	derivationOptions := tasks.DerivationOptions{StoreBlockData: cfg.StoreBlockData}
	return RunPreInteropProgram(logger, bootInfo, l1PreimageOracle, l2PreimageOracle, cfg.DB, derivationOptions)
}
