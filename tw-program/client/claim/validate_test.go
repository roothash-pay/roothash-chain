package claim

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/log"

	"github.com/roothash-pay/theweb3-chain/cp-service/eth"
	"github.com/roothash-pay/theweb3-chain/cp-service/testlog"
)

func TestValidateClaim(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		expected := eth.Bytes32{0x11}
		actual := eth.Bytes32{0x11}
		logger := testlog.Logger(t, log.LevelError)
		err := ValidateClaim(logger, expected, actual)
		require.NoError(t, err)
	})

	t.Run("Invalid", func(t *testing.T) {
		expected := eth.Bytes32{0x11}
		actual := eth.Bytes32{0x22}
		logger := testlog.Logger(t, log.LevelError)
		err := ValidateClaim(logger, expected, actual)
		require.ErrorIs(t, err, ErrClaimNotValid)
	})
}
