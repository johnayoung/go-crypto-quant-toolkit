// Package strategy provides the strategy framework for portfolio management
// and strategy coordination. It defines interfaces for implementing trading
// strategies that can compose multiple market mechanisms without coupling
// to specific implementations.
package strategy

import (
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// MarketSnapshot provides a point-in-time view of market data.
// It abstracts market data sources, allowing strategies to work with
// any data format without coupling to specific providers.
//
// Implementations should provide immutable snapshots representing market
// state at a specific time. This enables consistent decision-making and
// position valuation within a single strategy invocation.
//
// Thread Safety: MarketSnapshot implementations should be safe for concurrent
// reads from multiple goroutines, but are typically accessed sequentially
// within a single strategy execution.
type MarketSnapshot interface {
	// Time returns the timestamp of this market snapshot.
	// All data in the snapshot should be consistent as of this time.
	Time() primitives.Time

	// Price returns the current price for the given asset pair.
	// The pair format is implementation-specific but should be consistent
	// (e.g., "ETH/USDC", "ETH-USDC", "ETHUSDC").
	//
	// Returns error if the price is not available for the given pair.
	Price(pair string) (primitives.Price, error)

	// Prices returns all available prices in this snapshot.
	// Useful for strategies that need to iterate over all available markets.
	//
	// The returned map should not be modified by the caller.
	Prices() map[string]primitives.Price

	// Get retrieves custom metadata from the snapshot.
	// This allows implementations to provide mechanism-specific data
	// (e.g., liquidity depth, funding rates, volatility surfaces) without
	// extending the interface.
	//
	// Returns false if the key doesn't exist.
	//
	// Example:
	//   liquidity, ok := snapshot.Get("uniswap-v3:ETH/USDC:liquidity")
	//   fundingRate, ok := snapshot.Get("perpetual:ETH-PERP:funding_rate")
	Get(key string) (interface{}, bool)
}

// SimpleSnapshot provides a basic implementation of MarketSnapshot
// backed by an in-memory map. Useful for testing and simple strategies.
type SimpleSnapshot struct {
	time   primitives.Time
	prices map[string]primitives.Price
	data   map[string]interface{}
}

// NewSimpleSnapshot creates a new SimpleSnapshot with the given time and prices.
func NewSimpleSnapshot(time primitives.Time, prices map[string]primitives.Price) *SimpleSnapshot {
	return &SimpleSnapshot{
		time:   time,
		prices: prices,
		data:   make(map[string]interface{}),
	}
}

// Time returns the timestamp of this snapshot.
func (s *SimpleSnapshot) Time() primitives.Time {
	return s.time
}

// Price returns the price for the given pair.
func (s *SimpleSnapshot) Price(pair string) (primitives.Price, error) {
	price, ok := s.prices[pair]
	if !ok {
		return primitives.Price{}, ErrPriceNotAvailable
	}
	return price, nil
}

// Prices returns all available prices in this snapshot.
func (s *SimpleSnapshot) Prices() map[string]primitives.Price {
	return s.prices
}

// Get retrieves custom metadata from the snapshot.
func (s *SimpleSnapshot) Get(key string) (interface{}, bool) {
	val, ok := s.data[key]
	return val, ok
}

// Set stores custom metadata in the snapshot.
// This method is provided for test and setup purposes.
func (s *SimpleSnapshot) Set(key string, value interface{}) {
	s.data[key] = value
}
