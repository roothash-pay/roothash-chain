package derive

import "github.com/roothash-pay/theweb3-chain/tw-service/testutils"

var _ L1Fetcher = (*testutils.MockL1Source)(nil)

var _ Metrics = (*testutils.TestDerivationMetrics)(nil)
