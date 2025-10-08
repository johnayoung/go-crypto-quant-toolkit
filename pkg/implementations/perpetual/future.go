// Package perpetual implements perpetual futures contracts.
// This package provides a reference implementation of the Derivative interface
// for perpetual swap contracts with funding rate mechanics.
package perpetual

import (
	"context"
	"errors"
	"time"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

var (
	// ErrInvalidMarkPrice is returned when the mark price is invalid
	ErrInvalidMarkPrice = errors.New("mark price must be positive")

	// ErrInvalidIndexPrice is returned when the index price is invalid
	ErrInvalidIndexPrice = errors.New("index price must be positive")

	// ErrInvalidFundingRate is returned when the funding rate is invalid
	ErrInvalidFundingRate = errors.New("funding rate is invalid")

	// ErrInvalidLeverage is returned when leverage is invalid
	ErrInvalidLeverage = errors.New("leverage must be positive")

	// ErrInvalidPositionSize is returned when position size is invalid
	ErrInvalidPositionSize = errors.New("position size cannot be zero")
)

// Future represents a perpetual futures contract.
//
// Perpetual futures are derivatives that track an underlying asset but have no expiry date.
// They use a funding rate mechanism to keep the contract price close to the spot price:
//   - Positive funding rate: Longs pay shorts (perpetual trading above spot)
//   - Negative funding rate: Shorts pay longs (perpetual trading below spot)
//   - Funding is typically exchanged every 8 hours
//
// Thread Safety: This implementation is not thread-safe. Concurrent access
// should be protected by the caller.
type Future struct {
	// futureID uniquely identifies this perpetual contract
	futureID string

	// symbol is the trading symbol (e.g., "BTCUSDT", "ETHUSDT")
	symbol string

	// entryPrice is the price at which the position was entered
	entryPrice primitives.Price

	// positionSize is the size of the position (positive for long, negative for short)
	positionSize primitives.Decimal

	// leverage is the leverage multiplier used
	leverage primitives.Decimal

	// direction indicates long or short position
	direction mechanisms.PositionDirection

	// fundingPeriod is the time between funding payments (typically 8 hours)
	fundingPeriod time.Duration

	// accumulatedFunding tracks the total funding payments made/received
	accumulatedFunding primitives.Decimal

	// lastFundingTime tracks when the last funding was applied
	lastFundingTime time.Time

	// settled indicates if the position has been closed
	settled bool
}

// NewFuture creates a new perpetual futures contract.
//
// Parameters:
//   - futureID: Unique identifier for this contract
//   - symbol: Trading symbol (e.g., "BTCUSDT")
//   - entryPrice: Price at which position was entered
//   - positionSize: Size of position (positive for long, negative for short)
//   - leverage: Leverage multiplier (e.g., 10 for 10x)
//   - fundingPeriod: Time between funding payments (typically 8 hours)
//
// Returns error if any parameter is invalid.
func NewFuture(
	futureID string,
	symbol string,
	entryPrice primitives.Price,
	positionSize primitives.Decimal,
	leverage primitives.Decimal,
	fundingPeriod time.Duration,
) (*Future, error) {
	if futureID == "" {
		return nil, errors.New("futureID cannot be empty")
	}

	if symbol == "" {
		return nil, errors.New("symbol cannot be empty")
	}

	if entryPrice.IsZero() {
		return nil, errors.New("entry price must be positive")
	}

	if positionSize.IsZero() {
		return nil, ErrInvalidPositionSize
	}

	if leverage.LessThan(primitives.One()) {
		return nil, ErrInvalidLeverage
	}

	if fundingPeriod <= 0 {
		return nil, errors.New("funding period must be positive")
	}

	// Determine position direction
	direction := mechanisms.PositionDirectionLong
	if positionSize.IsNegative() {
		direction = mechanisms.PositionDirectionShort
	}

	return &Future{
		futureID:           futureID,
		symbol:             symbol,
		entryPrice:         entryPrice,
		positionSize:       positionSize,
		leverage:           leverage,
		direction:          direction,
		fundingPeriod:      fundingPeriod,
		accumulatedFunding: primitives.Zero(),
		lastFundingTime:    time.Now(),
		settled:            false,
	}, nil
}

// Mechanism returns the mechanism type identifier.
func (f *Future) Mechanism() mechanisms.MechanismType {
	return mechanisms.MechanismTypeDerivative
}

// Venue returns the venue identifier.
func (f *Future) Venue() string {
	return "perpetual"
}

// Price calculates the current value of the perpetual position.
//
// Required parameters:
//   - MarkPrice: Current mark price of the perpetual
//
// For perpetuals, the "price" is effectively the mark-to-market value,
// which depends on the difference between mark price and entry price.
//
// The value of a perpetual position is:
// Value = PositionSize * (MarkPrice - EntryPrice)
//
// However, this method returns the mark price itself as it represents
// the current trading price of the contract.
func (f *Future) Price(ctx context.Context, params mechanisms.PriceParams) (primitives.Price, error) {
	// Validate mark price
	if params.MarkPrice.IsZero() {
		return primitives.ZeroPrice(), ErrInvalidMarkPrice
	}

	// For perpetuals, return the mark price
	return params.MarkPrice, nil
}

// Greeks calculates the Greeks for the perpetual contract.
//
// For perpetuals:
//   - Delta: 1 for long positions, -1 for short positions (1:1 exposure to underlying)
//   - Gamma: 0 (linear payoff)
//   - Theta: Related to funding rate (not traditional time decay)
//   - Vega: 0 (no volatility sensitivity)
//   - Rho: 0 (no interest rate sensitivity in typical perpetuals)
func (f *Future) Greeks(ctx context.Context, params mechanisms.PriceParams) (mechanisms.Greeks, error) {
	// Delta is 1 for long, -1 for short (perpetuals have 1:1 exposure)
	var delta primitives.Decimal
	if f.direction == mechanisms.PositionDirectionLong {
		delta = primitives.NewDecimal(1)
	} else {
		delta = primitives.NewDecimal(-1)
	}

	// Theta for perpetuals is related to funding rate, not time decay
	// Funding payment = PositionValue * FundingRate
	// We'll represent this as theta for consistency
	var theta primitives.Decimal
	if !params.FundingRate.IsZero() {
		// Convert funding rate to theta (negative because it's a cost)
		// Theta = -FundingRate * PositionSize * MarkPrice
		if !params.MarkPrice.IsZero() {
			positionValue := params.MarkPrice.Decimal().Mul(f.positionSize.Abs())
			fundingPayment := positionValue.Mul(params.FundingRate)

			// For longs, positive funding is a cost (negative theta)
			// For shorts, positive funding is a gain (positive theta)
			if f.direction == mechanisms.PositionDirectionLong {
				theta = fundingPayment.Neg()
			} else {
				theta = fundingPayment
			}
		}
	}

	return mechanisms.Greeks{
		Delta: delta,
		Gamma: primitives.Zero(),
		Theta: theta,
		Vega:  primitives.Zero(),
		Rho:   primitives.Zero(),
	}, nil
}

// Settle calculates the settlement value when closing the perpetual position.
//
// Returns the P&L from the position: (MarkPrice - EntryPrice) * PositionSize - AccumulatedFunding
//
// Note: This requires the final mark price to be passed via context metadata.
func (f *Future) Settle(ctx context.Context) (primitives.Amount, error) {
	if f.settled {
		return primitives.ZeroAmount(), errors.New("position already settled")
	}

	// Extract final mark price from context metadata
	return primitives.ZeroAmount(), errors.New("settle requires final mark price in context metadata with key 'final_mark_price'")
}

// SettleWithPrice settles the position given a final mark price.
// This is a helper method that calculates final P&L including funding payments.
func (f *Future) SettleWithPrice(finalMarkPrice primitives.Price) (primitives.Amount, error) {
	if f.settled {
		return primitives.ZeroAmount(), errors.New("position already settled")
	}

	if finalMarkPrice.IsZero() {
		return primitives.ZeroAmount(), ErrInvalidMarkPrice
	}

	// Calculate price P&L: (FinalPrice - EntryPrice) * PositionSize
	priceDiff := finalMarkPrice.Decimal().Sub(f.entryPrice.Decimal())
	pricePnl := priceDiff.Mul(f.positionSize)

	// Subtract accumulated funding (funding is a cost/benefit separate from price movement)
	totalPnl := pricePnl.Sub(f.accumulatedFunding)

	f.settled = true

	// Return absolute value as Amount (sign indicates profit/loss)
	if totalPnl.IsNegative() {
		return primitives.NewAmount(totalPnl.Neg())
	}
	return primitives.NewAmount(totalPnl)
}

// ApplyFunding applies funding rate payments to the position.
//
// Funding payment is calculated as:
// Payment = PositionSize * MarkPrice * FundingRate
//
// Positive funding rate:
//   - Longs pay shorts (payment is positive for longs, negative for shorts)
//   - Negative funding rate:
//   - Shorts pay longs (payment is negative for longs, positive for shorts)
//
// Parameters:
//   - markPrice: Current mark price
//   - fundingRate: Funding rate for this period (as decimal, e.g., 0.0001 for 0.01%)
//
// Returns the funding payment amount.
func (f *Future) ApplyFunding(markPrice primitives.Price, fundingRate primitives.Decimal) (primitives.Decimal, error) {
	if markPrice.IsZero() {
		return primitives.Zero(), ErrInvalidMarkPrice
	}

	// Calculate funding payment
	// Payment = |PositionSize| * MarkPrice * FundingRate
	positionValue := f.positionSize.Abs().Mul(markPrice.Decimal())
	fundingPayment := positionValue.Mul(fundingRate)

	// For longs, positive funding is a payment (cost)
	// For shorts, positive funding is a receipt (benefit)
	var payment primitives.Decimal
	if f.direction == mechanisms.PositionDirectionLong {
		payment = fundingPayment
	} else {
		payment = fundingPayment.Neg()
	}

	// Accumulate funding
	f.accumulatedFunding = f.accumulatedFunding.Add(payment)
	f.lastFundingTime = time.Now()

	return payment, nil
}

// CalculateFundingRate calculates the funding rate based on mark and index prices.
//
// The funding rate is typically calculated as:
// FundingRate = (MarkPrice - IndexPrice) / IndexPrice * FundingRateMultiplier
//
// Where FundingRateMultiplier is typically around 1/3 for 8-hour periods (to normalize to 8-hour periods).
//
// Parameters:
//   - markPrice: Current mark price of the perpetual
//   - indexPrice: Current index (spot) price
//   - multiplier: Funding rate multiplier (typically 0.333 for 8-hour periods)
//
// Returns the calculated funding rate.
func CalculateFundingRate(
	markPrice primitives.Price,
	indexPrice primitives.Price,
	multiplier primitives.Decimal,
) (primitives.Decimal, error) {
	if indexPrice.IsZero() {
		return primitives.Zero(), ErrInvalidIndexPrice
	}

	// Calculate premium: (MarkPrice - IndexPrice) / IndexPrice
	premium := markPrice.Decimal().Sub(indexPrice.Decimal())
	premiumRate, err := premium.Div(indexPrice.Decimal())
	if err != nil {
		return primitives.Zero(), err
	}

	// Apply multiplier
	fundingRate := premiumRate.Mul(multiplier)

	return fundingRate, nil
}

// UnrealizedPnL calculates the unrealized P&L of the position.
//
// UnrealizedPnL = (CurrentMarkPrice - EntryPrice) * PositionSize - AccumulatedFunding
func (f *Future) UnrealizedPnL(currentMarkPrice primitives.Price) (primitives.Decimal, error) {
	if currentMarkPrice.IsZero() {
		return primitives.Zero(), ErrInvalidMarkPrice
	}

	// Calculate price P&L
	priceDiff := currentMarkPrice.Decimal().Sub(f.entryPrice.Decimal())
	pricePnl := priceDiff.Mul(f.positionSize)

	// Subtract accumulated funding
	totalPnl := pricePnl.Sub(f.accumulatedFunding)

	return totalPnl, nil
}

// Liquidation Price calculates the price at which the position would be liquidated.
//
// Liquidation occurs when losses exceed the margin (initial capital / leverage).
// LiquidationPrice = EntryPrice * (1 - 1/Leverage) for longs
// LiquidationPrice = EntryPrice * (1 + 1/Leverage) for shorts
//
// This is a simplified calculation that doesn't account for funding or maintenance margin.
func (f *Future) LiquidationPrice() (primitives.Price, error) {
	// Calculate liquidation distance: 1 / leverage
	one := primitives.NewDecimal(1)
	liquidationDistance, err := one.Div(f.leverage)
	if err != nil {
		return primitives.ZeroPrice(), err
	}

	// Calculate liquidation price based on direction
	entryPriceDec := f.entryPrice.Decimal()
	var liquidationPrice primitives.Decimal

	if f.direction == mechanisms.PositionDirectionLong {
		// For longs: EntryPrice * (1 - 1/Leverage)
		multiplier := one.Sub(liquidationDistance)
		liquidationPrice = entryPriceDec.Mul(multiplier)
	} else {
		// For shorts: EntryPrice * (1 + 1/Leverage)
		multiplier := one.Add(liquidationDistance)
		liquidationPrice = entryPriceDec.Mul(multiplier)
	}

	return primitives.NewPrice(liquidationPrice)
}

// FutureID returns the future contract identifier.
func (f *Future) FutureID() string {
	return f.futureID
}

// Symbol returns the trading symbol.
func (f *Future) Symbol() string {
	return f.symbol
}

// EntryPrice returns the entry price.
func (f *Future) EntryPrice() primitives.Price {
	return f.entryPrice
}

// PositionSize returns the position size.
func (f *Future) PositionSize() primitives.Decimal {
	return f.positionSize
}

// Leverage returns the leverage multiplier.
func (f *Future) Leverage() primitives.Decimal {
	return f.leverage
}

// Direction returns the position direction.
func (f *Future) Direction() mechanisms.PositionDirection {
	return f.direction
}

// AccumulatedFunding returns the total accumulated funding payments.
func (f *Future) AccumulatedFunding() primitives.Decimal {
	return f.accumulatedFunding
}

// IsSettled returns whether the position has been settled.
func (f *Future) IsSettled() bool {
	return f.settled
}
