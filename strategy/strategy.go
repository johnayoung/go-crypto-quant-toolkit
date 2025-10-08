// Package strategy provides the strategy framework for portfolio management
// and strategy coordination. It defines interfaces for implementing trading
// strategies that can compose multiple market mechanisms without coupling
// to specific implementations.
package strategy

import "context"

// Strategy makes portfolio rebalancing decisions based on market state.
// Implementations define the logic for when and how to adjust positions.
//
// The Strategy interface is intentionally minimal, allowing for diverse
// strategy implementations ranging from simple buy-and-hold to complex
// multi-venue arbitrage strategies.
//
// Thread Safety: Strategy implementations are not required to be thread-safe.
// The backtest engine will call Rebalance sequentially for each market event.
type Strategy interface {
	// Rebalance is called on each market data event and returns the desired
	// position changes to apply to the portfolio.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - portfolio: Current portfolio state (read-only view)
	//   - snapshot: Current market data snapshot
	//
	// Returns:
	//   - []Action: List of actions to apply to the portfolio (may be empty)
	//   - error: Any error encountered during rebalancing logic
	//
	// Error Handling:
	//   - Return error for unrecoverable issues (invalid state, missing data)
	//   - Empty action list indicates no changes needed
	//   - Actions are applied in order; partial application on error is allowed
	//
	// Performance: This method may be called frequently (every market tick).
	// Implementations should avoid expensive operations unless necessary.
	Rebalance(ctx context.Context, portfolio *Portfolio, snapshot MarketSnapshot) ([]Action, error)
}
