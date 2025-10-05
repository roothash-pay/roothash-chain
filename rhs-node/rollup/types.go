package rollup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/roothash-pay/roothash-chain/rhs-service/eth"
)

var (
	ErrBlockTimeZero                 = errors.New("block time cannot be 0")
	ErrMissingChannelTimeout         = errors.New("channel timeout must be set, this should cover at least a L1 block time")
	ErrInvalidSeqWindowSize          = errors.New("sequencing window size must at least be 2")
	ErrInvalidMaxSeqDrift            = errors.New("maximum sequencer drift must be greater than 0")
	ErrMissingGenesisL2Hash          = errors.New("genesis core hash cannot be empty")
	ErrMissingGenesisL2Time          = errors.New("missing core genesis time")
	ErrMissingGasLimit               = errors.New("missing genesis system config gas limit")
	ErrMissingBatchInboxAddress      = errors.New("missing batch inbox address")
	ErrMissingDepositContractAddress = errors.New("missing deposit contract address")
	ErrMissingL2ChainID              = errors.New("core chain ID must not be nil")
	ErrL2ChainIDNotPositive          = errors.New("core chain ID must be non-zero and positive")
)

type Genesis struct {
	// The L1 block that the rollup starts *after* (no derived transactions)
	// The L2 block the rollup starts from (no transactions, pre-configured state)
	L2 eth.BlockID `json:"l2"`
	// Timestamp of L2 block
	L2Time uint64 `json:"l2_time"`
	// Initial system configuration values.
	// The L2 genesis block may not include transactions, and thus cannot encode the config values,
	// unlike later L2 blocks.
	SystemConfig eth.SystemConfig `json:"system_config"`
}

type AltDAConfig struct {
	// L1 DataAvailabilityChallenge contract proxy address
	DAChallengeAddress common.Address `json:"da_challenge_contract_address,omitempty"`
	// CommitmentType specifies which commitment type can be used. Defaults to Keccak (type 0) if not present
	CommitmentType string `json:"da_commitment_type"`
	// DA challenge window value set on the DAC contract. Used in alt-da mode
	// to compute when a commitment can no longer be challenged.
	DAChallengeWindow uint64 `json:"da_challenge_window"`
	// DA resolve window value set on the DAC contract. Used in alt-da mode
	// to compute when a challenge expires and trigger a reorg if needed.
	DAResolveWindow uint64 `json:"da_resolve_window"`
}

type Config struct {
	// Genesis anchor point of the rollup
	Genesis Genesis `json:"genesis"`
	// Seconds per core block
	BlockTime uint64 `json:"block_time"`
	// Sequencer batches may not be more than MaxSequencerDrift seconds after
	// the L1 timestamp of their L1 origin time.
	//
	// Note: When L1 has many 1 second consecutive blocks, and core grows at fixed 2 seconds,
	// the core time may still grow beyond this difference.
	//
	// With Fjord, the MaxSequencerDrift becomes a constant. Use the ChainSpec
	// instead of reading this rollup configuration field directly to determine
	// the max sequencer drift for a given block based on the block's L1 origin.
	// Chains that activate Fjord at genesis may leave this field empty.
	MaxSequencerDrift uint64 `json:"max_sequencer_drift,omitempty"`
	// Number of epochs (L1 blocks) per sequencing window, including the epoch L1 origin block itself
	SeqWindowSize uint64 `json:"seq_window_size"`
	// Number of L1 blocks between when a channel can be opened and when it must be closed by.
	ChannelTimeoutBedrock uint64 `json:"channel_timeout"`
	// Required to identify the core network and create p2p signatures unique for this chain.
	L2ChainID *big.Int `json:"l2_chain_id"`

	// RegolithTime sets the activation time of the Regolith network-upgrade:
	// a pre-mainnet Bedrock change that addresses findings of the Sherlock contest related to deposit attributes.
	// "Regolith" is the loose deposited rock that sits on top of Bedrock.
	// Active if RegolithTime != nil && core block timestamp >= *RegolithTime, inactive otherwise.
	RegolithTime *uint64 `json:"regolith_time,omitempty"`

	// CanyonTime sets the activation time of the Canyon network upgrade.
	// Active if CanyonTime != nil && core block timestamp >= *CanyonTime, inactive otherwise.
	CanyonTime *uint64 `json:"canyon_time,omitempty"`

	// DeltaTime sets the activation time of the Delta network upgrade.
	// Active if DeltaTime != nil && core block timestamp >= *DeltaTime, inactive otherwise.
	DeltaTime *uint64 `json:"delta_time,omitempty"`

	// EcotoneTime sets the activation time of the Ecotone network upgrade.
	// Active if EcotoneTime != nil && core block timestamp >= *EcotoneTime, inactive otherwise.
	EcotoneTime *uint64 `json:"ecotone_time,omitempty"`

	// FjordTime sets the activation time of the Fjord network upgrade.
	// Active if FjordTime != nil && core block timestamp >= *FjordTime, inactive otherwise.
	FjordTime *uint64 `json:"fjord_time,omitempty"`

	// GraniteTime sets the activation time of the Granite network upgrade.
	// Active if GraniteTime != nil && core block timestamp >= *GraniteTime, inactive otherwise.
	GraniteTime *uint64 `json:"granite_time,omitempty"`

	// HoloceneTime sets the activation time of the Holocene network upgrade.
	// Active if HoloceneTime != nil && core block timestamp >= *HoloceneTime, inactive otherwise.
	HoloceneTime *uint64 `json:"holocene_time,omitempty"`

	// IsthmusTime sets the activation time of the Isthmus network upgrade.
	// Active if IsthmusTime != nil && core block timestamp >= *IsthmusTime, inactive otherwise.
	IsthmusTime *uint64 `json:"isthmus_time,omitempty"`

	// JovianTime sets the activation time of the Jovian network upgrade.
	// Active if JovianTime != nil && core block timestamp >= *JovianTime, inactive otherwise.
	JovianTime *uint64 `json:"jovian_time,omitempty"`

	// InteropTime sets the activation time for an experimental feature-set, activated like a hardfork.
	// Active if InteropTime != nil && core block timestamp >= *InteropTime, inactive otherwise.
	InteropTime *uint64 `json:"interop_time,omitempty"`

	// Note: below addresses are part of the block-derivation process,
	// and required to be the same network-wide to stay in consensus.

	// ChainOpConfig is the OptimismConfig of the execution layer ChainConfig.
	// It is used during safe chain consolidation to translate zero SystemConfig EIP1559
	// parameters to the protocol values, like the execution layer does.
	// If missing, it is loaded by the rhs-node from the embedded superchain config at startup.
	ChainOpConfig *params.OptimismConfig `json:"chain_op_config,omitempty"`

	// PectraBlobScheduleTime sets the time until which (but not including) the blob base fee
	// calculations for the L1 Block Info use the pre-Prague=Cancun blob parameters.
	// This feature is optional and if not active, the L1 Block Info calculation uses the Prague
	// blob parameters for the first L1 Prague block, as was intended.
	// This feature (de)activates by L1 origin timestamp, to keep a consistent L1 block info per core
	// epoch.
	PectraBlobScheduleTime *uint64 `json:"pectra_blob_schedule_time,omitempty"`
}

// ValidateL2Config checks core config variables for errors.
func (cfg *Config) ValidateL2Config(ctx context.Context, client L2Client, skipL2GenesisBlockHash bool) error {
	// Validate the core Client Chain ID
	if err := cfg.CheckL2ChainID(ctx, client); err != nil {
		return err
	}

	// Validate the Rollup core Genesis Blockhash if requested. We skip this when doing EL sync
	if skipL2GenesisBlockHash {
		return nil
	}
	if err := cfg.CheckL2GenesisBlockHash(ctx, client); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) TimestampForBlock(blockNumber uint64) uint64 {
	return cfg.Genesis.L2Time + ((blockNumber - cfg.Genesis.L2.Number) * cfg.BlockTime)
}

func (cfg *Config) TargetBlockNumber(timestamp uint64) (num uint64, err error) {
	// subtract genesis time from timestamp to get the time elapsed since genesis, and then divide that
	// difference by the block time to get the expected core block number at the current time. If the
	// unsafe head does not have this block number, then there is a gap in the queue.
	genesisTimestamp := cfg.Genesis.L2Time
	if timestamp < genesisTimestamp {
		return 0, fmt.Errorf("did not reach genesis time (%d) yet", genesisTimestamp)
	}
	wallClockGenesisDiff := timestamp - genesisTimestamp
	// Note: round down, we should not request blocks into the future.
	blocksSinceGenesis := wallClockGenesisDiff / cfg.BlockTime
	return cfg.Genesis.L2.Number + blocksSinceGenesis, nil
}

type L1Client interface {
	ChainID(context.Context) (*big.Int, error)
	L1BlockRefByNumber(context.Context, uint64) (eth.L1BlockRef, error)
}

type L2Client interface {
	ChainID(context.Context) (*big.Int, error)
	L2BlockRefByNumber(context.Context, uint64) (eth.L2BlockRef, error)
}

// CheckL2ChainID checks that the configured core chain ID matches the client's chain ID.
func (cfg *Config) CheckL2ChainID(ctx context.Context, client L2Client) error {
	id, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get core chain ID: %w", err)
	}
	if cfg.L2ChainID.Cmp(id) != 0 {
		return fmt.Errorf("incorrect core RPC chain id %d, expected %d", id, cfg.L2ChainID)
	}
	return nil
}

// CheckL2GenesisBlockHash checks that the configured core genesis block hash is valid for the given client.
func (cfg *Config) CheckL2GenesisBlockHash(ctx context.Context, client L2Client) error {
	l2GenesisBlockRef, err := client.L2BlockRefByNumber(ctx, cfg.Genesis.L2.Number)
	if err != nil {
		return fmt.Errorf("failed to get core genesis blockhash: %w", err)
	}
	if l2GenesisBlockRef.Hash != cfg.Genesis.L2.Hash {
		return fmt.Errorf("incorrect core genesis block hash %s, expected %s", l2GenesisBlockRef.Hash, cfg.Genesis.L2.Hash)
	}
	return nil
}

// Check verifies that the given configuration makes sense
func (cfg *Config) Check() error {
	if cfg.BlockTime == 0 {
		return ErrBlockTimeZero
	}
	if cfg.ChannelTimeoutBedrock == 0 {
		return ErrMissingChannelTimeout
	}
	if cfg.SeqWindowSize < 2 {
		return ErrInvalidSeqWindowSize
	}
	if cfg.MaxSequencerDrift == 0 {
		return ErrInvalidMaxSeqDrift
	}
	if cfg.Genesis.L2.Hash == (common.Hash{}) {
		return ErrMissingGenesisL2Hash
	}
	if cfg.Genesis.L2Time == 0 {
		return ErrMissingGenesisL2Time
	}
	if cfg.Genesis.SystemConfig.GasLimit == 0 {
		return ErrMissingGasLimit
	}
	if cfg.L2ChainID == nil {
		return ErrMissingL2ChainID
	}
	if cfg.L2ChainID.Sign() < 1 {
		return ErrL2ChainIDNotPositive
	}
	if err := validateAltDAConfig(cfg); err != nil {
		return err
	}

	if err := checkFork(cfg.RegolithTime, cfg.CanyonTime, Regolith, Canyon); err != nil {
		return err
	}
	if err := checkFork(cfg.CanyonTime, cfg.DeltaTime, Canyon, Delta); err != nil {
		return err
	}
	if err := checkFork(cfg.DeltaTime, cfg.EcotoneTime, Delta, Ecotone); err != nil {
		return err
	}
	if err := checkFork(cfg.EcotoneTime, cfg.FjordTime, Ecotone, Fjord); err != nil {
		return err
	}
	if err := checkFork(cfg.FjordTime, cfg.GraniteTime, Fjord, Granite); err != nil {
		return err
	}
	if err := checkFork(cfg.GraniteTime, cfg.HoloceneTime, Granite, Holocene); err != nil {
		return err
	}
	if err := checkFork(cfg.HoloceneTime, cfg.IsthmusTime, Holocene, Isthmus); err != nil {
		return err
	}

	return nil
}

func (cfg *Config) HasOptimismWithdrawalsRoot(timestamp uint64) bool {
	return cfg.IsIsthmus(timestamp)
}

// validateAltDAConfig checks the two approaches to configuring alt-da mode.
// If the legacy values are set, they are copied to the new location. If both are set, they are check for consistency.
func validateAltDAConfig(cfg *Config) error {
	return nil
}

// checkFork checks that fork A is before or at the same time as fork B
func checkFork(a, b *uint64, aName, bName ForkName) error {
	if a == nil && b == nil {
		return nil
	}
	if a == nil && b != nil {
		return fmt.Errorf("fork %s set (to %d), but prior fork %s missing", bName, *b, aName)
	}
	if a != nil && b == nil {
		return nil
	}
	if *a > *b {
		return fmt.Errorf("fork %s set to %d, but prior fork %s has higher offset %d", bName, *b, aName, *a)
	}
	return nil
}

// IsRegolith returns true if the Regolith hardfork is active at or past the given timestamp.
func (c *Config) IsRegolith(timestamp uint64) bool {
	return c.RegolithTime != nil && timestamp >= *c.RegolithTime
}

// IsCanyon returns true if the Canyon hardfork is active at or past the given timestamp.
func (c *Config) IsCanyon(timestamp uint64) bool {
	return c.CanyonTime != nil && timestamp >= *c.CanyonTime
}

// IsDelta returns true if the Delta hardfork is active at or past the given timestamp.
func (c *Config) IsDelta(timestamp uint64) bool {
	return c.DeltaTime != nil && timestamp >= *c.DeltaTime
}

// IsEcotone returns true if the Ecotone hardfork is active at or past the given timestamp.
func (c *Config) IsEcotone(timestamp uint64) bool {
	return c.EcotoneTime != nil && timestamp >= *c.EcotoneTime
}

// IsFjord returns true if the Fjord hardfork is active at or past the given timestamp.
func (c *Config) IsFjord(timestamp uint64) bool {
	return c.FjordTime != nil && timestamp >= *c.FjordTime
}

// IsGranite returns true if the Granite hardfork is active at or past the given timestamp.
func (c *Config) IsGranite(timestamp uint64) bool {
	return c.GraniteTime != nil && timestamp >= *c.GraniteTime
}

// IsHolocene returns true if the Holocene hardfork is active at or past the given timestamp.
func (c *Config) IsHolocene(timestamp uint64) bool {
	return c.HoloceneTime != nil && timestamp >= *c.HoloceneTime
}

// IsIsthmus returns true if the Isthmus hardfork is active at or past the given timestamp.
func (c *Config) IsIsthmus(timestamp uint64) bool {
	return c.IsthmusTime != nil && timestamp >= *c.IsthmusTime
}

// IsJovian returns true if the Jovian hardfork is active at or past the given timestamp.
func (c *Config) IsJovian(timestamp uint64) bool {
	return c.JovianTime != nil && timestamp >= *c.JovianTime
}

// IsInterop returns true if the Interop hardfork is active at or past the given timestamp.
func (c *Config) IsInterop(timestamp uint64) bool {
	return c.InteropTime != nil && timestamp >= *c.InteropTime
}

func (c *Config) IsRegolithActivationBlock(l2BlockTime uint64) bool {
	return c.IsRegolith(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsRegolith(l2BlockTime-c.BlockTime)
}

func (c *Config) IsCanyonActivationBlock(l2BlockTime uint64) bool {
	return c.IsCanyon(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsCanyon(l2BlockTime-c.BlockTime)
}

func (c *Config) IsDeltaActivationBlock(l2BlockTime uint64) bool {
	return c.IsDelta(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsDelta(l2BlockTime-c.BlockTime)
}

// IsEcotoneActivationBlock returns whether the specified block is the first block subject to the
// Ecotone upgrade. Ecotone activation at genesis does not count.
func (c *Config) IsEcotoneActivationBlock(l2BlockTime uint64) bool {
	return c.IsEcotone(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsEcotone(l2BlockTime-c.BlockTime)
}

// IsFjordActivationBlock returns whether the specified block is the first block subject to the
// Fjord upgrade.
func (c *Config) IsFjordActivationBlock(l2BlockTime uint64) bool {
	return c.IsFjord(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsFjord(l2BlockTime-c.BlockTime)
}

// IsGraniteActivationBlock returns whether the specified block is the first block subject to the
// Granite upgrade.
func (c *Config) IsGraniteActivationBlock(l2BlockTime uint64) bool {
	return c.IsGranite(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsGranite(l2BlockTime-c.BlockTime)
}

// IsHoloceneActivationBlock returns whether the specified block is the first block subject to the
// Holocene upgrade.
func (c *Config) IsHoloceneActivationBlock(l2BlockTime uint64) bool {
	return c.IsHolocene(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsHolocene(l2BlockTime-c.BlockTime)
}

// IsIsthmusActivationBlock returns whether the specified block is the first block subject to the
// Isthmus upgrade.
func (c *Config) IsIsthmusActivationBlock(l2BlockTime uint64) bool {
	return c.IsIsthmus(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsIsthmus(l2BlockTime-c.BlockTime)
}

// IsJovianActivationBlock returns whether the specified block is the first block subject to the
// Jovian upgrade.
func (c *Config) IsJovianActivationBlock(l2BlockTime uint64) bool {
	return c.IsJovian(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsJovian(l2BlockTime-c.BlockTime)
}

func (c *Config) IsInteropActivationBlock(l2BlockTime uint64) bool {
	return c.IsInterop(l2BlockTime) &&
		l2BlockTime >= c.BlockTime &&
		!c.IsInterop(l2BlockTime-c.BlockTime)
}

// IsActivationBlock returns the fork which activates at the block with time newTime if the previous
// block's time is oldTime. It return an empty ForkName if no fork activation takes place between
// those timestamps. It can be used for both, L1 and core blocks.
// TODO(12490): Currently only supports Holocene. Will be modularized in a follow-up.
func (c *Config) IsActivationBlock(oldTime, newTime uint64) ForkName {
	if c.IsHolocene(newTime) && !c.IsHolocene(oldTime) {
		return Holocene
	}
	return ""
}

func (c *Config) ActivateAtGenesis(hardfork ForkName) {
	// IMPORTANT! ordered from newest to oldest
	switch hardfork {
	case Interop:
		c.InteropTime = new(uint64)
		fallthrough
	case Jovian:
		c.JovianTime = new(uint64)
		fallthrough
	case Isthmus:
		c.IsthmusTime = new(uint64)
		fallthrough
	case Holocene:
		c.HoloceneTime = new(uint64)
		fallthrough
	case Granite:
		c.GraniteTime = new(uint64)
		fallthrough
	case Fjord:
		c.FjordTime = new(uint64)
		fallthrough
	case Ecotone:
		c.EcotoneTime = new(uint64)
		fallthrough
	case Delta:
		c.DeltaTime = new(uint64)
		fallthrough
	case Canyon:
		c.CanyonTime = new(uint64)
		fallthrough
	case Regolith:
		c.RegolithTime = new(uint64)
		fallthrough
	case Bedrock:
		// default
	case None:
		break
	}
}

// ForkchoiceUpdatedVersion returns the EngineAPIMethod suitable for the chain hard fork version.
func (c *Config) ForkchoiceUpdatedVersion(attr *eth.PayloadAttributes) eth.EngineAPIMethod {
	if attr == nil {
		// Don't begin payload build process.
		return eth.FCUV3
	}
	ts := uint64(attr.Timestamp)
	if c.IsEcotone(ts) {
		// Cancun
		return eth.FCUV3
	} else if c.IsCanyon(ts) {
		// Shanghai
		return eth.FCUV2
	} else {
		// According to Ethereum engine API spec, we can use fcuV2 here,
		// but upstream Geth v1.13.11 does not accept V2 before Shanghai.
		return eth.FCUV1
	}
}

// NewPayloadVersion returns the EngineAPIMethod suitable for the chain hard fork version.
func (c *Config) NewPayloadVersion(timestamp uint64) eth.EngineAPIMethod {
	if c.IsIsthmus(timestamp) {
		return eth.NewPayloadV4
	} else if c.IsEcotone(timestamp) {
		// Cancun
		return eth.NewPayloadV3
	} else {
		return eth.NewPayloadV2
	}
}

// GetPayloadVersion returns the EngineAPIMethod suitable for the chain hard fork version.
func (c *Config) GetPayloadVersion(timestamp uint64) eth.EngineAPIMethod {
	if c.IsIsthmus(timestamp) {
		return eth.GetPayloadV4
	} else if c.IsEcotone(timestamp) {
		// Cancun
		return eth.GetPayloadV3
	} else {
		return eth.GetPayloadV2
	}
}

// SyncLookback computes the number of blocks to walk back in order to find the correct L1 origin.
// In alt-da mode longest possible window is challenge + resolve windows.
func (c *Config) SyncLookback() uint64 {
	return c.SeqWindowSize
}

// Description outputs a banner describing the important parts of rollup configuration in a human-readable form.
// Optionally provide a mapping of core chain IDs to network names to label the core chain with if not unknown.
// The config should be config.Check()-ed before creating a description.
func (c *Config) Description(l2Chains map[string]string) string {
	// Find and report the network the user is running
	var banner string
	networkL2 := ""
	if l2Chains != nil {
		networkL2 = l2Chains[c.L2ChainID.String()]
	}
	if networkL2 == "" {
		networkL2 = "unknown core"
	}
	banner += fmt.Sprintf("core Chain ID: %v (%s)\n", c.L2ChainID, networkL2)
	// Report the genesis configuration
	banner += "Bedrock starting point:\n"
	banner += fmt.Sprintf("  core starting time: %d ~ %s\n", c.Genesis.L2Time, fmtTime(c.Genesis.L2Time))
	banner += fmt.Sprintf("  core block: %s %d\n", c.Genesis.L2.Hash, c.Genesis.L2.Number)
	// Report the upgrade configuration
	banner += "Post-Bedrock Network Upgrades (timestamp based):\n"
	c.forEachFork(func(name string, _ string, time *uint64) {
		banner += fmt.Sprintf("  - %v: %s\n", name, fmtForkTimeOrUnset(time))
	})
	// Report the protocol version
	banner += fmt.Sprintf("Node supports up to OP-Stack Protocol Version: %s\n", OPStackSupport)
	return banner
}

// LogDescription outputs a banner describing the important parts of rollup configuration in a log format.
// Optionally provide a mapping of core chain IDs to network names to label the core chain with if not unknown.
// The config should be config.Check()-ed before creating a description.
func (c *Config) LogDescription(log log.Logger, l2Chains map[string]string) {
	// Find and report the network the user is running
	networkL2 := ""
	if l2Chains != nil {
		networkL2 = l2Chains[c.L2ChainID.String()]
	}
	if networkL2 == "" {
		networkL2 = "unknown core"
	}

	ctx := []any{
		"l2_chain_id", c.L2ChainID,
		"l2_network", networkL2,
		"l2_start_time", c.Genesis.L2Time,
		"l2_block_hash", c.Genesis.L2.Hash.String(),
		"l2_block_number", c.Genesis.L2.Number,
	}
	c.forEachFork(func(_ string, logName string, time *uint64) {
		ctx = append(ctx, logName, fmtForkTimeOrUnset(time))
	})
	if c.PectraBlobScheduleTime != nil {
		// only print in config if set at all
		ctx = append(ctx, "pectra_blob_schedule_time", fmtForkTimeOrUnset(c.PectraBlobScheduleTime))
	}
	log.Info("Rollup Config", ctx...)
}

func (c *Config) forEachFork(callback func(name string, logName string, time *uint64)) {
	callback("Regolith", "regolith_time", c.RegolithTime)
	callback("Canyon", "canyon_time", c.CanyonTime)
	callback("Delta", "delta_time", c.DeltaTime)
	callback("Ecotone", "ecotone_time", c.EcotoneTime)
	callback("Fjord", "fjord_time", c.FjordTime)
	callback("Granite", "granite_time", c.GraniteTime)
	callback("Holocene", "holocene_time", c.HoloceneTime)
	if c.PectraBlobScheduleTime != nil {
		// only report if config is set
		callback("Pectra Blob Schedule", "pectra_blob_schedule_time", c.PectraBlobScheduleTime)
	}
	callback("Isthmus", "isthmus_time", c.IsthmusTime)
	callback("Jovian", "jovian_time", c.JovianTime)
	callback("Interop", "interop_time", c.InteropTime)
}

func (c *Config) ParseRollupConfig(in io.Reader) error {
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	if err := dec.Decode(c); err != nil {
		return fmt.Errorf("failed to decode rollup config: %w", err)
	}
	return nil
}

func fmtForkTimeOrUnset(v *uint64) string {
	if v == nil {
		return "(not configured)"
	}
	if *v == 0 { // don't output the unix epoch time if it's really just activated at genesis.
		return "@ genesis"
	}
	return fmt.Sprintf("@ %-10v ~ %s", *v, fmtTime(*v))
}

func fmtTime(v uint64) string {
	return time.Unix(int64(v), 0).Format(time.UnixDate)
}

type Epoch uint64
