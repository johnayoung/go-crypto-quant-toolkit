// Package main demonstrates a simple liquidity provision strategy using
// a basic constant-product AMM. This example shows how to:
//  1. Create a simple AMM pool mechanism
//  2. Implement a basic LP strategy
//  3. Run a backtest with historical price data
//  4. Analyze backtest results
//
// This serves as a template for building more complex LP strategies.
// Note: This uses simple constant-product math (x*y=k) for clarity.
// For concentrated liquidity, see the implementations/concentrated_liquidity package.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/backtest"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

// SimpleAMM implements a basic constant-product AMM (x * y = k).
// This is used for demonstration purposes with transparent, understandable math.
type SimpleAMM struct {
	poolID   string
	feeRate  primitives.Decimal // e.g., 0.003 for 0.3%
	reserveA primitives.Amount  // ETH reserve
	reserveB primitives.Amount  // USDC reserve
	k        primitives.Decimal // Constant product (x * y = k)
}

// NewSimpleAMM creates a new simple constant-product AMM.
func NewSimpleAMM(poolID string, feeRate primitives.Decimal, reserveA, reserveB primitives.Amount) *SimpleAMM {
	// Calculate k = x * y
	k := reserveA.Decimal().Mul(reserveB.Decimal())

	return &SimpleAMM{
		poolID:   poolID,
		feeRate:  feeRate,
		reserveA: reserveA,
		reserveB: reserveB,
		k:        k,
	}
}

// LPPosition wraps a mechanisms.PoolPosition to implement strategy.Position interface.
type LPPosition struct {
	poolPosition mechanisms.PoolPosition
	amm          *SimpleAMM
}

// NewLPPosition creates a new LP position wrapper.
func NewLPPosition(poolPos mechanisms.PoolPosition, amm *SimpleAMM) *LPPosition {
	return &LPPosition{
		poolPosition: poolPos,
		amm:          amm,
	}
}

// ID returns the unique identifier for this position.
func (lp *LPPosition) ID() string {
	return lp.poolPosition.PoolID
}

// Type returns the position type classification.
func (lp *LPPosition) Type() strategy.PositionType {
	return strategy.PositionTypeLiquidityPool
}

// Value calculates the current value of the LP position.
// Uses simple token amounts from the deposited position.
func (lp *LPPosition) Value(snapshot strategy.MarketSnapshot) (primitives.Amount, error) {
	// Get current token amounts (for simple AMM, these stay constant)
	amounts := lp.poolPosition.TokensDeposited

	// Get ETH price
	ethPrice, err := snapshot.Price("ETH/USD")
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to get ETH price: %w", err)
	}

	// USDC is $1
	usdcPrice := primitives.MustPrice(primitives.One())

	// Calculate total value
	valueETH := amounts.AmountA.MulPrice(ethPrice)
	valueUSDC := amounts.AmountB.MulPrice(usdcPrice)

	return valueETH.Add(valueUSDC), nil
}

// SimpleLPStrategy implements a passive liquidity provision strategy.
// It provides liquidity to a simple AMM pool and holds the position throughout the backtest.
type SimpleLPStrategy struct {
	amm         *SimpleAMM
	hasPosition bool
	depositETH  primitives.Amount
	depositUSDC primitives.Amount
}

// NewSimpleLPStrategy creates a new simple LP strategy.
func NewSimpleLPStrategy(amm *SimpleAMM, depositETH, depositUSDC primitives.Amount) *SimpleLPStrategy {
	return &SimpleLPStrategy{
		amm:         amm,
		hasPosition: false,
		depositETH:  depositETH,
		depositUSDC: depositUSDC,
	}
}

// Rebalance implements strategy.Strategy.
// On the first call, it adds a liquidity position. Subsequent calls do nothing (passive hold).
func (s *SimpleLPStrategy) Rebalance(
	ctx context.Context,
	portfolio *strategy.Portfolio,
	snapshot strategy.MarketSnapshot,
) ([]strategy.Action, error) {
	// Only add position on first rebalance
	if s.hasPosition {
		return nil, nil
	}

	// Create a mock pool position for demonstration
	// In a real implementation, this would come from AddLiquidity

	// Create a position with our token deposits
	poolPosition := mechanisms.PoolPosition{
		PoolID:    s.amm.poolID,
		Liquidity: s.depositETH, // Using depositETH as proxy for liquidity
		TokensDeposited: mechanisms.TokenAmounts{
			AmountA: s.depositETH,
			AmountB: s.depositUSDC,
		},
		Metadata: map[string]interface{}{
			"pool_type": "constant_product",
		},
	}

	// Wrap in our position adapter
	lpPos := NewLPPosition(poolPosition, s.amm)

	s.hasPosition = true

	// Calculate capital to deduct
	ethPrice, err := snapshot.Price("ETH/USD")
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH price: %w", err)
	}

	capitalRequired := s.depositETH.MulPrice(ethPrice).Add(s.depositUSDC)

	// Return action to add the position
	return []strategy.Action{
		strategy.NewAddPositionAction(lpPos),
		strategy.NewAdjustCashAction(
			capitalRequired.Decimal().Neg(),
			"capital for LP position",
		),
	}, nil
}

// createHistoricalSnapshots generates mock historical market data for backtesting.
// In a production system, this would load real historical data.
func createHistoricalSnapshots() []strategy.MarketSnapshot {
	snapshots := make([]strategy.MarketSnapshot, 0)
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Simulate 30 days of data with varying ETH price
	for day := 0; day < 30; day++ {
		t := primitives.NewTime(startTime.Add(time.Duration(day) * 24 * time.Hour))

		// ETH price oscillates between 1800 and 2200
		// Using simple sine wave for variation
		basePrice := 2000.0
		variation := 200.0 * (float64(day%10) / 10.0)
		if day%20 >= 10 {
			variation = -variation
		}
		ethPrice := basePrice + variation

		// Create snapshot with just prices (simple AMM doesn't need complex metadata)
		prices := map[string]primitives.Price{
			"ETH/USD": primitives.MustPrice(primitives.MustDecimalFromString(fmt.Sprintf("%.2f", ethPrice))),
		}

		snapshot := strategy.NewSimpleSnapshot(t, prices)
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

func main() {
	fmt.Println("=== Simple LP Strategy Backtest ===")
	fmt.Println()

	// 1. Create a simple constant-product AMM pool
	// Using ETH/USDC with 0.3% fee
	initialReserveETH := primitives.MustAmount(primitives.NewDecimal(100))     // 100 ETH
	initialReserveUSDC := primitives.MustAmount(primitives.NewDecimal(200000)) // 200k USDC
	feeRate := primitives.NewDecimalFromFloat(0.003)                           // 0.3%

	amm := NewSimpleAMM("eth-usdc-pool", feeRate, initialReserveETH, initialReserveUSDC)
	fmt.Printf("Created simple AMM pool with:\n")
	fmt.Printf("  - Initial reserves: %s ETH, %s USDC\n", initialReserveETH.String(), initialReserveUSDC.String())
	fmt.Printf("  - Fee rate: 0.3%%\n")
	fmt.Printf("  - Constant k: %s\n\n", amm.k.String())

	// 2. Define deposit amounts for our LP position
	depositETH := primitives.MustAmount(primitives.NewDecimal(5))      // 5 ETH
	depositUSDC := primitives.MustAmount(primitives.NewDecimal(10000)) // 10k USDC

	// 3. Create strategy
	strat := NewSimpleLPStrategy(amm, depositETH, depositUSDC)
	fmt.Printf("Strategy will deposit: %s ETH + %s USDC\n\n", depositETH.String(), depositUSDC.String())

	// 4. Create historical market snapshots
	snapshots := createHistoricalSnapshots()
	fmt.Printf("Generated %d days of historical data\n", len(snapshots))

	// 5. Configure backtest engine
	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(100000)), // $100k starting capital
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	// 6. Run backtest
	fmt.Println("Running backtest...")
	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	// 7. Display results
	fmt.Println("\n" + result.Summary())

	// Additional analysis
	fmt.Println("\n=== Position Analysis ===")
	positions := result.Portfolio.Positions()
	fmt.Printf("Final positions: %d\n", len(positions))
	for _, pos := range positions {
		fmt.Printf("  - Position ID: %s\n", pos.ID())
		fmt.Printf("    Type: %s\n", pos.Type())
		if len(result.ValueHistory) > 0 {
			lastSnapshot := snapshots[len(snapshots)-1]
			value, err := pos.Value(lastSnapshot)
			if err != nil {
				fmt.Printf("    Value: Error calculating - %v\n", err)
			} else {
				fmt.Printf("    Final Value: %s\n", value.String())
			}
		}
	}

	fmt.Println("\n=== Conclusion ===")
	if result.TotalReturn.IsPositive() {
		fmt.Println("✓ Strategy generated positive returns")
	} else {
		fmt.Println("✗ Strategy experienced losses")
	}
	fmt.Printf("This demonstrates basic LP position management and backtesting.\n")
	fmt.Printf("A production strategy would include rebalancing logic, fee collection,\n")
	fmt.Printf("and risk management based on market conditions.\n")
}
