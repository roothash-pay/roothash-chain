package boot

const (
	L1HeadLocalIndex = iota + 1
	L2OutputRootLocalIndex
	L2ClaimLocalIndex
	L2ClaimBlockNumberLocalIndex
	L2ChainIDLocalIndex

	// These local keys are only used for custom chains
	L2ChainConfigLocalIndex
	RollupConfigLocalIndex
	DependencySetLocalIndex
)

type oracleClient interface {
	Get(key uint64) []byte
}
