package strategy

import (
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// PositionType classifies the type of position held in a portfolio.
// This allows for type-specific logic and reporting without coupling
// to concrete position implementations.
type PositionType string

const (
	// PositionTypeSpot represents a spot holding of a single asset
	PositionTypeSpot PositionType = "spot"

	// PositionTypeLiquidityPool represents a liquidity pool position (LP)
	PositionTypeLiquidityPool PositionType = "liquidity_pool"

	// PositionTypeOption represents an options contract
	PositionTypeOption PositionType = "option"

	// PositionTypePerpetual represents a perpetual futures contract
	PositionTypePerpetual PositionType = "perpetual"

	// PositionTypeFuture represents a dated futures contract
	PositionTypeFuture PositionType = "future"

	// PositionTypeOrderBook represents an active order book position
	PositionTypeOrderBook PositionType = "orderbook"

	// Additional position types can be defined as needed:
	// PositionTypeLending, PositionTypeBorrowing, PositionTypeStaked, etc.
)

// Position represents any tradeable position in a portfolio.
// This interface is intentionally minimal to support diverse position types
// while providing the essential methods needed for portfolio management.
//
// Implementations should be immutable snapshots of position state at a point
// in time. Position modifications should create new Position instances rather
// than mutating existing ones.
//
// Thread Safety: Position implementations should be safe for concurrent reads.
type Position interface {
	// Value returns the current value of the position in the portfolio's
	// denomination currency, using prices from the provided market snapshot.
	//
	// For spot positions: quantity * price
	// For LP positions: current value of tokens + fees - impermanent loss
	// For derivatives: notional * price + unrealized PnL
	// For orders: expected value if filled
	//
	// Returns error if required prices are not available in the snapshot
	// or if position state is invalid.
	Value(snapshot MarketSnapshot) (primitives.Amount, error)

	// Type returns the classification of this position.
	// Used for reporting, risk analysis, and type-specific logic.
	Type() PositionType

	// ID returns a unique identifier for this position within the portfolio.
	// The ID format is implementation-specific but must be unique and stable.
	//
	// Example formats:
	//   - "spot:ETH"
	//   - "lp:uniswap-v3:ETH/USDC:0x123..."
	//   - "option:deribit:ETH-25000-C"
	//   - "perp:gmx:ETH-USD"
	ID() string
}

// RiskMetrics contains position-specific risk measures.
// Implementations may provide different metrics depending on position type.
//
// All monetary risk values (e.g., VaR, expected shortfall) should use
// the portfolio's denomination currency.
type RiskMetrics struct {
	// Delta measures the position's price sensitivity to the underlying asset.
	// For spot: always 1.0
	// For derivatives: delta from Greeks calculation
	// For LP: effective delta considering both tokens
	Delta primitives.Decimal

	// Gamma measures the rate of change of delta.
	// Only applicable for derivatives; zero for spot and LP positions.
	Gamma primitives.Decimal

	// Vega measures sensitivity to implied volatility changes.
	// Only applicable for options; zero for other position types.
	Vega primitives.Decimal

	// Theta measures time decay (value change per day).
	// Primarily for options, but can apply to funding-rate positions.
	Theta primitives.Decimal

	// Leverage indicates the position's leverage ratio.
	// 1.0 for spot, >1.0 for leveraged positions.
	Leverage primitives.Decimal

	// Liquidation indicates the price level at which position would be liquidated.
	// Zero if position has no liquidation risk (e.g., spot with no borrowed funds).
	LiquidationPrice primitives.Price

	// Additional risk metrics can be added as needed:
	// - VaR (Value at Risk)
	// - Expected Shortfall
	// - Beta (correlation to market)
	// - Concentration risk
	// Store in a map[string]interface{} field if needed for extensibility
}

// PositionWithRisk is an optional interface positions can implement to provide
// risk metrics. If a position doesn't implement this interface, default risk
// metrics should be assumed by the portfolio or strategy.
type PositionWithRisk interface {
	Position

	// Risk returns position-specific risk metrics using current market data.
	// Returns error if required data is unavailable or calculation fails.
	Risk(snapshot MarketSnapshot) (RiskMetrics, error)
}

// PositionMetadata provides optional descriptive information about a position.
// Useful for logging, debugging, and user interfaces.
type PositionMetadata interface {
	Position

	// Description returns a human-readable description of the position.
	// Example: "100 ETH spot", "ETH/USDC LP 1.5-2.0x range", "ETH Call $2500 exp 2024-12-31"
	Description() string

	// Venue returns the venue/protocol where this position exists.
	// Example: "uniswap-v3", "gmx", "deribit", "binance"
	Venue() string
}
