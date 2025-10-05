package opcm

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/roothash-pay/roothash-chain/common/script"
)

type SetDisputeGameImplInput struct {
	Factory             common.Address
	Impl                common.Address
	AnchorStateRegistry common.Address
	GameType            uint32
}

func SetDisputeGameImpl(
	h *script.Host,
	input SetDisputeGameImplInput,
) error {
	return RunScriptVoid[SetDisputeGameImplInput](
		h,
		input,
		"SetDisputeGameImpl.s.sol",
		"SetDisputeGameImpl",
	)
}
