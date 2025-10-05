package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/log"
	cl "github.com/roothash-pay/roothash-chain/rhs-program/client"
	"github.com/roothash-pay/roothash-chain/rhs-program/client/l2"
	"github.com/roothash-pay/roothash-chain/rhs-program/host/config"
	"github.com/roothash-pay/roothash-chain/rhs-program/host/kvstore"
)

type Prefetcher interface {
	Hint(hint string) error
	GetPreimage(ctx context.Context, key common.Hash) ([]byte, error)
}
type PrefetcherCreator func(ctx context.Context, logger log.Logger, kv kvstore.KV, cfg *config.Config) (Prefetcher, error)
type programCfg struct {
	prefetcher     PrefetcherCreator
	skipValidation bool
	db             l2.KeyValueStore
	storeBlockData bool
}

type ProgramOpt func(c *programCfg)

// WithPrefetcher configures the prefetcher used by the preimage server.
func WithPrefetcher(creator PrefetcherCreator) ProgramOpt {
	return func(c *programCfg) {
		c.prefetcher = creator
	}
}

// WithSkipValidation controls whether the program will skip validation of the derived block.
func WithSkipValidation(skip bool) ProgramOpt {
	return func(c *programCfg) {
		c.skipValidation = skip
	}
}

// WithDB sets the backing state database used by the program.
// If not set, the program will use an in-memory database.
func WithDB(db l2.KeyValueStore) ProgramOpt {
	return func(c *programCfg) {
		c.db = db
	}
}

// WithStoreBlockData controls whether block data, including intermediate trie nodes from transactions and receipts
// of the derived block should be stored in the database.
func WithStoreBlockData(store bool) ProgramOpt {
	return func(c *programCfg) {
		c.storeBlockData = store
	}
}

// FaultProofProgram is the programmatic entry-point for the fault proof program
func FaultProofProgram(ctx context.Context, logger log.Logger, cfg *config.Config, opts ...ProgramOpt) error {
	programConfig := &programCfg{
		db: memorydb.New(),
	}
	for _, opt := range opts {
		opt(programConfig)
	}
	if programConfig.prefetcher == nil {
		panic("prefetcher creator is not set")
	}
	var (
		serverErr chan error
	)

	// Use a channel to receive the server result so we can wait for it to complete before returning
	serverErr = make(chan error)
	go func() {
		defer close(serverErr)
		serverErr <- PreimageServer(ctx, logger, cfg, programConfig.prefetcher)
	}()

	var cmd *exec.Cmd
	if cfg.ExecCmd != "" {
		cmd = exec.CommandContext(ctx, cfg.ExecCmd)
		cmd.ExtraFiles = make([]*os.File, cl.MaxFd-3) // not including stdin, stdout and stderr
		cmd.Stdout = os.Stdout                        // for debugging
		cmd.Stderr = os.Stderr                        // for debugging

		err := cmd.Start()
		if err != nil {
			return fmt.Errorf("program cmd failed to start: %w", err)
		}
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("failed to wait for child program: %w", err)
		}
		logger.Debug("Client program completed successfully")
		return nil
	} else {
		var clientCfg cl.Config
		if programConfig.skipValidation {
			clientCfg.SkipValidation = true
		}
		clientCfg.InteropEnabled = cfg.InteropEnabled
		clientCfg.DB = programConfig.db
		clientCfg.StoreBlockData = programConfig.storeBlockData
		return cl.RunProgram(logger, nil, nil, clientCfg)
	}
}

// PreimageServer reads hints and preimage requests from the provided channels and processes those requests.
// This method will block until both the hinter and preimage handlers complete.
// If either returns an error both handlers are stopped.
// The supplied preimageChannel and hintChannel will be closed before this function returns.
func PreimageServer(ctx context.Context, logger log.Logger, cfg *config.Config, prefetcherCreator PrefetcherCreator) error {
	var serverDone chan error
	var hinterDone chan error
	logger.Info("Starting preimage server")
	var kv kvstore.KV

	// Close the preimage/hint channels, and then kv store once the server and hinter have exited.
	defer func() {
		if serverDone != nil {
			// Wait for pre-image server to complete
			<-serverDone
		}
		if hinterDone != nil {
			// Wait for hinter to complete
			<-hinterDone
		}

		if kv != nil {
			kv.Close()
		}
	}()

	if cfg.DataDir == "" {
		logger.Info("Using in-memory storage")
		kv = kvstore.NewMemKV()
	} else {
		if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
			return fmt.Errorf("creating datadir: %w", err)
		}
		store, err := kvstore.NewDiskKV(logger, cfg.DataDir, cfg.DataFormat)
		if err != nil {
			return fmt.Errorf("creating kvstore: %w", err)
		}
		kv = store
	}

	var _ kvstore.PreimageSource
	prefetch, err := prefetcherCreator(ctx, logger, kv, cfg)
	if err != nil {
		return fmt.Errorf("failed to create prefetcher: %w", err)
	}
	if prefetch != nil {
		_ = func(key common.Hash) ([]byte, error) { return prefetch.GetPreimage(ctx, key) }
	} else {
		logger.Info("Using offline mode. All required pre-images must be pre-populated.")
		_ = kv.Get
	}

	select {
	case err := <-serverDone:
		return err
	case err := <-hinterDone:
		return err
	case <-ctx.Done():
		logger.Info("Shutting down")
		if errors.Is(ctx.Err(), context.Canceled) {
			// We were asked to shutdown by the context being cancelled so don't treat it as an error condition.
			return nil
		}
		return ctx.Err()
	}
}

func routeHints(logger log.Logger, hHostRW io.ReadWriter) chan error {
	chErr := make(chan error)
	go func() {
		defer close(chErr)
	}()
	return chErr
}

func launchOracleServer(logger log.Logger, pHostRW io.ReadWriteCloser) chan error {
	chErr := make(chan error)
	go func() {
		defer close(chErr)
	}()
	return chErr
}
