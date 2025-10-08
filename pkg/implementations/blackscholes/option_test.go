package blackscholes_test

import (
	"context"
	"math"
	"testing"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/blackscholes"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// Test constants for option pricing
const (
	// Tolerance for floating point comparisons (0.2% to account for rounding)
	priceTolerance = 0.002
	// Tolerance for Greeks (0.5%)
	greeksTolerance = 0.005
)

// TestNewOption tests the creation of new options.
func TestNewOption(t *testing.T) {
	tests := []struct {
		name          string
		optionID      string
		optionType    mechanisms.OptionType
		strikePrice   primitives.Price
		timeToExpiry  primitives.Decimal
		entryPrice    primitives.Price
		positionSize  primitives.Decimal
		expectError   bool
		expectedError string
	}{
		{
			name:         "Valid call option",
			optionID:     "CALL-100",
			optionType:   mechanisms.OptionTypeCall,
			strikePrice:  primitives.MustPrice(primitives.NewDecimal(100)),
			timeToExpiry: primitives.NewDecimalFromFloat(1.0),
			entryPrice:   primitives.MustPrice(primitives.NewDecimal(10)),
			positionSize: primitives.NewDecimalFromFloat(1.0),
			expectError:  false,
		},
		{
			name:         "Valid put option",
			optionID:     "PUT-100",
			optionType:   mechanisms.OptionTypePut,
			strikePrice:  primitives.MustPrice(primitives.NewDecimal(100)),
			timeToExpiry: primitives.NewDecimalFromFloat(0.5),
			entryPrice:   primitives.MustPrice(primitives.NewDecimal(8)),
			positionSize: primitives.NewDecimalFromFloat(-1.0),
			expectError:  false,
		},
		{
			name:          "Empty optionID",
			optionID:      "",
			optionType:    mechanisms.OptionTypeCall,
			strikePrice:   primitives.MustPrice(primitives.NewDecimal(100)),
			timeToExpiry:  primitives.NewDecimalFromFloat(1.0),
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(10)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			expectError:   true,
			expectedError: "optionID cannot be empty",
		},
		{
			name:          "Zero strike price",
			optionID:      "CALL-0",
			optionType:    mechanisms.OptionTypeCall,
			strikePrice:   primitives.ZeroPrice(),
			timeToExpiry:  primitives.NewDecimalFromFloat(1.0),
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(10)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			expectError:   true,
			expectedError: blackscholes.ErrInvalidStrike.Error(),
		},
		{
			name:          "Negative time to expiry",
			optionID:      "CALL-100",
			optionType:    mechanisms.OptionTypeCall,
			strikePrice:   primitives.MustPrice(primitives.NewDecimal(100)),
			timeToExpiry:  primitives.NewDecimalFromFloat(-1.0),
			entryPrice:    primitives.MustPrice(primitives.NewDecimal(10)),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			expectError:   true,
			expectedError: blackscholes.ErrInvalidTimeToExpiry.Error(),
		},
		{
			name:          "Zero entry price",
			optionID:      "CALL-100",
			optionType:    mechanisms.OptionTypeCall,
			strikePrice:   primitives.MustPrice(primitives.NewDecimal(100)),
			timeToExpiry:  primitives.NewDecimalFromFloat(1.0),
			entryPrice:    primitives.ZeroPrice(),
			positionSize:  primitives.NewDecimalFromFloat(1.0),
			expectError:   true,
			expectedError: "entry price must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := blackscholes.NewOption(
				tt.optionID,
				tt.optionType,
				tt.strikePrice,
				tt.timeToExpiry,
				tt.entryPrice,
				tt.positionSize,
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

			if option == nil {
				t.Error("Expected option but got nil")
				return
			}

			// Verify option properties
			if option.OptionID() != tt.optionID {
				t.Errorf("Expected optionID %q but got %q", tt.optionID, option.OptionID())
			}
			if option.OptionType() != tt.optionType {
				t.Errorf("Expected optionType %v but got %v", tt.optionType, option.OptionType())
			}
		})
	}
}

// TestCallOptionPricing tests call option pricing against known values.
// Reference values calculated using standard Black-Scholes formula.
func TestCallOptionPricing(t *testing.T) {
	tests := []struct {
		name            string
		underlyingPrice float64
		strikePrice     float64
		timeToExpiry    float64
		volatility      float64
		riskFreeRate    float64
		expectedPrice   float64
	}{
		{
			name:            "ATM call",
			underlyingPrice: 100.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,  // 20%
			riskFreeRate:    0.05,  // 5%
			expectedPrice:   10.45, // Approximate Black-Scholes value
		},
		{
			name:            "ITM call",
			underlyingPrice: 110.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,
			riskFreeRate:    0.05,
			expectedPrice:   17.66, // Corrected Black-Scholes value
		},
		{
			name:            "OTM call",
			underlyingPrice: 90.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,
			riskFreeRate:    0.05,
			expectedPrice:   5.09, // Corrected Black-Scholes value
		},
		{
			name:            "Short-dated call",
			underlyingPrice: 100.0,
			strikePrice:     100.0,
			timeToExpiry:    0.25, // 3 months
			volatility:      0.30,
			riskFreeRate:    0.05,
			expectedPrice:   6.58, // Corrected Black-Scholes value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := blackscholes.NewOption(
				"TEST",
				mechanisms.OptionTypeCall,
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.strikePrice)),
				primitives.NewDecimalFromFloat(tt.timeToExpiry),
				primitives.MustPrice(primitives.NewDecimalFromFloat(1.0)),
				primitives.NewDecimalFromFloat(1.0),
			)
			if err != nil {
				t.Fatalf("Failed to create option: %v", err)
			}

			params := mechanisms.PriceParams{
				UnderlyingPrice: primitives.MustPrice(primitives.NewDecimalFromFloat(tt.underlyingPrice)),
				Volatility:      primitives.NewDecimalFromFloat(tt.volatility),
				RiskFreeRate:    primitives.NewDecimalFromFloat(tt.riskFreeRate),
				TimeToExpiry:    primitives.NewDecimalFromFloat(tt.timeToExpiry),
			}

			price, err := option.Price(context.Background(), params)
			if err != nil {
				t.Fatalf("Failed to price option: %v", err)
			}

			actualPrice := price.Decimal().Float64()
			percentDiff := math.Abs(actualPrice-tt.expectedPrice) / tt.expectedPrice

			if percentDiff > priceTolerance {
				t.Errorf("Price mismatch: expected %.4f, got %.4f (%.2f%% difference)",
					tt.expectedPrice, actualPrice, percentDiff*100)
			}
		})
	}
}

// TestPutOptionPricing tests put option pricing against known values.
func TestPutOptionPricing(t *testing.T) {
	tests := []struct {
		name            string
		underlyingPrice float64
		strikePrice     float64
		timeToExpiry    float64
		volatility      float64
		riskFreeRate    float64
		expectedPrice   float64
	}{
		{
			name:            "ATM put",
			underlyingPrice: 100.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,
			riskFreeRate:    0.05,
			expectedPrice:   5.57, // Approximate
		},
		{
			name:            "ITM put",
			underlyingPrice: 90.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,
			riskFreeRate:    0.05,
			expectedPrice:   10.21, // Corrected Black-Scholes value
		},
		{
			name:            "OTM put",
			underlyingPrice: 110.0,
			strikePrice:     100.0,
			timeToExpiry:    1.0,
			volatility:      0.20,
			riskFreeRate:    0.05,
			expectedPrice:   2.79, // Corrected Black-Scholes value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := blackscholes.NewOption(
				"TEST",
				mechanisms.OptionTypePut,
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.strikePrice)),
				primitives.NewDecimalFromFloat(tt.timeToExpiry),
				primitives.MustPrice(primitives.NewDecimalFromFloat(1.0)),
				primitives.NewDecimalFromFloat(1.0),
			)
			if err != nil {
				t.Fatalf("Failed to create option: %v", err)
			}

			params := mechanisms.PriceParams{
				UnderlyingPrice: primitives.MustPrice(primitives.NewDecimalFromFloat(tt.underlyingPrice)),
				Volatility:      primitives.NewDecimalFromFloat(tt.volatility),
				RiskFreeRate:    primitives.NewDecimalFromFloat(tt.riskFreeRate),
				TimeToExpiry:    primitives.NewDecimalFromFloat(tt.timeToExpiry),
			}

			price, err := option.Price(context.Background(), params)
			if err != nil {
				t.Fatalf("Failed to price option: %v", err)
			}

			actualPrice := price.Decimal().Float64()
			percentDiff := math.Abs(actualPrice-tt.expectedPrice) / tt.expectedPrice

			if percentDiff > priceTolerance {
				t.Errorf("Price mismatch: expected %.4f, got %.4f (%.2f%% difference)",
					tt.expectedPrice, actualPrice, percentDiff*100)
			}
		})
	}
}

// TestOptionGreeks tests the calculation of option Greeks.
func TestOptionGreeks(t *testing.T) {
	// Test ATM call option Greeks
	t.Run("ATM Call Greeks", func(t *testing.T) {
		option, err := blackscholes.NewOption(
			"TEST",
			mechanisms.OptionTypeCall,
			primitives.MustPrice(primitives.NewDecimal(100)),
			primitives.NewDecimalFromFloat(1.0),
			primitives.MustPrice(primitives.NewDecimal(10)),
			primitives.NewDecimalFromFloat(1.0),
		)
		if err != nil {
			t.Fatalf("Failed to create option: %v", err)
		}

		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.MustPrice(primitives.NewDecimal(100)),
			Volatility:      primitives.NewDecimalFromFloat(0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
			TimeToExpiry:    primitives.NewDecimalFromFloat(1.0),
		}

		greeks, err := option.Greeks(context.Background(), params)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// For ATM options with these parameters:
		// Delta should be around 0.64
		// Gamma should be positive and around 0.02
		// Vega should be positive
		// Theta should be negative
		delta := greeks.Delta.Float64()
		if delta < 0.5 || delta > 0.7 {
			t.Errorf("Delta out of expected range: %.4f (expected 0.5-0.7)", delta)
		}

		gamma := greeks.Gamma.Float64()
		if gamma <= 0 {
			t.Error("Gamma should be positive for long options")
		}

		vega := greeks.Vega.Float64()
		if vega <= 0 {
			t.Error("Vega should be positive for long options")
		}

		theta := greeks.Theta.Float64()
		if theta >= 0 {
			t.Error("Theta should be negative for long options")
		}
	})

	// Test ATM put option Greeks
	t.Run("ATM Put Greeks", func(t *testing.T) {
		option, err := blackscholes.NewOption(
			"TEST",
			mechanisms.OptionTypePut,
			primitives.MustPrice(primitives.NewDecimal(100)),
			primitives.NewDecimalFromFloat(1.0),
			primitives.MustPrice(primitives.NewDecimal(10)),
			primitives.NewDecimalFromFloat(1.0),
		)
		if err != nil {
			t.Fatalf("Failed to create option: %v", err)
		}

		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.MustPrice(primitives.NewDecimal(100)),
			Volatility:      primitives.NewDecimalFromFloat(0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
			TimeToExpiry:    primitives.NewDecimalFromFloat(1.0),
		}

		greeks, err := option.Greeks(context.Background(), params)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// For ATM put options:
		// Delta should be around -0.36 (negative for puts)
		delta := greeks.Delta.Float64()
		if delta > -0.25 || delta < -0.5 {
			t.Errorf("Delta out of expected range: %.4f (expected -0.5 to -0.25)", delta)
		}

		// Gamma should be positive (same as call for ATM)
		gamma := greeks.Gamma.Float64()
		if gamma <= 0 {
			t.Error("Gamma should be positive for long options")
		}
	})
}

// TestIntrinsicValue tests the intrinsic value calculation at expiry.
func TestIntrinsicValue(t *testing.T) {
	tests := []struct {
		name            string
		optionType      mechanisms.OptionType
		strikePrice     float64
		underlyingPrice float64
		expectedValue   float64
	}{
		{
			name:            "ITM call",
			optionType:      mechanisms.OptionTypeCall,
			strikePrice:     100.0,
			underlyingPrice: 110.0,
			expectedValue:   10.0,
		},
		{
			name:            "OTM call",
			optionType:      mechanisms.OptionTypeCall,
			strikePrice:     100.0,
			underlyingPrice: 90.0,
			expectedValue:   0.0,
		},
		{
			name:            "ITM put",
			optionType:      mechanisms.OptionTypePut,
			strikePrice:     100.0,
			underlyingPrice: 90.0,
			expectedValue:   10.0,
		},
		{
			name:            "OTM put",
			optionType:      mechanisms.OptionTypePut,
			strikePrice:     100.0,
			underlyingPrice: 110.0,
			expectedValue:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := blackscholes.NewOption(
				"TEST",
				tt.optionType,
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.strikePrice)),
				primitives.Zero(), // At expiry
				primitives.MustPrice(primitives.NewDecimal(10)),
				primitives.NewDecimalFromFloat(1.0),
			)
			if err != nil {
				t.Fatalf("Failed to create option: %v", err)
			}

			params := mechanisms.PriceParams{
				UnderlyingPrice: primitives.MustPrice(primitives.NewDecimalFromFloat(tt.underlyingPrice)),
				Volatility:      primitives.NewDecimalFromFloat(0.20),
				RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
				TimeToExpiry:    primitives.Zero(),
			}

			price, err := option.Price(context.Background(), params)
			if err != nil {
				t.Fatalf("Failed to price option: %v", err)
			}

			actualValue := price.Decimal().Float64()
			if math.Abs(actualValue-tt.expectedValue) > 0.01 {
				t.Errorf("Intrinsic value mismatch: expected %.2f, got %.2f",
					tt.expectedValue, actualValue)
			}
		})
	}
}

// TestPriceValidation tests input validation for pricing.
func TestPriceValidation(t *testing.T) {
	option, err := blackscholes.NewOption(
		"TEST",
		mechanisms.OptionTypeCall,
		primitives.MustPrice(primitives.NewDecimal(100)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.MustPrice(primitives.NewDecimal(10)),
		primitives.NewDecimalFromFloat(1.0),
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}

	t.Run("Zero underlying price", func(t *testing.T) {
		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.ZeroPrice(),
			Volatility:      primitives.NewDecimalFromFloat(0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
		}

		_, err := option.Price(context.Background(), params)
		if err != blackscholes.ErrInvalidUnderlying {
			t.Errorf("Expected ErrInvalidUnderlying but got: %v", err)
		}
	})

	t.Run("Negative volatility", func(t *testing.T) {
		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.MustPrice(primitives.NewDecimal(100)),
			Volatility:      primitives.NewDecimalFromFloat(-0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
		}

		_, err := option.Price(context.Background(), params)
		if err != blackscholes.ErrInvalidVolatility {
			t.Errorf("Expected ErrInvalidVolatility but got: %v", err)
		}
	})
}

// TestMechanismInterface tests that Option implements the Derivative interface.
func TestMechanismInterface(t *testing.T) {
	option, err := blackscholes.NewOption(
		"TEST",
		mechanisms.OptionTypeCall,
		primitives.MustPrice(primitives.NewDecimal(100)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.MustPrice(primitives.NewDecimal(10)),
		primitives.NewDecimalFromFloat(1.0),
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}

	// Test Mechanism() returns correct type
	if option.Mechanism() != mechanisms.MechanismTypeDerivative {
		t.Errorf("Expected mechanism type %v but got %v",
			mechanisms.MechanismTypeDerivative, option.Mechanism())
	}

	// Test Venue() returns expected value
	if option.Venue() != "black-scholes" {
		t.Errorf("Expected venue 'black-scholes' but got %v", option.Venue())
	}
}

// TestSettleWithPrice tests settlement with final price.
func TestSettleWithPrice(t *testing.T) {
	tests := []struct {
		name         string
		optionType   mechanisms.OptionType
		strikePrice  float64
		entryPrice   float64
		positionSize float64
		finalPrice   float64
	}{
		{
			name:         "Profitable ITM call",
			optionType:   mechanisms.OptionTypeCall,
			strikePrice:  100.0,
			entryPrice:   10.0,
			positionSize: 1.0,
			finalPrice:   110.0,
		},
		{
			name:         "Losing OTM call",
			optionType:   mechanisms.OptionTypeCall,
			strikePrice:  100.0,
			entryPrice:   10.0,
			positionSize: 1.0,
			finalPrice:   90.0,
		},
		{
			name:         "Profitable ITM put",
			optionType:   mechanisms.OptionTypePut,
			strikePrice:  100.0,
			entryPrice:   8.0,
			positionSize: 1.0,
			finalPrice:   90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option, err := blackscholes.NewOption(
				"TEST",
				tt.optionType,
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.strikePrice)),
				primitives.NewDecimalFromFloat(1.0),
				primitives.MustPrice(primitives.NewDecimalFromFloat(tt.entryPrice)),
				primitives.NewDecimalFromFloat(tt.positionSize),
			)
			if err != nil {
				t.Fatalf("Failed to create option: %v", err)
			}

			if option.IsSettled() {
				t.Error("Option should not be settled initially")
			}

			pnl, err := option.SettleWithPrice(primitives.MustPrice(primitives.NewDecimalFromFloat(tt.finalPrice)))
			if err != nil {
				t.Fatalf("Failed to settle option: %v", err)
			}

			if pnl.Decimal().IsNegative() {
				t.Error("Settlement amount should not be negative")
			}

			if !option.IsSettled() {
				t.Error("Option should be marked as settled")
			}

			// Try to settle again (should fail)
			_, err = option.SettleWithPrice(primitives.MustPrice(primitives.NewDecimalFromFloat(tt.finalPrice)))
			if err == nil {
				t.Error("Expected error when settling already settled option")
			}
		})
	}
}

// TestGetters tests the getter methods.
func TestGetters(t *testing.T) {
	optionID := "TEST-CALL-100"
	strikePrice := primitives.MustPrice(primitives.NewDecimal(100))
	timeToExpiry := primitives.NewDecimalFromFloat(1.0)

	option, err := blackscholes.NewOption(
		optionID,
		mechanisms.OptionTypeCall,
		strikePrice,
		timeToExpiry,
		primitives.MustPrice(primitives.NewDecimal(10)),
		primitives.NewDecimalFromFloat(1.0),
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}

	if option.OptionID() != optionID {
		t.Errorf("Expected optionID %q but got %q", optionID, option.OptionID())
	}

	if option.OptionType() != mechanisms.OptionTypeCall {
		t.Errorf("Expected option type Call but got %v", option.OptionType())
	}

	if !option.StrikePrice().Equal(strikePrice) {
		t.Errorf("Expected strike price %v but got %v", strikePrice, option.StrikePrice())
	}

	if !option.TimeToExpiry().Equal(timeToExpiry) {
		t.Errorf("Expected time to expiry %v but got %v", timeToExpiry, option.TimeToExpiry())
	}
}

// TestGreeksValidation tests input validation for Greeks calculation.
func TestGreeksValidation(t *testing.T) {
	option, err := blackscholes.NewOption(
		"TEST",
		mechanisms.OptionTypeCall,
		primitives.MustPrice(primitives.NewDecimal(100)),
		primitives.NewDecimalFromFloat(1.0),
		primitives.MustPrice(primitives.NewDecimal(10)),
		primitives.NewDecimalFromFloat(1.0),
	)
	if err != nil {
		t.Fatalf("Failed to create option: %v", err)
	}

	t.Run("Zero underlying price", func(t *testing.T) {
		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.ZeroPrice(),
			Volatility:      primitives.NewDecimalFromFloat(0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
		}

		_, err := option.Greeks(context.Background(), params)
		if err != blackscholes.ErrInvalidUnderlying {
			t.Errorf("Expected ErrInvalidUnderlying but got: %v", err)
		}
	})

	t.Run("Negative volatility", func(t *testing.T) {
		params := mechanisms.PriceParams{
			UnderlyingPrice: primitives.MustPrice(primitives.NewDecimal(100)),
			Volatility:      primitives.NewDecimalFromFloat(-0.20),
			RiskFreeRate:    primitives.NewDecimalFromFloat(0.05),
		}

		_, err := option.Greeks(context.Background(), params)
		if err != blackscholes.ErrInvalidVolatility {
			t.Errorf("Expected ErrInvalidVolatility but got: %v", err)
		}
	})
}
