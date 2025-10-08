package backtest_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/backtest"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

// mockStrategy implements strategy.Strategy for testing
type mockStrategy struct {
	rebalanceFunc func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error)
	callCount     int
}

func (m *mockStrategy) Rebalance(ctx context.Context, p *strategy.Portfolio, m2 strategy.MarketSnapshot) ([]strategy.Action, error) {
	m.callCount++
	if m.rebalanceFunc != nil {
		return m.rebalanceFunc(ctx, p, m2)
	}
	return nil, nil
}

// mockSnapshot implements strategy.MarketSnapshot for testing
type mockSnapshot struct {
	time   primitives.Time
	prices map[string]primitives.Price
	data   map[string]interface{}
}

func (m *mockSnapshot) Time() primitives.Time {
	return m.time
}

func (m *mockSnapshot) Price(pair string) (primitives.Price, error) {
	if price, ok := m.prices[pair]; ok {
		return price, nil
	}
	return primitives.ZeroPrice(), fmt.Errorf("price not found for %s", pair)
}

func (m *mockSnapshot) Prices() map[string]primitives.Price {
	return m.prices
}

func (m *mockSnapshot) Get(key string) (interface{}, bool) {
	val, ok := m.data[key]
	return val, ok
}

// mockPosition implements strategy.Position for testing
type mockPosition struct {
	id          string
	posType     strategy.PositionType
	value       primitives.Amount
	valueFunc   func(m strategy.MarketSnapshot) (primitives.Amount, error)
	riskMetrics strategy.RiskMetrics
}

func (m *mockPosition) ID() string {
	return m.id
}

func (m *mockPosition) Type() strategy.PositionType {
	return m.posType
}

func (m *mockPosition) Value(snap strategy.MarketSnapshot) (primitives.Amount, error) {
	if m.valueFunc != nil {
		return m.valueFunc(snap)
	}
	return m.value, nil
}

func (m *mockPosition) Risk(snap strategy.MarketSnapshot) (strategy.RiskMetrics, error) {
	return m.riskMetrics, nil
}

// Helper to create mock snapshots
func createMockSnapshots(count int, startTime time.Time, interval time.Duration) []strategy.MarketSnapshot {
	snapshots := make([]strategy.MarketSnapshot, count)

	for i := 0; i < count; i++ {
		t := primitives.NewTime(startTime.Add(time.Duration(i) * interval))
		// Price varies: 100, 105, 110, 115, 120...
		priceValue := primitives.NewDecimal(100 + int64(i*5))
		price := primitives.MustPrice(priceValue)

		snapshots[i] = &mockSnapshot{
			time: t,
			prices: map[string]primitives.Price{
				"ETH/USD": price,
			},
			data: make(map[string]interface{}),
		}
	}

	return snapshots
}

func TestEngineBasicExecution(t *testing.T) {
	// Test basic backtest execution with a simple strategy
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			// Simple strategy: do nothing
			return []strategy.Action{}, nil
		},
	}

	snapshots := createMockSnapshots(10, time.Now(), time.Hour)

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	// Verify strategy was called for each snapshot
	if strat.callCount != len(snapshots) {
		t.Errorf("expected strategy called %d times, got %d", len(snapshots), strat.callCount)
	}

	// Verify value history was recorded
	if len(result.ValueHistory) != len(snapshots) {
		t.Errorf("expected %d value points, got %d", len(snapshots), len(result.ValueHistory))
	}

	// Verify final value equals initial cash (no trades)
	if !result.FinalValue.Equal(config.InitialCash) {
		t.Errorf("expected final value %s, got %s", config.InitialCash, result.FinalValue)
	}
}

func TestEngineWithPositions(t *testing.T) {
	// Test backtest with a strategy that adds positions
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			// On first call, add a position
			if !p.HasPosition("test-position") {
				pos := &mockPosition{
					id:      "test-position",
					posType: strategy.PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				}
				return []strategy.Action{strategy.NewAddPositionAction(pos)}, nil
			}
			return []strategy.Action{}, nil
		},
	}

	snapshots := createMockSnapshots(5, time.Now(), time.Hour)

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify position was added
	positions := result.Portfolio.Positions()
	if len(positions) != 1 {
		t.Errorf("expected 1 position, got %d", len(positions))
	}

	// Verify final value includes position value
	expectedValue := config.InitialCash.Add(primitives.MustAmount(primitives.NewDecimal(1000)))
	if !result.FinalValue.Equal(expectedValue) {
		t.Errorf("expected final value %s, got %s", expectedValue, result.FinalValue)
	}
}

func TestEngineContextCancellation(t *testing.T) {
	// Test that engine respects context cancellation
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			// Check context in strategy (engine also checks)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}
			return []strategy.Action{}, nil
		},
	}

	snapshots := createMockSnapshots(100, time.Now(), time.Hour)

	// Create a pre-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(ctx, strat, snapshots)

	// Should get an error about cancellation
	if err == nil {
		t.Fatal("expected error due to cancellation, got nil")
	}

	// Result should be nil on cancellation
	if result != nil {
		t.Errorf("expected nil result on cancellation, got %v", result)
	}
}

func TestEngineStrategyError(t *testing.T) {
	// Test that engine handles strategy errors
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			return nil, fmt.Errorf("strategy error")
		},
	}

	snapshots := createMockSnapshots(5, time.Now(), time.Hour)

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)

	if err == nil {
		t.Fatal("expected error from strategy, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}
}

func TestEngineActionError(t *testing.T) {
	// Test that engine handles action application errors
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			// Try to add a position with duplicate ID
			pos := &mockPosition{
				id:      "test-position",
				posType: strategy.PositionTypeSpot,
				value:   primitives.MustAmount(primitives.NewDecimal(100)),
			}
			// Add the same position twice to cause an error
			return []strategy.Action{
				strategy.NewAddPositionAction(pos),
				strategy.NewAddPositionAction(pos),
			}, nil
		},
	}

	snapshots := createMockSnapshots(2, time.Now(), time.Hour)

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)

	if err == nil {
		t.Fatal("expected error from duplicate position, got nil")
	}

	if result != nil {
		t.Errorf("expected nil result on action error, got %v", result)
	}
}

func TestEngineValidation(t *testing.T) {
	t.Run("nil strategy", func(t *testing.T) {
		snapshots := createMockSnapshots(5, time.Now(), time.Hour)
		engine := backtest.NewEngineWithDefaults()

		result, err := engine.Run(context.Background(), nil, snapshots)

		if err == nil {
			t.Fatal("expected error for nil strategy")
		}
		if result != nil {
			t.Error("expected nil result")
		}
	})

	t.Run("empty snapshots", func(t *testing.T) {
		strat := &mockStrategy{}
		engine := backtest.NewEngineWithDefaults()

		result, err := engine.Run(context.Background(), strat, []strategy.MarketSnapshot{})

		if err == nil {
			t.Fatal("expected error for empty snapshots")
		}
		if result != nil {
			t.Error("expected nil result")
		}
	})
}

func TestResultMetrics(t *testing.T) {
	// Test that performance metrics are calculated correctly
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			return []strategy.Action{}, nil
		},
	}

	// Create snapshots with known value progression
	snapshots := createMockSnapshots(5, time.Now(), 24*time.Hour)

	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(10000)),
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify metrics were calculated
	if result.TotalReturn.IsZero() && !result.InitialValue.Equal(result.FinalValue) {
		t.Error("expected non-zero total return")
	}

	// Verify value history
	if len(result.ValueHistory) != len(snapshots) {
		t.Errorf("expected %d value points, got %d", len(snapshots), len(result.ValueHistory))
	}

	// Verify summary doesn't panic
	summary := result.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestResultWithGains(t *testing.T) {
	// Test backtest with strategy that generates gains
	callNum := 0
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			callNum++
			// On first call, add position worth 1000
			if callNum == 1 {
				pos := &mockPosition{
					id:      "appreciating-position",
					posType: strategy.PositionTypeSpot,
					valueFunc: func(m strategy.MarketSnapshot) (primitives.Amount, error) {
						// Position value increases with each snapshot (simulating appreciation)
						ethPrice, _ := m.Price("ETH/USD")
						// Return 10 * ETH price
						return primitives.MustAmount(ethPrice.Decimal().Mul(primitives.NewDecimal(10))), nil
					},
				}
				return []strategy.Action{strategy.NewAddPositionAction(pos)}, nil
			}
			return []strategy.Action{}, nil
		},
	}

	// Prices increase from 100 to 120 over 5 snapshots
	snapshots := createMockSnapshots(5, time.Now(), 24*time.Hour)

	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(10000)),
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify positive return (position appreciated)
	if !result.TotalReturn.IsPositive() {
		t.Errorf("expected positive return, got %s", result.TotalReturn.String())
	}

	// Verify final value > initial value
	if !result.FinalValue.GreaterThan(result.InitialValue) {
		t.Errorf("expected final value > initial value")
	}
}

func TestResultWithDrawdown(t *testing.T) {
	// Test max drawdown calculation
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			// Strategy that adjusts cash based on ETH price (simulating gains/losses)
			ethPrice, _ := m.Price("ETH/USD")
			priceFloat := ethPrice.Decimal().Float64()

			// Adjust cash: increase on low prices, decrease on high prices (creating drawdown)
			if priceFloat > 110 {
				// Lose money when price is high
				loss := primitives.MustAmount(primitives.NewDecimal(500))
				return []strategy.Action{strategy.NewAdjustCashAction(loss.Decimal().Neg(), "market loss")}, nil
			}
			return []strategy.Action{}, nil
		},
	}

	// Create snapshots with price going up (100 -> 120), causing portfolio drawdown
	snapshots := createMockSnapshots(5, time.Now(), 24*time.Hour)

	config := backtest.Config{
		InitialCash:           primitives.MustAmount(primitives.NewDecimal(10000)),
		EnableDetailedLogging: false,
	}
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify drawdown was calculated
	if result.MaxDrawdown.IsZero() {
		t.Error("expected non-zero max drawdown")
	}

	// Verify drawdown is positive (it's a decline)
	if result.MaxDrawdown.IsNegative() {
		t.Errorf("drawdown should not be negative, got %s", result.MaxDrawdown.String())
	}
}

func TestMultiMechanismStrategy(t *testing.T) {
	// Test that engine works with positions from multiple mechanism types
	callNum := 0
	strat := &mockStrategy{
		rebalanceFunc: func(ctx context.Context, p *strategy.Portfolio, m strategy.MarketSnapshot) ([]strategy.Action, error) {
			callNum++
			if callNum == 1 {
				// Add multiple position types
				spotPos := &mockPosition{
					id:      "spot-position",
					posType: strategy.PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				}
				lpPos := &mockPosition{
					id:      "lp-position",
					posType: strategy.PositionTypeLiquidityPool,
					value:   primitives.MustAmount(primitives.NewDecimal(2000)),
				}
				optionPos := &mockPosition{
					id:      "option-position",
					posType: strategy.PositionTypeOption,
					value:   primitives.MustAmount(primitives.NewDecimal(500)),
				}
				perpPos := &mockPosition{
					id:      "perp-position",
					posType: strategy.PositionTypePerpetual,
					value:   primitives.MustAmount(primitives.NewDecimal(1500)),
				}

				return []strategy.Action{
					strategy.NewAddPositionAction(spotPos),
					strategy.NewAddPositionAction(lpPos),
					strategy.NewAddPositionAction(optionPos),
					strategy.NewAddPositionAction(perpPos),
				}, nil
			}
			return []strategy.Action{}, nil
		},
	}

	// Use longer interval to avoid infinity in annualized return calculation
	snapshots := createMockSnapshots(30, time.Now(), 24*time.Hour)

	config := backtest.DefaultConfig()
	engine := backtest.NewEngine(config)

	result, err := engine.Run(context.Background(), strat, snapshots)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify all positions were added
	positions := result.Portfolio.Positions()
	if len(positions) != 4 {
		t.Errorf("expected 4 positions, got %d", len(positions))
	}

	// Verify each position type is present
	types := make(map[strategy.PositionType]bool)
	for _, pos := range positions {
		types[pos.Type()] = true
	}

	expectedTypes := []strategy.PositionType{
		strategy.PositionTypeSpot,
		strategy.PositionTypeLiquidityPool,
		strategy.PositionTypeOption,
		strategy.PositionTypePerpetual,
	}

	for _, pt := range expectedTypes {
		if !types[pt] {
			t.Errorf("expected position type %s to be present", pt)
		}
	}

	// Verify final value includes all positions
	expectedValue := config.InitialCash.
		Add(primitives.MustAmount(primitives.NewDecimal(1000))).
		Add(primitives.MustAmount(primitives.NewDecimal(2000))).
		Add(primitives.MustAmount(primitives.NewDecimal(500))).
		Add(primitives.MustAmount(primitives.NewDecimal(1500)))

	if !result.FinalValue.Equal(expectedValue) {
		t.Errorf("expected final value %s, got %s", expectedValue, result.FinalValue)
	}
}
