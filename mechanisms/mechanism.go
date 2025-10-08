// Package mechanisms provides extensible interface contracts for market mechanisms.
// These interfaces define how different trading venues and protocols can be integrated
// into the framework without modifying core abstractions. Users can add new mechanism
// types by implementing these interfaces.
package mechanisms

// MechanismType identifies the category of market mechanism.
// Common types include liquidity pools, derivatives, order books, and more.
// Users can define custom mechanism types for novel protocols.
type MechanismType string

const (
	// MechanismTypeLiquidityPool represents AMM-style liquidity pools
	// (e.g., Uniswap, Curve, Balancer)
	MechanismTypeLiquidityPool MechanismType = "liquidity_pool"

	// MechanismTypeDerivative represents derivative instruments
	// (e.g., options, perpetuals, futures)
	MechanismTypeDerivative MechanismType = "derivative"

	// MechanismTypeOrderBook represents order book based trading
	// (e.g., CEX-style limit order books)
	MechanismTypeOrderBook MechanismType = "orderbook"

	// Additional types can be defined as needed:
	// MechanismTypeBatchAuction, MechanismTypeFlashLoan,
	// MechanismTypeIntentPool, MechanismTypeBridge, etc.
)

// MarketMechanism is the base interface that all market mechanisms must implement.
// It provides identification and context about where the mechanism exists.
//
// Implementations should embed this interface when defining specific mechanism
// categories (e.g., LiquidityPool, Derivative, OrderBook).
//
// Thread Safety: Implementations are not required to be thread-safe by default.
// Concurrent access should be protected by the caller if needed.
type MarketMechanism interface {
	// Mechanism returns the type of market mechanism this implements.
	// This allows type-safe casting and routing logic based on mechanism category.
	Mechanism() MechanismType

	// Venue returns an identifier for where this mechanism exists.
	// Examples: "uniswap-v3", "gmx", "dydx", "binance"
	//
	// Optional: can return empty string if venue identification is not relevant.
	// Useful for strategies that interact with multiple venues or need
	// venue-specific logic.
	Venue() string
}
