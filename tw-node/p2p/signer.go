package p2p

import (
	"context"
	opsigner "github.com/roothash-pay/theweb3-chain/tw-service/signer"
)

type Signer interface {
	opsigner.BlockSigner
}

type PreparedSigner struct {
	Signer opsigner.BlockSigner
}

func (p *PreparedSigner) SetupSigner(ctx context.Context) (Signer, error) {
	return p.Signer, nil
}

type SignerSetup interface {
	SetupSigner(ctx context.Context) (Signer, error)
}
