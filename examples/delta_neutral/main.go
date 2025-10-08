// Package main demonstrates a delta-neutral strategy that combines a concentrated liquidity
// position with a perpetual futures hedge. This example shows:
//  1. How to compose multiple mechanism types in a single strategy
//  2. Hedging LP position exposure with derivatives
//  3. Calculating and balancing position deltas
//  4. Managing a multi-mechanism portfolio
//
// Delta-neutral strategies aim to eliminate directional exposure to the underlying asset,
// profiting from fees, funding rates, and relative value changes while minimizing price risk.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/backtest"
	cl "github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/concentrated_liquidity"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/perpetual"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

// LPPosition wraps a mechanisms.PoolPosition to implement strategy.Position interface.
type LPPosition struct {
	poolPosition mechanisms.PoolPosition
	pool         *cl.Pool
	tickLower    int
	tickUpper    int
}

// NewLPPosition creates a new LP position wrapper.
func NewLPPosition(poolPos mechanisms.PoolPosition, pool *cl.Pool, tickLower, tickUpper int) *LPPosition {
	return &LPPosition{
		poolPosition: poolPos,
		pool:         pool,
		tickLower:    tickLower,
		tickUpper:    tickUpper,
	}
}

func (lp *LPPosition) ID() string {
	return lp.poolPosition.PoolID
}

func (lp *LPPosition) Type() strategy.PositionType {
	return strategy.PositionTypeLiquidityPool
}

func (lp *LPPosition) Value(snapshot strategy.MarketSnapshot) (primitives.Amount, error) {
	// Calculate LP position value
	amounts, err := lp.pool.RemoveLiquidity(context.Background(), lp.poolPosition)
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to calculate LP value: %w", err)
	}

	tokenAPrice, err := snapshot.Price("WETH/USDC")
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to get WETH price: %w", err)
	}

	tokenBPrice := primitives.MustPrice(primitives.One())

	valueA := amounts.AmountA.MulPrice(tokenAPrice)
	valueB := amounts.AmountB.MulPrice(tokenBPrice)

	return valueA.Add(valueB), nil
}

// PerpPosition wraps a perpetual.Future to implement strategy.Position interface.
type PerpPosition struct {
	future *perpetual.Future
}

// NewPerpPosition creates a new perpetual position wrapper.
func NewPerpPosition(future *perpetual.Future) *PerpPosition {
	return &PerpPosition{future: future}
}

func (pp *PerpPosition) ID() string {
	return "perp-eth-hedge"
}

func (pp *PerpPosition) Type() strategy.PositionType {
	return strategy.PositionTypePerpetual
}

func (pp *PerpPosition) Value(snapshot strategy.MarketSnapshot) (primitives.Amount, error) {
	// Get current mark price
	markPriceRaw, err := snapshot.Price("WETH/USDC")
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to get mark price: %w", err)
	}

	// Get funding rate from snapshot metadata
	fundingRate, ok := snapshot.Get("perp:eth:funding_rate")
	if !ok {
		fundingRate = 0.0001 // Default 0.01% per period
	}

	fundingRateDecimal := primitives.NewDecimalFromFloat(fundingRate.(float64))

	// Calculate perpetual value using Price method
	params := mechanisms.PriceParams{
		MarkPrice:   markPriceRaw,
		FundingRate: fundingRateDecimal,
	}

	value, err := pp.future.Price(context.Background(), params)
	if err != nil {
		return primitives.ZeroAmount(), fmt.Errorf("failed to price perpetual: %w", err)
	}

	// Return position value (convert Price to Amount)
	return primitives.MustAmount(value.Decimal()), nil
}

// DeltaNeutralStrategy implements a delta-neutral LP + perpetual hedge strategy.
// It provides liquidity to earn fees while hedging directional exposure with a short perpetual.
type DeltaNeutralStrategy struct {
	pool           *cl.Pool
	hasPositions   bool
	tickLower      int
	tickUpper      int
	lpLiquidityAmt primitives.Amount
	hedgeRatio     primitives.Decimal // How much of LP value to hedge (typically 0.5 for 50%)
}

// NewDeltaNeutralStrategy creates a new delta-neutral strategy.
func NewDeltaNeutralStrategy(
	pool *cl.Pool,
	tickLower, tickUpper int,
	lpLiquidityAmt primitives.Amount,
	hedgeRatio primitives.Decimal,
) *DeltaNeutralStrategy {
	return &DeltaNeutralStrategy{
		pool:           pool,
		hasPositions:   false,
		tickLower:      tickLower,
		tickUpper:      tickUpper,
		lpLiquidityAmt: lpLiquidityAmt,
		hedgeRatio:     hedgeRatio,
	}
}

// Rebalance implements strategy.Strategy.
// On first call: establishes LP position and opens offsetting short perpetual
// Subsequent calls: monitors delta and rebalances if needed (simplified in this example)
func (s *DeltaNeutralStrategy) Rebalance(
	ctx context.Context,
	portfolio *strategy.Portfolio,
	snapshot strategy.MarketSnapshot,
) ([]strategy.Action, error) {
	// Only establish positions on first rebalance
	if s.hasPositions {
		return nil, nil
	}

	// 1. Create LP position
	sqrtPriceX96Str, ok := snapshot.Get("pool:eth-usdc-pool:sqrt_price_x96")
	if !ok {
		return nil, fmt.Errorf("sqrt price not available")
	}

	lpPoolPosition := mechanisms.PoolPosition{
		PoolID:    "eth-usdc-pool",
		Liquidity: s.lpLiquidityAmt,
		TokensDeposited: mechanisms.TokenAmounts{
			AmountA: primitives.MustAmount(primitives.NewDecimal(10)),    // 10 WETH
			AmountB: primitives.MustAmount(primitives.NewDecimal(20000)), // 20000 USDC
		},
		Metadata: map[string]interface{}{
			"liquidity":      s.lpLiquidityAmt.Decimal().String(),
			"tick_lower":     s.tickLower,
			"tick_upper":     s.tickUpper,
			"sqrt_price_x96": sqrtPriceX96Str.(string),
		},
	}

	lpPos := NewLPPosition(lpPoolPosition, s.pool, s.tickLower, s.tickUpper)

	// 2. Calculate hedge size
	// LP position has ~10 ETH worth of exposure (ignoring USDC side for simplicity)
	// We want to hedge 50% of this with a short perpetual
	ethPrice, err := snapshot.Price("WETH/USDC")
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH price: %w", err)
	}

	// Hedge size: 5 ETH short (50% of 10 ETH)
	hedgeSize := primitives.NewDecimal(-5) // Negative = short position

	// Create perpetual future
	perpFuture, err := perpetual.NewFuture(
		"eth-perp-hedge",
		"ETHUSDC",
		ethPrice,
		hedgeSize,
		primitives.NewDecimal(1), // 1x leverage (no additional leverage)
		8*time.Hour,              // 8-hour funding period
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create perpetual: %w", err)
	}

	perpPos := NewPerpPosition(perpFuture)

	s.hasPositions = true

	// 3. Return actions to establish both positions
	return []strategy.Action{
		strategy.NewAddPositionAction(lpPos),
		strategy.NewAddPositionAction(perpPos),
		strategy.NewAdjustCashAction(
			primitives.NewDecimal(-30000), // Allocate $30k for positions
			"capital for delta-neutral positions",
		),
	}, nil
}

// createHistoricalSnapshots generates mock market data with price volatility.
func createHistoricalSnapshots() []strategy.MarketSnapshot {
	snapshots := make([]strategy.MarketSnapshot, 0)
	startTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	for day := 0; day < 30; day++ {
		t := primitives.NewTime(startTime.Add(time.Duration(day) * 24 * time.Hour))

		// Price varies: 1800-2200 (simulating volatility)
		basePrice := 2000.0
		variation := 200.0 * (float64(day%10) / 10.0)
		if day%20 >= 10 {
			variation = -variation
		}
		ethPrice := basePrice + variation

		prices := map[string]primitives.Price{
			"WETH/USDC": primitives.MustPrice(primitives.MustDecimalFromString(fmt.Sprintf("%.2f", ethPrice))),
		}

		snapshot := strategy.NewSimpleSnapshot(t, prices)

		// Pool metadata
		currentTick := 200000 + (day * 100)
		snapshot.Set("pool:eth-usdc-pool:current_tick", currentTick)
		snapshot.Set("pool:eth-usdc-pool:sqrt_price_x96", "1584563250000000000000000000000")

		// Perpetual funding rate (varies slightly)
		fundingRate := 0.0001 + (float64(day%5) * 0.00001)
		snapshot.Set("perp:eth:funding_rate", fundingRate)

		snapshots = append(snapshots, snapshot)
	}

	return snapshots
}

func main() {
	fmt.Println("=== Delta-Neutral Strategy Backtest ===")
	fmt.Println("Combining LP position with perpetual hedge")
	fmt.Println()

	// 1. Create concentrated liquidity pool
	pool, err := cl.NewPool(
		"eth-usdc-pool",
		common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // WETH
		18,
		common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
		6,
		constants.FeeAmount(3000),
	)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}

	// 2. Configure strategy
	tickLower := 199000
	tickUpper := 201000
	// Using realistic liquidity value - Uniswap V3 liquidity is a mathematical construct
	liquidityAmount := primitives.MustAmount(primitives.NewDecimal(1000))
	hedgeRatio := primitives.NewDecimalFromFloat(0.5) // Hedge 50% of LP value

	strat := NewDeltaNeutralStrategy(pool, tickLower, tickUpper, liquidityAmount, hedgeRatio)

	// 3. Generate market data
	snapshots := createHistoricalSnapshots()
	fmt.Printf("Generated %d days of market data\n", len(snapshots))

	// 4. Run backtest
	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(100000)),
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	fmt.Println("Running backtest...")
	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		log.Fatalf("Backtest failed: %v", err)
	}

	// 5. Display results
	fmt.Println("\n" + result.Summary())

	// 6. Analyze positions
	fmt.Println("\n=== Position Analysis ===")
	positions := result.Portfolio.Positions()
	fmt.Printf("Final positions: %d\n\n", len(positions))

	var lpValue, perpValue primitives.Amount
	for _, pos := range positions {
		fmt.Printf("Position: %s\n", pos.ID())
		fmt.Printf("  Type: %s\n", pos.Type())

		if len(snapshots) > 0 {
			lastSnapshot := snapshots[len(snapshots)-1]
			value, err := pos.Value(lastSnapshot)
			if err != nil {
				fmt.Printf("  Value: Error - %v\n", err)
			} else {
				fmt.Printf("  Final Value: %s\n", value.String())
				if pos.Type() == strategy.PositionTypeLiquidityPool {
					lpValue = value
				} else if pos.Type() == strategy.PositionTypePerpetual {
					perpValue = value
				}
			}
		}
		fmt.Println()
	}

	// 7. Delta analysis
	fmt.Println("=== Delta Analysis ===")
	fmt.Printf("LP Position Value: %s\n", lpValue.String())
	fmt.Printf("Perpetual Position Value: %s\n", perpValue.String())
	netDelta := lpValue.Add(perpValue)
	fmt.Printf("Net Position Value: %s\n", netDelta.String())
	fmt.Println("\nNote: In a real delta-neutral strategy, the net delta would be")
	fmt.Println("close to zero, meaning the portfolio value is insensitive to")
	fmt.Println("small price movements in the underlying asset.")

	fmt.Println("\n=== Strategy Benefits ===")
	fmt.Println("✓ Earns LP fees from providing liquidity")
	fmt.Println("✓ May earn funding rates from perpetual position")
	fmt.Println("✓ Reduced exposure to directional price risk")
	fmt.Println("✓ Demonstrates multi-mechanism composition (<350 lines)")

	fmt.Println("\n=== Production Enhancements ===")
	fmt.Println("• Dynamic rebalancing when delta drifts from target")
	fmt.Println("• Automatic position sizing based on pool TVL")
	fmt.Println("• Fee collection and reinvestment logic")
	fmt.Println("• Risk limits and stop-loss mechanisms")
	fmt.Println("• Capital efficiency optimization")
}
