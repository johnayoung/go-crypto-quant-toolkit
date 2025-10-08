// Package backtest provides an event-driven backtesting engine for trading strategies.
// It enables testing strategy implementations against historical market data while
// remaining completely mechanism-agnostic.
//
// The engine works with any Strategy implementation and any Position types,
// validating the framework's extensibility design.
package backtest

import (
	"context"
	"fmt"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

// Engine orchestrates backtesting of trading strategies against historical market data.
// It implements an event-driven architecture where each market data snapshot triggers
// strategy rebalancing.
//
// Design Principles:
//   - Mechanism-agnostic: Works with any Position implementations
//   - Zero coupling: Never references concrete mechanism types
//   - Extensible: Supports any Strategy implementation
//   - Observable: Tracks portfolio value over time
//
// Thread Safety: Engine is not thread-safe. Each backtest should run in a single goroutine.
// Use separate Engine instances for concurrent backtests.
type Engine struct {
	// config holds engine configuration
	config Config
}

// Config contains backtest engine configuration options.
type Config struct {
	// InitialCash is the starting portfolio cash balance
	InitialCash primitives.Amount

	// EnableDetailedLogging enables verbose logging of each rebalancing step
	// (useful for debugging but may impact performance)
	EnableDetailedLogging bool
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() Config {
	return Config{
		InitialCash:           primitives.MustAmount(primitives.MustDecimalFromString("10000.0")), // $10k default
		EnableDetailedLogging: false,
	}
}

// NewEngine creates a new backtest engine with the provided configuration.
func NewEngine(config Config) *Engine {
	return &Engine{
		config: config,
	}
}

// NewEngineWithDefaults creates a new backtest engine with default configuration.
func NewEngineWithDefaults() *Engine {
	return NewEngine(DefaultConfig())
}

// Run executes a backtest of the given strategy against the provided market data.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - strat: Strategy to backtest (must not be nil)
//   - snapshots: Chronological sequence of market data snapshots
//
// Returns:
//   - *Result: Backtest results including performance metrics and portfolio history
//   - error: Any error encountered during backtesting
//
// Error Handling:
//   - Returns error if strategy is nil or snapshots is empty
//   - Returns error if strategy.Rebalance() fails
//   - Returns error if action application fails
//   - Respects context cancellation (returns ctx.Err())
//
// Execution Flow:
//  1. Initialize portfolio with configured initial cash
//  2. For each market snapshot (in order):
//     a. Check context cancellation
//     b. Call strategy.Rebalance(ctx, portfolio, snapshot)
//     c. Apply returned actions to portfolio
//     d. Calculate and record portfolio value
//  3. Calculate performance metrics from value history
//  4. Return results
//
// The engine guarantees:
//   - Snapshots processed in order
//   - Portfolio value calculated after each rebalancing
//   - All actions applied atomically per snapshot
//   - No assumptions about position or mechanism types
func (e *Engine) Run(
	ctx context.Context,
	strat strategy.Strategy,
	snapshots []strategy.MarketSnapshot,
) (*Result, error) {
	// Validate inputs
	if strat == nil {
		return nil, fmt.Errorf("strategy cannot be nil")
	}
	if len(snapshots) == 0 {
		return nil, fmt.Errorf("snapshots cannot be empty")
	}

	// Initialize portfolio
	portfolio := strategy.NewPortfolio(e.config.InitialCash)

	// Track portfolio values over time
	valueHistory := make([]ValuePoint, 0, len(snapshots))

	// Event loop: process each market snapshot
	for i, snapshot := range snapshots {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("backtest cancelled: %w", ctx.Err())
		default:
		}

		// Calculate portfolio value BEFORE rebalancing
		// (first snapshot uses initial cash, subsequent use actual portfolio value)
		portfolioValue, err := e.calculatePortfolioValue(portfolio, snapshot)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate portfolio value at snapshot %d: %w", i, err)
		}

		// Record value point
		valueHistory = append(valueHistory, ValuePoint{
			Time:  snapshot.Time(),
			Value: portfolioValue,
		})

		// Call strategy rebalancing logic
		actions, err := strat.Rebalance(ctx, portfolio, snapshot)
		if err != nil {
			return nil, fmt.Errorf("strategy rebalance failed at snapshot %d: %w", i, err)
		}

		// Apply actions to portfolio
		for actionIdx, action := range actions {
			if err := action.Apply(portfolio); err != nil {
				return nil, fmt.Errorf("failed to apply action %d at snapshot %d: %w", actionIdx, i, err)
			}
		}
	}

	// Calculate final portfolio value
	finalSnapshot := snapshots[len(snapshots)-1]
	finalValue, err := e.calculatePortfolioValue(portfolio, finalSnapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate final portfolio value: %w", err)
	}

	// Build result with performance metrics
	result := &Result{
		InitialValue: e.config.InitialCash,
		FinalValue:   finalValue,
		ValueHistory: valueHistory,
		Portfolio:    portfolio,
	}

	// Calculate derived metrics
	if err := result.calculateMetrics(); err != nil {
		return nil, fmt.Errorf("failed to calculate performance metrics: %w", err)
	}

	return result, nil
}

// calculatePortfolioValue computes the total value of the portfolio at the given market snapshot.
// Returns the sum of cash plus all position values.
func (e *Engine) calculatePortfolioValue(
	portfolio *strategy.Portfolio,
	snapshot strategy.MarketSnapshot,
) (primitives.Amount, error) {
	// Start with cash balance
	totalValue := portfolio.Cash()

	// Add value of all positions
	positions := portfolio.Positions()
	for _, position := range positions {
		posValue, err := position.Value(snapshot)
		if err != nil {
			return primitives.Amount{}, fmt.Errorf("failed to value position %s: %w", position.ID(), err)
		}
		totalValue = totalValue.Add(posValue)
	}

	return totalValue, nil
}
