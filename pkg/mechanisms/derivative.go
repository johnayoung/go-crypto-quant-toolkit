package mechanisms

import (
	"context"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// Derivative represents any derivative instrument including options, perpetuals,
// futures, and more complex structures.
//
// Contract:
//   - Price must return the fair value of the derivative given market parameters
//   - Greeks must return risk sensitivities when applicable (may return zero values for non-option derivatives)
//   - Settle must compute the settlement value at expiration/close
//   - All methods should validate inputs and return descriptive errors
//
// Error Conditions:
//   - Invalid price parameters (negative prices, invalid volatilities)
//   - Time values that don't make sense (expiry in past, negative time to expiry)
//   - Mathematical errors (overflow, invalid Greeks calculations)
//
// Thread Safety: Implementations are not required to be thread-safe.
type Derivative interface {
	MarketMechanism

	// Price returns the current fair value of the derivative.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - params: Market parameters affecting derivative pricing
	//
	// Returns:
	//   - Price: Current fair value of the derivative
	//   - error: Returns error if parameters are invalid or pricing fails
	Price(ctx context.Context, params PriceParams) (primitives.Price, error)

	// Greeks returns the risk sensitivities of the derivative.
	// For non-option derivatives (e.g., perpetuals), some Greeks may be zero or not applicable.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - params: Market parameters for Greeks calculation
	//
	// Returns:
	//   - Greeks: Risk sensitivities (delta, gamma, theta, vega, rho)
	//   - error: Returns error if parameters are invalid or calculation fails
	Greeks(ctx context.Context, params PriceParams) (Greeks, error)

	// Settle computes the settlement value of the derivative.
	// For options: intrinsic value at expiration
	// For perpetuals: final mark price difference from entry
	// For futures: difference between final and entry price
	//
	// Returns:
	//   - Amount: Settlement value (positive = profit, negative = loss)
	//   - error: Returns error if settlement calculation fails
	Settle(ctx context.Context) (primitives.Amount, error)
}

// PriceParams contains market parameters for derivative pricing.
// Different derivative types use different subsets of these parameters.
//
// Examples:
//   - European option: UnderlyingPrice, StrikePrice, TimeToExpiry, Volatility, RiskFreeRate
//   - Perpetual: UnderlyingPrice, MarkPrice, FundingRate
//   - Future: UnderlyingPrice, TimeToExpiry
type PriceParams struct {
	// UnderlyingPrice is the current price of the underlying asset
	UnderlyingPrice primitives.Price

	// StrikePrice is the strike price (for options)
	StrikePrice primitives.Price

	// TimeToExpiry is the time remaining until expiration (in years)
	// For perpetuals, this may be irrelevant (use 0 or omit)
	TimeToExpiry primitives.Decimal

	// Volatility is the implied or historical volatility (annualized)
	// Used for options pricing (e.g., Black-Scholes)
	Volatility primitives.Decimal

	// RiskFreeRate is the risk-free interest rate (annualized)
	// Used for options pricing
	RiskFreeRate primitives.Decimal

	// FundingRate is the funding rate for perpetuals (per funding period)
	FundingRate primitives.Decimal

	// MarkPrice is the mark price for perpetuals
	MarkPrice primitives.Price

	// Additional derivative-specific parameters
	Metadata map[string]interface{}
}

// Greeks represents the risk sensitivities of a derivative.
// These are the first-order and second-order partial derivatives of
// the option price with respect to various parameters.
//
// For non-option derivatives, some Greeks may be zero or not applicable.
// Implementations should document which Greeks are relevant.
type Greeks struct {
	// Delta: ∂V/∂S - sensitivity to underlying price
	// Range: [0, 1] for calls, [-1, 0] for puts
	// For perpetuals: typically 1 or -1 depending on position direction
	Delta primitives.Decimal

	// Gamma: ∂²V/∂S² - rate of change of delta
	// Always positive for long options
	// For perpetuals: typically 0
	Gamma primitives.Decimal

	// Theta: ∂V/∂t - sensitivity to time decay
	// Typically negative for long options (value decreases over time)
	// For perpetuals: typically 0 or related to funding
	Theta primitives.Decimal

	// Vega: ∂V/∂σ - sensitivity to volatility
	// Always positive for long options
	// For perpetuals: typically 0
	Vega primitives.Decimal

	// Rho: ∂V/∂r - sensitivity to interest rate
	// Positive for calls, negative for puts
	// For perpetuals: typically 0
	Rho primitives.Decimal

	// Additional Greeks can be added as needed (e.g., Vanna, Volga, Charm)
}

// OptionType represents the type of option (call or put).
type OptionType string

const (
	// OptionTypeCall represents a call option (right to buy)
	OptionTypeCall OptionType = "call"

	// OptionTypePut represents a put option (right to sell)
	OptionTypePut OptionType = "put"
)

// PositionDirection represents the direction of a derivative position.
type PositionDirection string

const (
	// PositionDirectionLong represents a long position
	PositionDirectionLong PositionDirection = "long"

	// PositionDirectionShort represents a short position
	PositionDirectionShort PositionDirection = "short"
)
