// Package main demonstrates how to add a custom market mechanism without modifying
// the framework code. This example implements a simple constant-product AMM
// (Uniswap V2 style) and uses it in a strategy, proving the framework's extensibility.
//
// Key demonstration:
//  1. Implement custom mechanism (ConstantProductPool)
//  2. Create position wrapper for the custom mechanism
//  3. Use it in a strategy alongside framework's backtest engine
//  4. Zero modifications to framework code required
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

// ====================================================================
// CUSTOM MECHANISM IMPLEMENTATION
// ====================================================================

// ConstantProductPool implements a simple constant-product AMM (x * y = k).
// This is a custom mechanism implementation demonstrating framework extensibility.
//
// The constant product formula: reserveA * reserveB = k (constant)
// Price = reserveB / reserveA
//
// This type implements mechanisms.LiquidityPool without modifying framework code.
type ConstantProductPool struct {
	poolID       string
	tokenASymbol string
	tokenBSymbol string
	feeRate      primitives.Decimal // e.g., 0.003 for 0.3%
}

// NewConstantProductPool creates a new constant-product AMM pool.
func NewConstantProductPool(poolID, tokenA, tokenB string, feeRate primitives.Decimal) *ConstantProductPool {
	return &ConstantProductPool{
		poolID:       poolID,
		tokenASymbol: tokenA,
		tokenBSymbol: tokenB,
		feeRate:      feeRate,
	}
}

// Mechanism returns the mechanism type (implements mechanisms.MarketMechanism).
func (p *ConstantProductPool) Mechanism() mechanisms.MechanismType {
	return mechanisms.MechanismTypeLiquidityPool
}

// Venue returns the venue identifier.
func (p *ConstantProductPool) Venue() string {
	return "custom-amm"
}

// Calculate computes pool state using constant product formula (implements mechanisms.LiquidityPool).
func (p *ConstantProductPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
	if params.ReserveA.IsZero() || params.ReserveB.IsZero() {
		return mechanisms.PoolState{}, fmt.Errorf("reserves cannot be zero")
	}

	// Spot price = reserveB / reserveA
	priceDecimal, err := params.ReserveB.Decimal().Div(params.ReserveA.Decimal())
	if err != nil {
		return mechanisms.PoolState{}, fmt.Errorf("failed to calculate price: %w", err)
	}

	spotPrice := primitives.MustPrice(priceDecimal)

	// Liquidity = geometric mean approximation using min(reserveA, reserveB) for simplicity
	// In a real implementation, use sqrt(reserveA * reserveB)
	liquidity := params.ReserveA
	if params.ReserveB.LessThan(params.ReserveA) {
		liquidity = params.ReserveB
	}

	return mechanisms.PoolState{
		SpotPrice:          spotPrice,
		Liquidity:          liquidity,
		EffectiveLiquidity: liquidity,
		AccumulatedFeesA:   primitives.ZeroAmount(),
		AccumulatedFeesB:   primitives.ZeroAmount(),
		Metadata:           make(map[string]interface{}),
	}, nil
}

// AddLiquidity simulates adding liquidity to the pool.
func (p *ConstantProductPool) AddLiquidity(ctx context.Context, amounts mechanisms.TokenAmounts) (mechanisms.PoolPosition, error) {
	// In a real implementation, this would:
	// 1. Calculate optimal token ratio
	// 2. Determine LP tokens minted
	// 3. Return position with appropriate metadata

	// For this example, we create a simplified position
	// Liquidity = min(amountA, amountB) for simplicity
	liquidity := amounts.AmountA
	if amounts.AmountB.LessThan(amounts.AmountA) {
		liquidity = amounts.AmountB
	}

	return mechanisms.PoolPosition{
		PoolID:          p.poolID,
		Liquidity:       liquidity,
		TokensDeposited: amounts,
		Metadata: map[string]interface{}{
			"pool_type": "constant_product",
		},
	}, nil
}

// RemoveLiquidity simulates removing liquidity from the pool.
func (p *ConstantProductPool) RemoveLiquidity(ctx context.Context, position mechanisms.PoolPosition) (mechanisms.TokenAmounts, error) {
	// For simplicity, return the deposited amounts
	// In a real implementation, this would:
	// 1. Calculate current pool state
	// 2. Determine token amounts based on current reserves and LP share
	// 3. Account for fees earned
	// 4. Account for impermanent loss/gain

	return position.TokensDeposited, nil
}

// ====================================================================
// POSITION WRAPPER
// ====================================================================

// CustomAMMPosition wraps our custom pool position to implement strategy.Position.
type CustomAMMPosition struct {
	poolPosition mechanisms.PoolPosition
	pool         *ConstantProductPool
}

// NewCustomAMMPosition creates a position wrapper for our custom AMM.
func NewCustomAMMPosition(poolPos mechanisms.PoolPosition, pool *ConstantProductPool) *CustomAMMPosition {
	return &CustomAMMPosition{
		poolPosition: poolPos,
		pool:         pool,
	}
}

func (cap *CustomAMMPosition) ID() string {
	return cap.poolPosition.PoolID
}

func (cap *CustomAMMPosition) Type() strategy.PositionType {
	return strategy.PositionTypeLiquidityPool
}

func (cap *CustomAMMPosition) Value(snapshot strategy.MarketSnapshot) (primitives.Amount, error) {
	// Get current token amounts from the position
	amounts, err := cap.pool.RemoveLiquidity(context.Background(), cap.poolPosition)
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to calculate position value: %w", err)
	}

	// Get prices
	// For this example, assume token A is ETH and token B is USDC
	tokenAPrice, err := snapshot.Price("ETH/USD")
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to get token A price: %w", err)
	}

	tokenBPrice := primitives.MustPrice(primitives.One()) // USDC = $1

	// Calculate total value
	valueA := amounts.AmountA.MulPrice(tokenAPrice)
	valueB := amounts.AmountB.MulPrice(tokenBPrice)

	return valueA.Add(valueB), nil
}

// ====================================================================
// STRATEGY USING CUSTOM MECHANISM
// ====================================================================

// CustomAMMStrategy demonstrates using a custom mechanism in a strategy.
type CustomAMMStrategy struct {
	pool            *ConstantProductPool
	hasPosition     bool
	initialDepositA primitives.Amount
	initialDepositB primitives.Amount
}

// NewCustomAMMStrategy creates a strategy using our custom AMM.
func NewCustomAMMStrategy(pool *ConstantProductPool, depositA, depositB primitives.Amount) *CustomAMMStrategy {
	return &CustomAMMStrategy{
		pool:            pool,
		hasPosition:     false,
		initialDepositA: depositA,
		initialDepositB: depositB,
	}
}

// Rebalance implements strategy.Strategy.
func (s *CustomAMMStrategy) Rebalance(
	ctx context.Context,
	portfolio *strategy.Portfolio,
	snapshot strategy.MarketSnapshot,
) ([]strategy.Action, error) {
	// On first call, add liquidity to the custom AMM
	if !s.hasPosition {
		// Create position using our custom AMM
		amounts := mechanisms.TokenAmounts{
			AmountA: s.initialDepositA,
			AmountB: s.initialDepositB,
		}

		poolPosition, err := s.pool.AddLiquidity(ctx, amounts)
		if err != nil {
			return nil, fmt.Errorf("failed to add liquidity: %w", err)
		}

		// Wrap in our custom position type
		customPos := NewCustomAMMPosition(poolPosition, s.pool)

		s.hasPosition = true

		// Calculate capital requirement
		ethPrice, _ := snapshot.Price("ETH/USD")
		capitalRequired := s.initialDepositA.MulPrice(ethPrice).Add(s.initialDepositB)

		return []strategy.Action{
			strategy.NewAddPositionAction(customPos),
			strategy.NewAdjustCashAction(
				capitalRequired.Decimal().Neg(),
				"capital for custom AMM position",
			),
		}, nil
	}

	// Subsequent calls: passive hold (could add rebalancing logic here)
	return nil, nil
}

// ====================================================================
// MAIN EXECUTION
// ====================================================================

func createSnapshots() []strategy.MarketSnapshot {
	snapshots := make([]strategy.MarketSnapshot, 0)
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for day := 0; day < 30; day++ {
		t := primitives.NewTime(startTime.Add(time.Duration(day) * 24 * time.Hour))

		// Simple price variation
		basePrice := 2000.0
		variation := 100.0 * (float64(day%10) / 10.0)
		if day%20 >= 10 {
			variation = -variation
		}
		ethPrice := basePrice + variation

		prices := map[string]primitives.Price{
			"ETH/USD": primitives.MustPrice(primitives.MustDecimalFromString(fmt.Sprintf("%.2f", ethPrice))),
		}

		snapshot := strategy.NewSimpleSnapshot(t, prices)
		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

func main() {
	fmt.Println("=== Custom Mechanism Example ===")
	fmt.Println("Demonstrating framework extensibility")
	fmt.Println()

	// 1. Create custom mechanism (constant-product AMM)
	fmt.Println("Step 1: Creating custom constant-product AMM...")
	pool := NewConstantProductPool(
		"custom-eth-usdc-pool",
		"ETH",
		"USDC",
		primitives.NewDecimalFromFloat(0.003), // 0.3% fee
	)
	fmt.Printf("  Pool ID: %s\n", pool.poolID)
	fmt.Printf("  Type: Constant Product (x * y = k)\n")
	fmt.Printf("  Fee: 0.3%%\n\n")

	// 2. Verify mechanism implements required interfaces
	fmt.Println("Step 2: Verifying interface implementation...")
	var _ mechanisms.MarketMechanism = pool
	var _ mechanisms.LiquidityPool = pool
	fmt.Println("  ✓ Implements MarketMechanism interface")
	fmt.Println("  ✓ Implements LiquidityPool interface")
	fmt.Println()

	// 3. Create strategy using custom mechanism
	fmt.Println("Step 3: Creating strategy with custom mechanism...")
	depositA := primitives.MustAmount(primitives.NewDecimal(5))     // 5 ETH
	depositB := primitives.MustAmount(primitives.NewDecimal(10000)) // 10000 USDC
	strat := NewCustomAMMStrategy(pool, depositA, depositB)
	fmt.Printf("  Initial deposit: %s ETH + %s USDC\n\n", depositA.String(), depositB.String())

	// 4. Run backtest using framework engine
	fmt.Println("Step 4: Running backtest with framework engine...")
	snapshots := createSnapshots()
	fmt.Printf("  Generated %d days of market data\n", len(snapshots))

	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(100000)),
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	// 5. Display results
	fmt.Println("\n" + result.Summary())

	// 6. Verify custom position integration
	fmt.Println("\n=== Custom Mechanism Validation ===")
	positions := result.Portfolio.Positions()
	fmt.Printf("Positions tracked by framework: %d\n", len(positions))

	for _, pos := range positions {
		fmt.Printf("\nPosition Details:\n")
		fmt.Printf("  ID: %s\n", pos.ID())
		fmt.Printf("  Type: %s\n", pos.Type())

		if len(snapshots) > 0 {
			lastSnapshot := snapshots[len(snapshots)-1]
			value, err := pos.Value(lastSnapshot)
			if err != nil {
				fmt.Printf("  Value: Error - %v\n", err)
			} else {
				fmt.Printf("  Final Value: %s\n", value.String())
			}
		}

		// Verify it's our custom position
		if customPos, ok := pos.(*CustomAMMPosition); ok {
			fmt.Printf("  ✓ Successfully using CustomAMMPosition\n")
			fmt.Printf("  ✓ Pool Type: %v\n", customPos.poolPosition.Metadata["pool_type"])
		}
	}

	// 7. Extensibility summary
	fmt.Println("\n=== Extensibility Demonstration Summary ===")
	fmt.Println("✓ Created custom ConstantProductPool mechanism")
	fmt.Println("✓ Implemented mechanisms.LiquidityPool interface")
	fmt.Println("✓ Created CustomAMMPosition wrapper")
	fmt.Println("✓ Implemented strategy.Position interface")
	fmt.Println("✓ Used in strategy without framework modifications")
	fmt.Println("✓ Backtest engine processed custom mechanism seamlessly")
	fmt.Println("✓ Portfolio manager tracked custom position type")
	fmt.Println("\n✅ Zero framework modifications required!")
	fmt.Println("\nThis proves the framework is truly extensible:")
	fmt.Println("• Add new mechanism types without touching framework code")
	fmt.Println("• Compose custom mechanisms with existing ones")
	fmt.Println("• All framework features work with custom mechanisms")
	fmt.Println("• Type safety maintained throughout")
}
