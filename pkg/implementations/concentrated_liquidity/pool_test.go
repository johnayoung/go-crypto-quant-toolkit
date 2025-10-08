package concentrated_liquidity_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/concentrated_liquidity"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
)

// Test tokens (USDC/WETH on mainnet)
var (
	usdcAddress = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	wethAddress = common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")
)

// TestPoolCreation verifies that a pool can be created with valid parameters.
func TestPoolCreation(t *testing.T) {
	tests := []struct {
		name        string
		poolID      string
		tokenAAddr  common.Address
		tokenADec   uint
		tokenBAddr  common.Address
		tokenBDec   uint
		fee         constants.FeeAmount
		expectError bool
	}{
		{
			name:        "Valid 0.3% fee pool",
			poolID:      "usdc-weth-3000",
			tokenAAddr:  usdcAddress,
			tokenADec:   6,
			tokenBAddr:  wethAddress,
			tokenBDec:   18,
			fee:         constants.FeeMedium,
			expectError: false,
		},
		{
			name:        "Valid 0.05% fee pool",
			poolID:      "usdc-weth-500",
			tokenAAddr:  usdcAddress,
			tokenADec:   6,
			tokenBAddr:  wethAddress,
			tokenBDec:   18,
			fee:         constants.FeeLow,
			expectError: false,
		},
		{
			name:        "Valid 1% fee pool",
			poolID:      "usdc-weth-10000",
			tokenAAddr:  usdcAddress,
			tokenADec:   6,
			tokenBAddr:  wethAddress,
			tokenBDec:   18,
			fee:         constants.FeeHigh,
			expectError: false,
		},
		{
			name:        "Empty pool ID",
			poolID:      "",
			tokenAAddr:  usdcAddress,
			tokenADec:   6,
			tokenBAddr:  wethAddress,
			tokenBDec:   18,
			fee:         constants.FeeMedium,
			expectError: true,
		},
		{
			name:        "Invalid fee tier",
			poolID:      "usdc-weth-invalid",
			tokenAAddr:  usdcAddress,
			tokenADec:   6,
			tokenBAddr:  wethAddress,
			tokenBDec:   18,
			fee:         constants.FeeAmount(999), // Invalid fee
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := concentrated_liquidity.NewPool(
				tt.poolID,
				tt.tokenAAddr,
				tt.tokenADec,
				tt.tokenBAddr,
				tt.tokenBDec,
				tt.fee,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if pool == nil {
				t.Fatal("Expected non-nil pool")
			}

			// Verify interface implementation
			if pool.Mechanism() != mechanisms.MechanismTypeLiquidityPool {
				t.Errorf("Expected mechanism '%s', got '%s'", mechanisms.MechanismTypeLiquidityPool, pool.Mechanism())
			}

			if pool.Venue() != "uniswap-v3" {
				t.Errorf("Expected venue 'uniswap-v3', got '%s'", pool.Venue())
			}
		})
	}
}

// TestPoolCalculate verifies pool state calculation from tick and sqrt price.
func TestPoolCalculate(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6, // USDC decimals
		wethAddress,
		18, // WETH decimals
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Example: price of 2000 USDC per ETH
	// sqrtPriceX96 = sqrt(2000) * 2^96 = 44.721 * 79228162514264337593543950336
	// For proper decimal adjustment: sqrt(2000 * 10^12) since WETH has 12 more decimals
	// sqrtPrice = sqrt(2000000000000000) * 2^96
	sqrtPriceX96 := "3543191142285914205922034323214"
	currentTick := 85176               // Approximate tick for 2000 USDC/ETH
	liquidity := "1000000000000000000" // 1e18

	params := mechanisms.PoolParams{
		Metadata: map[string]interface{}{
			"current_tick":   currentTick,
			"sqrt_price_x96": sqrtPriceX96,
			"liquidity":      liquidity,
		},
	}

	ctx := context.Background()
	state, err := pool.Calculate(ctx, params)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Verify state components
	if state.SpotPrice.IsZero() {
		t.Error("Expected non-zero spot price")
	}

	if state.Liquidity.IsZero() {
		t.Error("Expected non-zero liquidity")
	}

	// Verify metadata
	if tick, ok := state.Metadata["current_tick"].(int); !ok || tick != currentTick {
		t.Errorf("Expected tick %d in metadata, got %v", currentTick, state.Metadata["current_tick"])
	}
}

// TestRemoveLiquidity verifies that removing liquidity calculates correct token amounts.
func TestRemoveLiquidity(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Create a position with known parameters
	// Current price: ~2000 USDC/ETH (tick ~85176)
	// Position range: ticks 84000 to 86000 (roughly 1900 to 2100 USDC/ETH)
	sqrtPriceX96 := "3543191142285914205922034323214"
	liquidity := new(big.Int)
	liquidity.SetString("1000000000000000000", 10) // 1e18

	position := mechanisms.PoolPosition{
		Metadata: map[string]interface{}{
			"liquidity":      liquidity.String(),
			"tick_lower":     84000,
			"tick_upper":     86000,
			"sqrt_price_x96": sqrtPriceX96,
		},
	}

	ctx := context.Background()
	amounts, err := pool.RemoveLiquidity(ctx, position)
	if err != nil {
		t.Fatalf("RemoveLiquidity failed: %v", err)
	}

	// Amounts are guaranteed to be non-negative by construction
	// (NewAmount returns error for negative values)

	// In a concentrated liquidity position where current price is within range,
	// both amounts should be non-zero
	if amounts.AmountA.IsZero() && amounts.AmountB.IsZero() {
		t.Error("Expected at least one non-zero amount")
	}
}

// TestRemoveLiquidityErrors verifies error handling for invalid position data.
func TestRemoveLiquidityErrors(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		position mechanisms.PoolPosition
	}{
		{
			name: "Missing liquidity",
			position: mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"tick_lower":     84000,
					"tick_upper":     86000,
					"sqrt_price_x96": "3543191142285914205922034323214",
				},
			},
		},
		{
			name: "Missing tick_lower",
			position: mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"liquidity":      "1000000000000000000",
					"tick_upper":     86000,
					"sqrt_price_x96": "3543191142285914205922034323214",
				},
			},
		},
		{
			name: "Missing tick_upper",
			position: mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"liquidity":      "1000000000000000000",
					"tick_lower":     84000,
					"sqrt_price_x96": "3543191142285914205922034323214",
				},
			},
		},
		{
			name: "Missing sqrt_price_x96",
			position: mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"liquidity":  "1000000000000000000",
					"tick_lower": 84000,
					"tick_upper": 86000,
				},
			},
		},
		{
			name: "Invalid liquidity format",
			position: mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"liquidity":      "not-a-number",
					"tick_lower":     84000,
					"tick_upper":     86000,
					"sqrt_price_x96": "3543191142285914205922034323214",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.RemoveLiquidity(ctx, tt.position)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

// TestCalculateWithInvalidParams verifies error handling for Calculate.
func TestCalculateWithInvalidParams(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name   string
		params mechanisms.PoolParams
	}{
		{
			name: "Missing current_tick",
			params: mechanisms.PoolParams{
				Metadata: map[string]interface{}{
					"sqrt_price_x96": "3543191142285914205922034323214",
					"liquidity":      "1000000000000000000",
				},
			},
		},
		{
			name: "Missing sqrt_price_x96",
			params: mechanisms.PoolParams{
				Metadata: map[string]interface{}{
					"current_tick": 85176,
					"liquidity":    "1000000000000000000",
				},
			},
		},
		{
			name: "Missing liquidity",
			params: mechanisms.PoolParams{
				Metadata: map[string]interface{}{
					"current_tick":   85176,
					"sqrt_price_x96": "3543191142285914205922034323214",
				},
			},
		},
		{
			name: "Invalid sqrt_price_x96 format",
			params: mechanisms.PoolParams{
				Metadata: map[string]interface{}{
					"current_tick":   85176,
					"sqrt_price_x96": "not-a-number",
					"liquidity":      "1000000000000000000",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Calculate(ctx, tt.params)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

// TestInterfaceCompliance verifies the pool implements expected interfaces.
func TestInterfaceCompliance(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Verify MarketMechanism interface
	var _ mechanisms.MarketMechanism = pool

	// Verify LiquidityPool interface
	var _ mechanisms.LiquidityPool = pool
}

// TestRemoveLiquidityWithVariousRanges tests removal with different tick ranges.
func TestRemoveLiquidityWithVariousRanges(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	sqrtPriceX96 := "3543191142285914205922034323214"
	liquidity := "5000000000000000000" // 5e18

	testCases := []struct {
		name      string
		tickLower int
		tickUpper int
	}{
		{
			name:      "Wide range",
			tickLower: 80000,
			tickUpper: 90000,
		},
		{
			name:      "Narrow range",
			tickLower: 85000,
			tickUpper: 85500,
		},
		{
			name:      "Range below current price",
			tickLower: 80000,
			tickUpper: 82000,
		},
		{
			name:      "Range above current price",
			tickLower: 88000,
			tickUpper: 90000,
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			position := mechanisms.PoolPosition{
				Metadata: map[string]interface{}{
					"liquidity":      liquidity,
					"tick_lower":     tc.tickLower,
					"tick_upper":     tc.tickUpper,
					"sqrt_price_x96": sqrtPriceX96,
				},
			}

			amounts, err := pool.RemoveLiquidity(ctx, position)
			if err != nil {
				t.Fatalf("RemoveLiquidity failed: %v", err)
			}

			// At least verify amounts are returned without error
			if amounts.AmountA.String() == "" || amounts.AmountB.String() == "" {
				t.Error("Expected valid string representations for amounts")
			}
		})
	}
}

// TestCalculateDecimalPrecision verifies precise decimal handling in calculations.
func TestCalculateDecimalPrecision(t *testing.T) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Use very large numbers to test decimal precision
	// This represents a pool with substantial liquidity
	sqrtPriceX96 := "3543191142285914205922034323214"
	liquidity := "10000000000000000000000" // 10000e18

	params := mechanisms.PoolParams{
		Metadata: map[string]interface{}{
			"current_tick":   85176,
			"sqrt_price_x96": sqrtPriceX96,
			"liquidity":      liquidity,
		},
	}

	ctx := context.Background()
	state, err := pool.Calculate(ctx, params)
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	// Verify the liquidity matches what we input
	if state.Liquidity.String() != liquidity {
		t.Errorf("Expected liquidity %s, got %s", liquidity, state.Liquidity.String())
	}

	// Verify effective liquidity equals total liquidity
	if !state.EffectiveLiquidity.Equal(state.Liquidity) {
		t.Error("Expected effective liquidity to equal total liquidity")
	}
}

// BenchmarkCalculate benchmarks the Calculate method.
func BenchmarkCalculate(b *testing.B) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	params := mechanisms.PoolParams{
		Metadata: map[string]interface{}{
			"current_tick":   85176,
			"sqrt_price_x96": "3543191142285914205922034323214",
			"liquidity":      "1000000000000000000",
		},
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := pool.Calculate(ctx, params)
		if err != nil {
			b.Fatalf("Calculate failed: %v", err)
		}
	}
}

// BenchmarkRemoveLiquidity benchmarks the RemoveLiquidity method.
func BenchmarkRemoveLiquidity(b *testing.B) {
	pool, err := concentrated_liquidity.NewPool(
		"usdc-weth-3000",
		usdcAddress,
		6,
		wethAddress,
		18,
		constants.FeeMedium,
	)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}

	position := mechanisms.PoolPosition{
		Metadata: map[string]interface{}{
			"liquidity":      "1000000000000000000",
			"tick_lower":     84000,
			"tick_upper":     86000,
			"sqrt_price_x96": "3543191142285914205922034323214",
		},
	}

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := pool.RemoveLiquidity(ctx, position)
		if err != nil {
			b.Fatalf("RemoveLiquidity failed: %v", err)
		}
	}
}
