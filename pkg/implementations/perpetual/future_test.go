package perpetual_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/perpetual"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// Test constants
const (
	tolerance = 0.0001 // 0.01% tolerance
)

// TestNewFuture tests the creation of new perpetual futures.
func TestNewFuture(t *testing.T) {
	tests := []struct {
		name          string
		futureID      string
		symbol        string
		entryPrice    primitives.Price
		positionSize  primitives.Decimal
		leverage      primitives.Decimal
		fundingPeriod time.Duration
		expectError   bool
		expectedError string
	}{
		{
			name:          "Valid long position",
			futureID:      "BTC-PERP-1",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 8 * time.Hour,
			expectError:   false,
		},
		{
			name:          "Valid short position",
			futureID:      "ETH-PERP-1",
			symbol:        "ETHUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(3000)),
			positionSize:  primitives.NewDecimalFromFloat(-2.0),
			leverage:      primitives.NewDecimal(5),
			fundingPeriod: 8 * time.Hour,
			expectError:   false,
		},
		{
			name:          "Empty futureID",
			futureID:      "",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 8 * time.Hour,
			expectError:   true,
			expectedError: "futureID cannot be empty",
		},
		{
			name:          "Empty symbol",
			futureID:      "BTC-PERP-1",
			symbol:        "",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 8 * time.Hour,
			expectError:   true,
			expectedError: "symbol cannot be empty",
		},
		{
			name:          "Zero entry price",
			futureID:      "BTC-PERP-1",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.ZeroPrice(),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 8 * time.Hour,
			expectError:   true,
			expectedError: "entry price must be positive",
		},
		{
			name:          "Zero position size",
			futureID:      "BTC-PERP-1",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.Zero(),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 8 * time.Hour,
			expectError:   true,
			expectedError: perpetual.ErrInvalidPositionSize.Error(),
		},
		{
			name:          "Invalid leverage (<1)",
			futureID:      "BTC-PERP-1",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimalFromFloat(0.5),
			fundingPeriod: 8 * time.Hour,
			expectError:   true,
			expectedError: perpetual.ErrInvalidLeverage.Error(),
		},
		{
			name:          "Zero funding period",
			futureID:      "BTC-PERP-1",
			symbol:        "BTCUSDT",
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(50000)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			leverage:      primitives.NewDecimal(10),
			fundingPeriod: 0,
			expectError:   true,
			expectedError: "funding period must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future, err := perpetual.NewFuture(
				tt.futureID,
				tt.symbol,
				tt.entryPrice,
				tt.positionSize,
				tt.leverage,
				tt.fundingPeriod,
			)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
					return
				}
				if tt.expectedError != "" && err.Error() != tt.expectedError {
					t.Errorf("Expected error %q but got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if future == nil {
				t.Error("Expected future but got nil")
				return
			}

			// Verify future properties
			if future.FutureID() != tt.futureID {
				t.Errorf("Expected futureID %q but got %q", tt.futureID, future.FutureID())
			}
			if future.Symbol() != tt.symbol {
				t.Errorf("Expected symbol %q but got %q", tt.symbol, future.Symbol())
			}

			// Verify direction
			expectedDirection := mechanisms.PositionDirectionLong
			if tt.positionSize.IsNegative() {
				expectedDirection = mechanisms.PositionDirectionShort
			}
			if future.Direction() != expectedDirection {
				t.Errorf("Expected direction %v but got %v", expectedDirection, future.Direction())
			}
		})
	}
}

// TestFuturePricing tests the pricing (mark price) of perpetual futures.
func TestFuturePricing(t *testing.T) {
	future, err := perpetual.NewFuture(
		"BTC-PERP",
		"BTCUSDT",
		primitives.MustPrice(primitives.NewDecimal(50000)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.NewDecimal(10),
		8*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create future: %v", err)
	}

	tests := []struct {
		name          string
		markPrice     primitives.Price
		expectedPrice primitives.Price
		expectError   bool
	}{
		{
			name:          "Valid mark price",
			markPrice:     primitives.MustPrice(primitives.NewDecimal(51000)),
			expectedPrice: primitives.MustPrice(primitives.NewDecimal(51000)),
			expectError:   false,
		},
		{
			name:        "Zero mark price",
			markPrice:   primitives.ZeroPrice(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := mechanisms.PriceParams{
				MarkPrice: tt.markPrice,
			}

			price, err := future.Price(context.Background(), params)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !price.Equal(tt.expectedPrice) {
				t.Errorf("Expected price %v but got %v", tt.expectedPrice, price)
			}
		})
	}
}

// TestFutureGreeks tests the Greeks calculation for perpetuals.
func TestFutureGreeks(t *testing.T) {
	tests := []struct {
		name          string
		positionSize  primitives.Decimal
		markPrice     float64
		fundingRate   float64
		expectedDelta float64
	}{
		{
			name:          "Long position",
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			markPrice:     50000.0,
			fundingRate:   0.0001, // 0.01%
			expectedDelta: 1.0,
		},
		{
			name:          "Short position",
			positionSize:  primitives.NewDecimalFromFloat(-1.0),
			markPrice:     50000.0,
			fundingRate:   0.0001,
			expectedDelta: -1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future, err := perpetual.NewFuture(
				"TEST",
				"BTCUSDT",
				primitives.MustPrice(primitives.NewDecimal(50000)),
				tt.positionSize,
				primitives.NewDecimal(10),
				8*time.Hour,
			)
			if err != nil {
				t.Fatalf("Failed to create future: %v", err)
			}

			params := mechanisms.PriceParams{
				MarkPrice:   primitives.MustPrice(primitives.NewDecimalFromFloat(tt.markPrice)),
				FundingRate: primitives.NewDecimalFromFloat(tt.fundingRate),
			}

			greeks, err := future.Greeks(context.Background(), params)
			if err != nil {
				t.Fatalf("Failed to calculate Greeks: %v", err)
			}

			// Check Delta
			delta := greeks.Delta.Float64()
			if math.Abs(delta-tt.expectedDelta) > tolerance {
				t.Errorf("Delta mismatch: expected %.2f, got %.2f", tt.expectedDelta, delta)
			}

			// Check that other Greeks are zero (perpetuals don't have gamma, vega, rho)
			if !greeks.Gamma.IsZero() {
				t.Error("Gamma should be zero for perpetuals")
			}
			if !greeks.Vega.IsZero() {
				t.Error("Vega should be zero for perpetuals")
			}
			if !greeks.Rho.IsZero() {
				t.Error("Rho should be zero for perpetuals")
			}
		})
	}
}

// TestUnrealizedPnL tests unrealized P&L calculations.
func TestUnrealizedPnL(t *testing.T) {
	tests := []struct {
		name         string
		entryPrice   float64
		positionSize float64
		currentPrice float64
		expectedPnL  float64
	}{
		{
			name:         "Profitable long",
			entryPrice:   50000.0,
			positionSize: 1.0,
			currentPrice: 51000.0,
			expectedPnL:  1000.0,
		},
		{
			name:         "Losing long",
			entryPrice:   50000.0,
			positionSize: 1.0,
			currentPrice: 49000.0,
			expectedPnL:  -1000.0,
		},
		{
			name:         "Profitable short",
			entryPrice:   50000.0,
			positionSize: -1.0,
			currentPrice: 49000.0,
			expectedPnL:  1000.0,
		},
		{
			name:         "Losing short",
			entryPrice:   50000.0,
			positionSize: -1.0,
			currentPrice: 51000.0,
			expectedPnL:  -1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future, err := perpetual.NewFuture(
				"TEST",
				"BTCUSDT",
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.entryPrice)),
				primitives.NewDecimalFromFloat(tt.positionSize),
				primitives.NewDecimal(10),
				8*time.Hour,
			)
			if err != nil {
				t.Fatalf("Failed to create future: %v", err)
			}

			currentPrice := primitives.MustPrice(primitives.NewDecimalFromFloat(tt.currentPrice))
			pnl, err := future.UnrealizedPnL(currentPrice)
			if err != nil {
				t.Fatalf("Failed to calculate unrealized P&L: %v", err)
			}

			actualPnL := pnl.Float64()
			if math.Abs(actualPnL-tt.expectedPnL) > tolerance {
				t.Errorf("P&L mismatch: expected %.2f, got %.2f", tt.expectedPnL, actualPnL)
			}
		})
	}
}

// TestApplyFunding tests funding rate application.
func TestApplyFunding(t *testing.T) {
	tests := []struct {
		name                string
		positionSize        float64
		markPrice           float64
		fundingRate         float64
		expectedPayment     float64
		expectedAccumulated float64
	}{
		{
			name:                "Long pays positive funding",
			positionSize:        1.0,
			markPrice:           50000.0,
			fundingRate:         0.0001, // 0.01%
			expectedPayment:     5.0,    // 1 * 50000 * 0.0001
			expectedAccumulated: 5.0,
		},
		{
			name:                "Short receives positive funding",
			positionSize:        -1.0,
			markPrice:           50000.0,
			fundingRate:         0.0001,
			expectedPayment:     -5.0, // Negative payment = receipt
			expectedAccumulated: -5.0,
		},
		{
			name:                "Long receives negative funding",
			positionSize:        1.0,
			markPrice:           50000.0,
			fundingRate:         -0.0001,
			expectedPayment:     -5.0, // Negative payment = receipt
			expectedAccumulated: -5.0,
		},
		{
			name:                "Short pays negative funding",
			positionSize:        -1.0,
			markPrice:           50000.0,
			fundingRate:         -0.0001,
			expectedPayment:     5.0,
			expectedAccumulated: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future, err := perpetual.NewFuture(
				"TEST",
				"BTCUSDT",
				primitives.MustPrice(primitives.NewDecimal(50000)),
				primitives.NewDecimalFromFloat(tt.positionSize),
				primitives.NewDecimal(10),
				8*time.Hour,
			)
			if err != nil {
				t.Fatalf("Failed to create future: %v", err)
			}

			payment, err := future.ApplyFunding(
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.markPrice)),
				primitives.NewDecimalFromFloat(tt.fundingRate),
			)
			if err != nil {
				t.Fatalf("Failed to apply funding: %v", err)
			}

			actualPayment := payment.Float64()
			if math.Abs(actualPayment-tt.expectedPayment) > tolerance {
				t.Errorf("Payment mismatch: expected %.2f, got %.2f",
					tt.expectedPayment, actualPayment)
			}

			actualAccumulated := future.AccumulatedFunding().Float64()
			if math.Abs(actualAccumulated-tt.expectedAccumulated) > tolerance {
				t.Errorf("Accumulated funding mismatch: expected %.2f, got %.2f",
					tt.expectedAccumulated, actualAccumulated)
			}
		})
	}
}

// TestCalculateFundingRate tests the funding rate calculation utility.
func TestCalculateFundingRate(t *testing.T) {
	tests := []struct {
		name                string
		markPrice           float64
		indexPrice          float64
		multiplier          float64
		expectedFundingRate float64
	}{
		{
			name:                "Perpetual trading at premium",
			markPrice:           50500.0,
			indexPrice:          50000.0,
			multiplier:          0.333,
			expectedFundingRate: 0.00333, // (50500-50000)/50000 * 0.333
		},
		{
			name:                "Perpetual trading at discount",
			markPrice:           49500.0,
			indexPrice:          50000.0,
			multiplier:          0.333,
			expectedFundingRate: -0.00333, // (49500-50000)/50000 * 0.333
		},
		{
			name:                "Perpetual at parity",
			markPrice:           50000.0,
			indexPrice:          50000.0,
			multiplier:          0.333,
			expectedFundingRate: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fundingRate, err := perpetual.CalculateFundingRate(
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.markPrice)),
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.indexPrice)),
				primitives.NewDecimalFromFloat(tt.multiplier),
			)
			if err != nil {
				t.Fatalf("Failed to calculate funding rate: %v", err)
			}

			actualRate := fundingRate.Float64()
			percentDiff := math.Abs(actualRate-tt.expectedFundingRate) / math.Abs(tt.expectedFundingRate+0.0001)
			if percentDiff > tolerance {
				t.Errorf("Funding rate mismatch: expected %.6f, got %.6f (%.2f%% diff)",
					tt.expectedFundingRate, actualRate, percentDiff*100)
			}
		})
	}
}

// TestLiquidationPrice tests liquidation price calculations.
func TestLiquidationPrice(t *testing.T) {
	tests := []struct {
		name                string
		entryPrice          float64
		leverage            float64
		positionSize        float64
		expectedLiquidation float64
	}{
		{
			name:                "10x long",
			entryPrice:          50000.0,
			leverage:            10.0,
			positionSize:        1.0,
			expectedLiquidation: 45000.0, // 50000 * (1 - 1/10)
		},
		{
			name:                "10x short",
			entryPrice:          50000.0,
			leverage:            10.0,
			positionSize:        -1.0,
			expectedLiquidation: 55000.0, // 50000 * (1 + 1/10)
		},
		{
			name:                "5x long",
			entryPrice:          50000.0,
			leverage:            5.0,
			positionSize:        1.0,
			expectedLiquidation: 40000.0, // 50000 * (1 - 1/5)
		},
		{
			name:                "20x long",
			entryPrice:          50000.0,
			leverage:            20.0,
			positionSize:        1.0,
			expectedLiquidation: 47500.0, // 50000 * (1 - 1/20)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			future, err := perpetual.NewFuture(
				"TEST",
				"BTCUSDT",
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.entryPrice)),
				primitives.NewDecimalFromFloat(tt.positionSize),
				primitives.NewDecimalFromFloat(tt.leverage),
				8*time.Hour,
			)
			if err != nil {
				t.Fatalf("Failed to create future: %v", err)
			}

			liquidationPrice, err := future.LiquidationPrice()
			if err != nil {
				t.Fatalf("Failed to calculate liquidation price: %v", err)
			}

			actualPrice := liquidationPrice.Decimal().Float64()
			if math.Abs(actualPrice-tt.expectedLiquidation) > tolerance {
				t.Errorf("Liquidation price mismatch: expected %.2f, got %.2f",
					tt.expectedLiquidation, actualPrice)
			}
		})
	}
}

// TestSettlement tests position settlement.
func TestSettlement(t *testing.T) {
	future, err := perpetual.NewFuture(
		"TEST",
		"BTCUSDT",
		primitives.MustPrice(primitives.NewDecimal(50000)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.NewDecimal(10),
		8*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create future: %v", err)
	}

	// Apply some funding before settlement
	_, err = future.ApplyFunding(
		primitives.MustPrice(primitives.NewDecimal(50000)),
		primitives.NewDecimalFromFloat(0.0001),
	)
	if err != nil {
		t.Fatalf("Failed to apply funding: %v", err)
	}

	// Settle at a profit
	finalPrice := primitives.MustPrice(primitives.NewDecimal(51000))
	pnl, err := future.SettleWithPrice(finalPrice)
	if err != nil {
		t.Fatalf("Failed to settle: %v", err)
	}

	// Expected: 1000 (price gain) - 5 (funding) = 995
	expectedPnL := 995.0
	actualPnL := pnl.Decimal().Float64()

	if math.Abs(actualPnL-expectedPnL) > tolerance {
		t.Errorf("Settlement P&L mismatch: expected %.2f, got %.2f",
			expectedPnL, actualPnL)
	}

	// Verify settled status
	if !future.IsSettled() {
		t.Error("Future should be marked as settled")
	}

	// Try to settle again (should error)
	_, err = future.SettleWithPrice(finalPrice)
	if err == nil {
		t.Error("Expected error when settling already settled position")
	}
}

// TestMechanismInterface tests that Future implements the Derivative interface.
func TestMechanismInterface(t *testing.T) {
	future, err := perpetual.NewFuture(
		"TEST",
		"BTCUSDT",
		primitives.MustPrice(primitives.NewDecimal(50000)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.NewDecimal(10),
		8*time.Hour,
	)
	if err != nil {
		t.Fatalf("Failed to create future: %v", err)
	}

	// Test Mechanism() returns correct type
	if future.Mechanism() != mechanisms.MechanismTypeDerivative {
		t.Errorf("Expected mechanism type %v but got %v",
			mechanisms.MechanismTypeDerivative, future.Mechanism())
	}

	// Test Venue() returns expected value
	if future.Venue() != "perpetual" {
		t.Errorf("Expected venue 'perpetual' but got %v", future.Venue())
	}
}
