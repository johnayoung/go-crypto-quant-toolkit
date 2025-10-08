package mechanisms

import (
	"context"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// LiquidityPool represents AMM-style liquidity provision mechanisms.
// Implementations include constant product (Uniswap V2), concentrated liquidity
// (Uniswap V3), stable swaps (Curve), weighted pools (Balancer), and more.
//
// Contract:
//   - Calculate must return position state for given parameters without modifying pool state
//   - AddLiquidity must return a valid PoolPosition that can be used with RemoveLiquidity
//   - RemoveLiquidity must accept positions created by AddLiquidity
//   - All methods should validate inputs and return descriptive errors for invalid parameters
//
// Error Conditions:
//   - Invalid token amounts (negative, zero when required)
//   - Insufficient liquidity for operations
//   - Price out of range (for concentrated liquidity)
//   - Mathematical errors (overflow, division by zero)
//
// Thread Safety: Implementations are not required to be thread-safe.
type LiquidityPool interface {
	MarketMechanism

	// Calculate computes the current state of a liquidity pool given parameters.
	// This is a pure function that does not modify pool state.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - params: Pool-specific parameters (reserves, fees, tick ranges, etc.)
	//
	// Returns:
	//   - PoolState: Current calculated state including prices, liquidity, fees
	//   - error: Returns error if parameters are invalid or calculation fails
	Calculate(ctx context.Context, params PoolParams) (PoolState, error)

	// AddLiquidity simulates adding liquidity to the pool.
	// Returns a position representing the liquidity provision.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - amounts: Token amounts to add to the pool
	//
	// Returns:
	//   - PoolPosition: Position representing the added liquidity
	//   - error: Returns error if amounts are invalid or operation would fail
	AddLiquidity(ctx context.Context, amounts TokenAmounts) (PoolPosition, error)

	// RemoveLiquidity simulates removing liquidity from the pool.
	// Returns the token amounts that would be withdrawn.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - position: Position to remove (must be from AddLiquidity)
	//
	// Returns:
	//   - TokenAmounts: Amounts that would be withdrawn
	//   - error: Returns error if position is invalid or operation would fail
	RemoveLiquidity(ctx context.Context, position PoolPosition) (TokenAmounts, error)
}

// PoolParams contains pool-specific parameters for calculations.
// Different pool types use different subsets of these parameters.
//
// Examples:
//   - Constant product: ReserveA, ReserveB, FeeRate
//   - Concentrated liquidity: ReserveA, ReserveB, CurrentTick, FeeRate
//   - Stable swap: Reserves (multiple), AmplificationFactor, FeeRate
type PoolParams struct {
	// ReserveA is the reserve of token A
	ReserveA primitives.Amount

	// ReserveB is the reserve of token B
	ReserveB primitives.Amount

	// FeeRate is the swap fee rate (e.g., 0.003 for 0.3%)
	FeeRate primitives.Decimal

	// Additional pool-specific parameters can be stored here
	// Implementations should document which fields they use
	Metadata map[string]interface{}
}

// PoolState represents the computed state of a liquidity pool.
// Contains derived values from pool parameters.
type PoolState struct {
	// SpotPrice is the current spot price of token A in terms of token B
	SpotPrice primitives.Price

	// Liquidity is the total liquidity in the pool (pool-specific definition)
	Liquidity primitives.Amount

	// EffectiveLiquidity is the liquidity available for trades (may differ from total)
	EffectiveLiquidity primitives.Amount

	// AccumulatedFeesA is the accumulated fees in token A
	AccumulatedFeesA primitives.Amount

	// AccumulatedFeesB is the accumulated fees in token B
	AccumulatedFeesB primitives.Amount

	// Additional state values can be stored here
	Metadata map[string]interface{}
}

// TokenAmounts represents quantities of two tokens in a pool.
// Used for liquidity operations and withdrawal amounts.
type TokenAmounts struct {
	// AmountA is the amount of token A
	AmountA primitives.Amount

	// AmountB is the amount of token B
	AmountB primitives.Amount
}

// PoolPosition represents a liquidity position in a pool.
// This is the concrete type returned by AddLiquidity.
//
// Position Lifecycle:
//  1. Created by AddLiquidity
//  2. Can be valued using Calculate
//  3. Can be closed using RemoveLiquidity
//
// Implementations should include enough information to:
//   - Identify the position uniquely
//   - Calculate current value
//   - Support removal operations
type PoolPosition struct {
	// PoolID identifies the pool this position belongs to
	PoolID string

	// Liquidity is the amount of liquidity this position represents
	Liquidity primitives.Amount

	// TokensDeposited are the original token amounts deposited
	TokensDeposited TokenAmounts

	// Additional position-specific data (e.g., tick range for concentrated liquidity)
	Metadata map[string]interface{}
}
