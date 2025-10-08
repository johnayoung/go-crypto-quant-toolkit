// Package blackscholes implements Black-Scholes option pricing model.
// This package provides a reference implementation of the Derivative interface
// using the classic Black-Scholes-Merton formula for European options.
package blackscholes

import (
	"context"
	"errors"
	"math"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

var (
	// ErrInvalidStrike is returned when the strike price is invalid
	ErrInvalidStrike = errors.New("strike price must be positive")

	// ErrInvalidUnderlying is returned when the underlying price is invalid
	ErrInvalidUnderlying = errors.New("underlying price must be positive")

	// ErrInvalidVolatility is returned when volatility is invalid
	ErrInvalidVolatility = errors.New("volatility must be non-negative")

	// ErrInvalidTimeToExpiry is returned when time to expiry is invalid
	ErrInvalidTimeToExpiry = errors.New("time to expiry must be non-negative")

	// ErrOptionExpired is returned when attempting operations on expired options
	ErrOptionExpired = errors.New("option has expired")
)

// Option represents a European option using the Black-Scholes pricing model.
//
// The Black-Scholes model makes several assumptions:
//   - European-style exercise (only at expiration)
//   - No dividends during the option's life
//   - Markets are efficient (no arbitrage)
//   - Volatility and risk-free rate are constant
//   - Returns are log-normally distributed
//
// Thread Safety: This implementation is not thread-safe. Concurrent access
// should be protected by the caller.
type Option struct {
	// optionID uniquely identifies this option
	optionID string

	// optionType is either call or put
	optionType mechanisms.OptionType

	// strikePrice is the strike price (K) of the option
	strikePrice primitives.Price

	// timeToExpiry is the time remaining until expiration in years (T)
	// This is stored at creation but can be overridden in pricing calls
	timeToExpiry primitives.Decimal

	// entryPrice is the price at which the position was entered (for settlement)
	entryPrice primitives.Price

	// positionSize is the number of contracts held (positive for long, negative for short)
	positionSize primitives.Decimal

	// direction indicates long or short position
	direction mechanisms.PositionDirection

	// settled indicates if the option has been settled
	settled bool
}

// NewOption creates a new European option.
//
// Parameters:
//   - optionID: Unique identifier for this option
//   - optionType: Call or Put
//   - strikePrice: Strike price (must be positive)
//   - timeToExpiry: Time to expiry in years (must be non-negative)
//   - entryPrice: Price at which position was entered
//   - positionSize: Number of contracts (positive for long, negative for short)
//
// Returns error if any parameter is invalid.
func NewOption(
	optionID string,
	optionType mechanisms.OptionType,
	strikePrice primitives.Price,
	timeToExpiry primitives.Decimal,
	entryPrice primitives.Price,
	positionSize primitives.Decimal,
) (*Option, error) {
	if optionID == "" {
		return nil, errors.New("optionID cannot be empty")
	}

	if optionType != mechanisms.OptionTypeCall && optionType != mechanisms.OptionTypePut {
		return nil, errors.New("invalid option type")
	}

	if strikePrice.IsZero() {
		return nil, ErrInvalidStrike
	}

	if timeToExpiry.LessThan(primitives.Zero()) {
		return nil, ErrInvalidTimeToExpiry
	}

	if entryPrice.IsZero() {
		return nil, errors.New("entry price must be positive")
	}

	// Determine position direction from position size
	direction := mechanisms.PositionDirectionLong
	if positionSize.IsNegative() {
		direction = mechanisms.PositionDirectionShort
	}

	return &Option{
		optionID:     optionID,
		optionType:   optionType,
		strikePrice:  strikePrice,
		timeToExpiry: timeToExpiry,
		entryPrice:   entryPrice,
		positionSize: positionSize,
		direction:    direction,
		settled:      false,
	}, nil
}

// Mechanism returns the mechanism type identifier.
func (o *Option) Mechanism() mechanisms.MechanismType {
	return mechanisms.MechanismTypeDerivative
}

// Venue returns the venue identifier.
func (o *Option) Venue() string {
	return "black-scholes"
}

// Price calculates the Black-Scholes price for the option.
//
// Required parameters:
//   - UnderlyingPrice: Current price of the underlying asset (S)
//   - Volatility: Implied volatility (σ) as decimal (e.g., 0.20 for 20%)
//   - RiskFreeRate: Risk-free rate (r) as decimal (e.g., 0.05 for 5%)
//   - TimeToExpiry: Time to expiry in years (T) - optional, uses stored value if zero
//
// The Black-Scholes formula:
//   - Call: C = S*N(d1) - K*e^(-rT)*N(d2)
//   - Put: P = K*e^(-rT)*N(-d2) - S*N(-d1)
//
// Where:
//   - d1 = [ln(S/K) + (r + σ²/2)*T] / (σ*√T)
//   - d2 = d1 - σ*√T
//   - N(x) is the cumulative standard normal distribution
func (o *Option) Price(ctx context.Context, params mechanisms.PriceParams) (primitives.Price, error) {
	// Validate required parameters
	if params.UnderlyingPrice.IsZero() {
		return primitives.ZeroPrice(), ErrInvalidUnderlying
	}

	if params.Volatility.LessThan(primitives.Zero()) {
		return primitives.ZeroPrice(), ErrInvalidVolatility
	}

	// Use provided TimeToExpiry or fall back to stored value
	timeToExpiry := params.TimeToExpiry
	if timeToExpiry.IsZero() {
		timeToExpiry = o.timeToExpiry
	}

	if timeToExpiry.LessThan(primitives.Zero()) {
		return primitives.ZeroPrice(), ErrInvalidTimeToExpiry
	}

	// Handle expiry case (T = 0)
	if timeToExpiry.IsZero() {
		// At expiry, option value is intrinsic value
		return o.intrinsicValue(params.UnderlyingPrice)
	}

	// Convert to float64 for mathematical operations
	S := params.UnderlyingPrice.Decimal().Float64()
	K := o.strikePrice.Decimal().Float64()
	sigma := params.Volatility.Float64()
	r := params.RiskFreeRate.Float64()
	T := timeToExpiry.Float64()

	// Calculate d1 and d2
	// d1 = [ln(S/K) + (r + σ²/2)*T] / (σ*√T)
	sqrtT := math.Sqrt(T)
	sigmaT := sigma * sqrtT

	if sigmaT == 0 {
		// When sigma*sqrt(T) = 0, use intrinsic value
		return o.intrinsicValue(params.UnderlyingPrice)
	}

	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / sigmaT
	d2 := d1 - sigmaT

	// Calculate option price using Black-Scholes formula
	var price float64
	if o.optionType == mechanisms.OptionTypeCall {
		// Call: C = S*N(d1) - K*e^(-rT)*N(d2)
		price = S*cumulativeNormal(d1) - K*math.Exp(-r*T)*cumulativeNormal(d2)
	} else {
		// Put: P = K*e^(-rT)*N(-d2) - S*N(-d1)
		price = K*math.Exp(-r*T)*cumulativeNormal(-d2) - S*cumulativeNormal(-d1)
	}

	// Ensure non-negative price
	if price < 0 {
		price = 0
	}

	// Convert back to primitives.Price
	priceDec := primitives.NewDecimalFromFloat(price)
	return primitives.NewPrice(priceDec)
}

// Greeks calculates the option Greeks (risk sensitivities).
//
// Returns:
//   - Delta: Rate of change of option price with respect to underlying price
//   - Gamma: Rate of change of delta with respect to underlying price
//   - Theta: Rate of change of option price with respect to time (per year)
//   - Vega: Rate of change of option price with respect to volatility (per 1% change)
//   - Rho: Rate of change of option price with respect to risk-free rate (per 1% change)
func (o *Option) Greeks(ctx context.Context, params mechanisms.PriceParams) (mechanisms.Greeks, error) {
	// Validate required parameters
	if params.UnderlyingPrice.IsZero() {
		return mechanisms.Greeks{}, ErrInvalidUnderlying
	}

	if params.Volatility.LessThan(primitives.Zero()) {
		return mechanisms.Greeks{}, ErrInvalidVolatility
	}

	// Use provided TimeToExpiry or fall back to stored value
	timeToExpiry := params.TimeToExpiry
	if timeToExpiry.IsZero() {
		timeToExpiry = o.timeToExpiry
	}

	if timeToExpiry.LessThan(primitives.Zero()) {
		return mechanisms.Greeks{}, ErrInvalidTimeToExpiry
	}

	// At expiry, most Greeks are zero or undefined
	if timeToExpiry.IsZero() {
		// Delta is 1 for ITM call, -1 for ITM put, 0 otherwise
		S := params.UnderlyingPrice.Decimal()
		K := o.strikePrice.Decimal()
		var delta primitives.Decimal
		if o.optionType == mechanisms.OptionTypeCall {
			if S.GreaterThan(K) {
				delta = primitives.NewDecimal(1)
			} else {
				delta = primitives.Zero()
			}
		} else {
			if S.LessThan(K) {
				delta = primitives.NewDecimal(-1)
			} else {
				delta = primitives.Zero()
			}
		}

		return mechanisms.Greeks{
			Delta: delta,
			Gamma: primitives.Zero(),
			Theta: primitives.Zero(),
			Vega:  primitives.Zero(),
			Rho:   primitives.Zero(),
		}, nil
	}

	// Convert to float64 for calculations
	S := params.UnderlyingPrice.Decimal().Float64()
	K := o.strikePrice.Decimal().Float64()
	sigma := params.Volatility.Float64()
	r := params.RiskFreeRate.Float64()
	T := timeToExpiry.Float64()

	// Calculate d1 and d2
	sqrtT := math.Sqrt(T)
	sigmaT := sigma * sqrtT
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / sigmaT
	d2 := d1 - sigmaT

	// Calculate Greeks
	var delta, gamma, theta, vega, rho float64

	// Delta: ∂V/∂S
	if o.optionType == mechanisms.OptionTypeCall {
		delta = cumulativeNormal(d1)
	} else {
		delta = cumulativeNormal(d1) - 1
	}

	// Gamma: ∂²V/∂S² (same for calls and puts)
	gamma = standardNormal(d1) / (S * sigma * sqrtT)

	// Vega: ∂V/∂σ (same for calls and puts, per 1% change)
	vega = S * standardNormal(d1) * sqrtT / 100

	// Theta: ∂V/∂t (per year)
	discountFactor := math.Exp(-r * T)
	term1 := -(S * standardNormal(d1) * sigma) / (2 * sqrtT)
	if o.optionType == mechanisms.OptionTypeCall {
		theta = term1 - r*K*discountFactor*cumulativeNormal(d2)
	} else {
		theta = term1 + r*K*discountFactor*cumulativeNormal(-d2)
	}

	// Rho: ∂V/∂r (per 1% change)
	if o.optionType == mechanisms.OptionTypeCall {
		rho = K * T * discountFactor * cumulativeNormal(d2) / 100
	} else {
		rho = -K * T * discountFactor * cumulativeNormal(-d2) / 100
	}

	// Convert to primitives.Decimal
	deltaDec := primitives.NewDecimalFromFloat(delta)
	gammaDec := primitives.NewDecimalFromFloat(gamma)
	thetaDec := primitives.NewDecimalFromFloat(theta)
	vegaDec := primitives.NewDecimalFromFloat(vega)
	rhoDec := primitives.NewDecimalFromFloat(rho)

	return mechanisms.Greeks{
		Delta: deltaDec,
		Gamma: gammaDec,
		Theta: thetaDec,
		Vega:  vegaDec,
		Rho:   rhoDec,
	}, nil
}

// Settle calculates the settlement value of the option at expiration.
//
// Returns the intrinsic value: max(S-K, 0) for calls, max(K-S, 0) for puts.
// For positioned options, this is multiplied by position size and direction.
//
// Note: This method requires the final underlying price to be passed via context metadata
// with key "final_price". In practice, strategies would call this after receiving
// the final price at expiration.
func (o *Option) Settle(ctx context.Context) (primitives.Amount, error) {
	if o.settled {
		return primitives.ZeroAmount(), errors.New("option already settled")
	}

	// Extract final price from context metadata
	// In a real implementation, this would come from the strategy's market snapshot
	return primitives.ZeroAmount(), errors.New("settle requires final underlying price in context metadata with key 'final_price'")
}

// SettleWithPrice settles the option given a final underlying price.
// This is a helper method that calculates settlement value.
func (o *Option) SettleWithPrice(finalPrice primitives.Price) (primitives.Amount, error) {
	if o.settled {
		return primitives.ZeroAmount(), errors.New("option already settled")
	}

	intrinsic, err := o.intrinsicValue(finalPrice)
	if err != nil {
		return primitives.ZeroAmount(), err
	}

	// Calculate P&L: (intrinsic value - entry price) * position size
	pnlPerContract := intrinsic.Decimal().Sub(o.entryPrice.Decimal())
	totalPnl := pnlPerContract.Mul(o.positionSize)

	o.settled = true

	return primitives.NewAmount(totalPnl.Abs())
}

// intrinsicValue calculates the intrinsic value of the option.
// Call: max(S - K, 0)
// Put: max(K - S, 0)
func (o *Option) intrinsicValue(underlyingPrice primitives.Price) (primitives.Price, error) {
	S := underlyingPrice.Decimal()
	K := o.strikePrice.Decimal()

	var intrinsic primitives.Decimal
	if o.optionType == mechanisms.OptionTypeCall {
		intrinsic = S.Sub(K)
		if intrinsic.LessThan(primitives.Zero()) {
			intrinsic = primitives.Zero()
		}
	} else {
		intrinsic = K.Sub(S)
		if intrinsic.LessThan(primitives.Zero()) {
			intrinsic = primitives.Zero()
		}
	}

	return primitives.NewPrice(intrinsic)
}

// cumulativeNormal calculates the cumulative standard normal distribution N(x).
// Uses the approximation by Abramowitz and Stegun (1964).
// Accurate to about 7.5e-8.
func cumulativeNormal(x float64) float64 {
	// Constants for the approximation
	const (
		a1 = 0.31938153
		a2 = -0.356563782
		a3 = 1.781477937
		a4 = -1.821255978
		a5 = 1.330274429
	)

	k := 1.0 / (1.0 + 0.2316419*math.Abs(x))
	w := ((((a5*k+a4)*k+a3)*k+a2)*k + a1) * k

	// Standard normal PDF
	phi := standardNormal(x)

	if x >= 0 {
		return 1.0 - phi*w
	}
	return phi * w
}

// standardNormal calculates the standard normal probability density function φ(x).
// φ(x) = (1/√(2π)) * e^(-x²/2)
func standardNormal(x float64) float64 {
	return math.Exp(-0.5*x*x) / math.Sqrt(2*math.Pi)
}

// OptionID returns the option identifier.
func (o *Option) OptionID() string {
	return o.optionID
}

// OptionType returns the option type (call or put).
func (o *Option) OptionType() mechanisms.OptionType {
	return o.optionType
}

// StrikePrice returns the strike price.
func (o *Option) StrikePrice() primitives.Price {
	return o.strikePrice
}

// TimeToExpiry returns the stored time to expiry.
func (o *Option) TimeToExpiry() primitives.Decimal {
	return o.timeToExpiry
}

// IsSettled returns whether the option has been settled.
func (o *Option) IsSettled() bool {
	return o.settled
}
