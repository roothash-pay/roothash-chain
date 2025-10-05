package cli

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	"github.com/roothash-pay/roothash-chain/rhs-node/flags"
	"github.com/roothash-pay/roothash-chain/rhs-node/p2p"
	opsigner "github.com/roothash-pay/roothash-chain/rhs-service/signer"
)

// LoadSignerSetup loads a configuration for a Signer to be set up later
func LoadSignerSetup(ctx *cli.Context, logger log.Logger) (p2p.SignerSetup, common.Address, error) {
	key := ctx.String(flags.SequencerP2PKeyName)
	address := ctx.String(flags.SequencerP2PAddressName)
	signerCfg := opsigner.ReadCLIConfig(ctx)
	if key != "" {
		// Mnemonics are bad because they leak *all* keys when they leak.
		// Unencrypted keys from file are bad because they are easy to leak (and we are not checking file permissions).
		priv, err := crypto.HexToECDSA(strings.TrimPrefix(key, "0x"))
		if err != nil {
			return nil, common.Address{}, fmt.Errorf("failed to read batch submitter key: %w", err)
		}
		return &p2p.PreparedSigner{Signer: opsigner.NewLocalSigner(priv)}, crypto.PubkeyToAddress(priv.PublicKey), nil
	} else if signerCfg.Enabled() {
		remoteSigner, err := opsigner.NewRemoteSigner(logger, signerCfg)
		if err != nil {
			return nil, common.Address{}, err
		}
		return &p2p.PreparedSigner{Signer: remoteSigner}, common.Address{}, nil
	}

	return nil, common.HexToAddress(address), nil
}
