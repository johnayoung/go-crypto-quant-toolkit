package strategy

import (
	"context"
	"errors"
	"testing"

	"github.com/johnayoung/go-crypto-quant-toolkit/primitives"
)

// mockPosition is a test implementation of the Position interface
type mockPosition struct {
	id       string
	posType  PositionType
	value    primitives.Amount
	valueErr error
	risk     RiskMetrics
	riskErr  error
	withRisk bool
	venue    string
	desc     string
	withMeta bool
}

func (m *mockPosition) ID() string {
	return m.id
}

func (m *mockPosition) Type() PositionType {
	return m.posType
}

func (m *mockPosition) Value(snapshot MarketSnapshot) (primitives.Amount, error) {
	if m.valueErr != nil {
		return primitives.ZeroAmount(), m.valueErr
	}
	return m.value, nil
}

func (m *mockPosition) Risk(snapshot MarketSnapshot) (RiskMetrics, error) {
	if !m.withRisk {
		return RiskMetrics{}, errors.New("risk not implemented")
	}
	if m.riskErr != nil {
		return RiskMetrics{}, m.riskErr
	}
	return m.risk, nil
}

func (m *mockPosition) Description() string {
	if !m.withMeta {
		return ""
	}
	return m.desc
}

func (m *mockPosition) Venue() string {
	if !m.withMeta {
		return ""
	}
	return m.venue
}

// TestNewPortfolio tests portfolio creation
func TestNewPortfolio(t *testing.T) {
	tests := []struct {
		name        string
		initialCash primitives.Amount
	}{
		{
			name:        "zero cash",
			initialCash: primitives.ZeroAmount(),
		},
		{
			name:        "positive cash",
			initialCash: primitives.MustAmount(primitives.NewDecimal(10000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPortfolio(tt.initialCash)
			if p == nil {
				t.Fatal("expected non-nil portfolio")
			}
			if !p.Cash().Equal(tt.initialCash) {
				t.Errorf("cash = %v, want %v", p.Cash(), tt.initialCash)
			}
			if p.PositionCount() != 0 {
				t.Errorf("position count = %d, want 0", p.PositionCount())
			}
		})
	}
}

// TestPortfolioAddPosition tests adding positions
func TestPortfolioAddPosition(t *testing.T) {
	tests := []struct {
		name      string
		positions []*mockPosition
		wantErr   bool
		errMsg    string
	}{
		{
			name: "add single position",
			positions: []*mockPosition{
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				},
			},
			wantErr: false,
		},
		{
			name: "add multiple positions",
			positions: []*mockPosition{
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				},
				{
					id:      "pos2",
					posType: PositionTypeLiquidityPool,
					value:   primitives.MustAmount(primitives.NewDecimal(2000)),
				},
			},
			wantErr: false,
		},
		{
			name: "add duplicate position ID",
			positions: []*mockPosition{
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				},
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(1000)),
				},
			},
			wantErr: true,
			errMsg:  "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPortfolio(primitives.ZeroAmount())

			var err error
			for _, pos := range tt.positions {
				err = p.AddPosition(pos)
				if err != nil && !tt.wantErr {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
			} else {
				if p.PositionCount() != len(tt.positions) {
					t.Errorf("position count = %d, want %d", p.PositionCount(), len(tt.positions))
				}
			}
		})
	}
}

// TestPortfolioAddPositionNil tests adding nil position
func TestPortfolioAddPositionNil(t *testing.T) {
	p := NewPortfolio(primitives.ZeroAmount())
	err := p.AddPosition(nil)
	if err == nil {
		t.Fatal("expected error when adding nil position")
	}
	if !errors.Is(err, ErrNilPosition) {
		t.Errorf("error = %v, want %v", err, ErrNilPosition)
	}
}

// TestPortfolioRemovePosition tests removing positions
func TestPortfolioRemovePosition(t *testing.T) {
	pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
	pos2 := &mockPosition{id: "pos2", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(2000))}

	tests := []struct {
		name       string
		setup      func(*Portfolio)
		removeID   string
		wantErr    bool
		wantCount  int
		errContain string
	}{
		{
			name: "remove existing position",
			setup: func(p *Portfolio) {
				_ = p.AddPosition(pos1)
			},
			removeID:  "pos1",
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "remove non-existent position",
			setup: func(p *Portfolio) {
				_ = p.AddPosition(pos1)
			},
			removeID:   "pos2",
			wantErr:    true,
			wantCount:  1,
			errContain: "not found",
		},
		{
			name: "remove one of multiple",
			setup: func(p *Portfolio) {
				_ = p.AddPosition(pos1)
				_ = p.AddPosition(pos2)
			},
			removeID:  "pos1",
			wantErr:   false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPortfolio(primitives.ZeroAmount())
			tt.setup(p)

			err := p.RemovePosition(tt.removeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemovePosition() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errContain != "" && !contains(err.Error(), tt.errContain) {
				t.Errorf("error = %v, want substring %q", err, tt.errContain)
			}
			if p.PositionCount() != tt.wantCount {
				t.Errorf("position count = %d, want %d", p.PositionCount(), tt.wantCount)
			}
		})
	}
}

// TestPortfolioGetPosition tests retrieving positions
func TestPortfolioGetPosition(t *testing.T) {
	pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}

	p := NewPortfolio(primitives.ZeroAmount())
	_ = p.AddPosition(pos1)

	t.Run("get existing position", func(t *testing.T) {
		got, err := p.GetPosition("pos1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID() != "pos1" {
			t.Errorf("got position ID %q, want %q", got.ID(), "pos1")
		}
	})

	t.Run("get non-existent position", func(t *testing.T) {
		_, err := p.GetPosition("pos2")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, ErrPositionNotFound) {
			t.Errorf("error = %v, want %v", err, ErrPositionNotFound)
		}
	})
}

// TestPortfolioHasPosition tests position existence check
func TestPortfolioHasPosition(t *testing.T) {
	pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}

	p := NewPortfolio(primitives.ZeroAmount())
	_ = p.AddPosition(pos1)

	if !p.HasPosition("pos1") {
		t.Error("HasPosition(pos1) = false, want true")
	}
	if p.HasPosition("pos2") {
		t.Error("HasPosition(pos2) = true, want false")
	}
}

// TestPortfolioPositions tests getting all positions
func TestPortfolioPositions(t *testing.T) {
	pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
	pos2 := &mockPosition{id: "pos2", posType: PositionTypeLiquidityPool, value: primitives.MustAmount(primitives.NewDecimal(2000))}

	p := NewPortfolio(primitives.ZeroAmount())
	_ = p.AddPosition(pos1)
	_ = p.AddPosition(pos2)

	positions := p.Positions()
	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2", len(positions))
	}

	// Verify positions slice is a copy (mutation doesn't affect portfolio)
	positions[0] = nil
	if p.PositionCount() != 2 {
		t.Error("modifying returned slice affected portfolio")
	}
}

// TestPortfolioPositionsByType tests filtering positions by type
func TestPortfolioPositionsByType(t *testing.T) {
	pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
	pos2 := &mockPosition{id: "pos2", posType: PositionTypeLiquidityPool, value: primitives.MustAmount(primitives.NewDecimal(2000))}
	pos3 := &mockPosition{id: "pos3", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(3000))}

	p := NewPortfolio(primitives.ZeroAmount())
	_ = p.AddPosition(pos1)
	_ = p.AddPosition(pos2)
	_ = p.AddPosition(pos3)

	spotPositions := p.PositionsByType(PositionTypeSpot)
	if len(spotPositions) != 2 {
		t.Errorf("got %d spot positions, want 2", len(spotPositions))
	}

	lpPositions := p.PositionsByType(PositionTypeLiquidityPool)
	if len(lpPositions) != 1 {
		t.Errorf("got %d LP positions, want 1", len(lpPositions))
	}

	optionPositions := p.PositionsByType(PositionTypeOption)
	if len(optionPositions) != 0 {
		t.Errorf("got %d option positions, want 0", len(optionPositions))
	}
}

// TestPortfolioCashOperations tests cash management
func TestPortfolioCashOperations(t *testing.T) {
	initialCash := primitives.MustAmount(primitives.NewDecimal(10000))

	tests := []struct {
		name      string
		initial   primitives.Amount
		operation func(*Portfolio) error
		wantCash  primitives.Amount
		wantErr   bool
	}{
		{
			name:    "adjust cash positive",
			initial: initialCash,
			operation: func(p *Portfolio) error {
				return p.AdjustCash(primitives.NewDecimal(5000))
			},
			wantCash: primitives.MustAmount(primitives.NewDecimal(15000)),
			wantErr:  false,
		},
		{
			name:    "adjust cash negative",
			initial: initialCash,
			operation: func(p *Portfolio) error {
				return p.AdjustCash(primitives.NewDecimal(-3000))
			},
			wantCash: primitives.MustAmount(primitives.NewDecimal(7000)),
			wantErr:  false,
		},
		{
			name:    "set cash",
			initial: initialCash,
			operation: func(p *Portfolio) error {
				p.SetCash(primitives.MustAmount(primitives.NewDecimal(20000)))
				return nil
			},
			wantCash: primitives.MustAmount(primitives.NewDecimal(20000)),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPortfolio(tt.initial)
			err := tt.operation(p)
			if (err != nil) != tt.wantErr {
				t.Errorf("operation error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !p.Cash().Equal(tt.wantCash) {
				t.Errorf("cash = %v, want %v", p.Cash(), tt.wantCash)
			}
		})
	}
}

// TestPortfolioValue tests portfolio valuation
func TestPortfolioValue(t *testing.T) {
	prices := map[string]primitives.Price{
		"ETH/USD": primitives.MustPrice(primitives.NewDecimal(2000)),
		"BTC/USD": primitives.MustPrice(primitives.NewDecimal(50000)),
	}
	snapshot := NewSimpleSnapshot(primitives.Time{}, prices)

	tests := []struct {
		name      string
		cash      primitives.Amount
		positions []*mockPosition
		wantValue primitives.Amount
		wantErr   bool
	}{
		{
			name:      "cash only",
			cash:      primitives.MustAmount(primitives.NewDecimal(10000)),
			positions: nil,
			wantValue: primitives.MustAmount(primitives.NewDecimal(10000)),
			wantErr:   false,
		},
		{
			name: "cash and one position",
			cash: primitives.MustAmount(primitives.NewDecimal(10000)),
			positions: []*mockPosition{
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(5000)),
				},
			},
			wantValue: primitives.MustAmount(primitives.NewDecimal(15000)),
			wantErr:   false,
		},
		{
			name: "cash and multiple positions",
			cash: primitives.MustAmount(primitives.NewDecimal(10000)),
			positions: []*mockPosition{
				{
					id:      "pos1",
					posType: PositionTypeSpot,
					value:   primitives.MustAmount(primitives.NewDecimal(5000)),
				},
				{
					id:      "pos2",
					posType: PositionTypeLiquidityPool,
					value:   primitives.MustAmount(primitives.NewDecimal(3000)),
				},
			},
			wantValue: primitives.MustAmount(primitives.NewDecimal(18000)),
			wantErr:   false,
		},
		{
			name: "position value error",
			cash: primitives.MustAmount(primitives.NewDecimal(10000)),
			positions: []*mockPosition{
				{
					id:       "pos1",
					posType:  PositionTypeSpot,
					valueErr: errors.New("price not available"),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPortfolio(tt.cash)
			for _, pos := range tt.positions {
				_ = p.AddPosition(pos)
			}

			value, err := p.Value(snapshot)
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !value.Equal(tt.wantValue) {
				t.Errorf("value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

// TestPortfolioPositionsValue tests positions-only valuation
func TestPortfolioPositionsValue(t *testing.T) {
	snapshot := NewSimpleSnapshot(primitives.Time{}, nil)

	p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
	_ = p.AddPosition(&mockPosition{
		id:      "pos1",
		posType: PositionTypeSpot,
		value:   primitives.MustAmount(primitives.NewDecimal(5000)),
	})
	_ = p.AddPosition(&mockPosition{
		id:      "pos2",
		posType: PositionTypeLiquidityPool,
		value:   primitives.MustAmount(primitives.NewDecimal(3000)),
	})

	value, err := p.PositionsValue(snapshot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := primitives.MustAmount(primitives.NewDecimal(8000))
	if !value.Equal(want) {
		t.Errorf("positions value = %v, want %v (excluding cash)", value, want)
	}
}

// TestPortfolioClone tests portfolio cloning
func TestPortfolioClone(t *testing.T) {
	p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
	_ = p.AddPosition(&mockPosition{
		id:      "pos1",
		posType: PositionTypeSpot,
		value:   primitives.MustAmount(primitives.NewDecimal(5000)),
	})

	clone := p.Clone()

	// Verify clone has same state
	if !clone.Cash().Equal(p.Cash()) {
		t.Errorf("clone cash = %v, want %v", clone.Cash(), p.Cash())
	}
	if clone.PositionCount() != p.PositionCount() {
		t.Errorf("clone position count = %d, want %d", clone.PositionCount(), p.PositionCount())
	}

	// Verify clone is independent
	_ = clone.AddPosition(&mockPosition{
		id:      "pos2",
		posType: PositionTypeSpot,
		value:   primitives.MustAmount(primitives.NewDecimal(3000)),
	})
	if p.PositionCount() == clone.PositionCount() {
		t.Error("modifying clone affected original portfolio")
	}
}

// TestPortfolioClear tests clearing portfolio
func TestPortfolioClear(t *testing.T) {
	p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
	_ = p.AddPosition(&mockPosition{
		id:      "pos1",
		posType: PositionTypeSpot,
		value:   primitives.MustAmount(primitives.NewDecimal(5000)),
	})

	p.Clear()

	if p.PositionCount() != 0 {
		t.Errorf("position count after clear = %d, want 0", p.PositionCount())
	}
	if !p.Cash().Equal(primitives.ZeroAmount()) {
		t.Errorf("cash after clear = %v, want %v", p.Cash(), primitives.ZeroAmount())
	}
}

// TestActions tests action implementations
func TestActions(t *testing.T) {
	t.Run("AddPositionAction", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		pos := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
		action := NewAddPositionAction(pos)

		if err := action.Apply(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !p.HasPosition("pos1") {
			t.Error("position not added")
		}
	})

	t.Run("AddPositionAction nil position", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		action := NewAddPositionAction(nil)

		if err := action.Apply(p); err == nil {
			t.Fatal("expected error for nil position")
		}
	})

	t.Run("AddPositionAction nil portfolio", func(t *testing.T) {
		pos := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
		action := NewAddPositionAction(pos)

		if err := action.Apply(nil); !errors.Is(err, ErrNilPortfolio) {
			t.Errorf("error = %v, want %v", err, ErrNilPortfolio)
		}
	})

	t.Run("RemovePositionAction", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		pos := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
		_ = p.AddPosition(pos)

		action := NewRemovePositionAction("pos1")
		if err := action.Apply(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.HasPosition("pos1") {
			t.Error("position not removed")
		}
	})

	t.Run("RemovePositionAction non-existent", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		action := NewRemovePositionAction("pos1")

		if err := action.Apply(p); !errors.Is(err, ErrPositionNotFound) {
			t.Errorf("error = %v, want %v", err, ErrPositionNotFound)
		}
	})

	t.Run("ReplacePositionAction", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		oldPos := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
		newPos := &mockPosition{id: "pos2", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(2000))}
		_ = p.AddPosition(oldPos)

		action := NewReplacePositionAction("pos1", newPos)
		if err := action.Apply(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.HasPosition("pos1") {
			t.Error("old position still exists")
		}
		if !p.HasPosition("pos2") {
			t.Error("new position not added")
		}
	})

	t.Run("AdjustCashAction", func(t *testing.T) {
		p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
		action := NewAdjustCashAction(primitives.NewDecimal(5000), "test")

		if err := action.Apply(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := primitives.MustAmount(primitives.NewDecimal(15000))
		if !p.Cash().Equal(want) {
			t.Errorf("cash = %v, want %v", p.Cash(), want)
		}
	})

	t.Run("BatchAction", func(t *testing.T) {
		p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
		pos1 := &mockPosition{id: "pos1", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(1000))}
		pos2 := &mockPosition{id: "pos2", posType: PositionTypeSpot, value: primitives.MustAmount(primitives.NewDecimal(2000))}

		batch := NewBatchAction(
			NewAddPositionAction(pos1),
			NewAddPositionAction(pos2),
			NewAdjustCashAction(primitives.NewDecimal(-1000), "fee"),
		)

		if err := batch.Apply(p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.PositionCount() != 2 {
			t.Errorf("position count = %d, want 2", p.PositionCount())
		}
		wantCash := primitives.MustAmount(primitives.NewDecimal(9000))
		if !p.Cash().Equal(wantCash) {
			t.Errorf("cash = %v, want %v", p.Cash(), wantCash)
		}
	})
}

// TestMarketSnapshot tests the SimpleSnapshot implementation
func TestMarketSnapshot(t *testing.T) {
	ethPrice := primitives.MustPrice(primitives.NewDecimal(2000))
	btcPrice := primitives.MustPrice(primitives.NewDecimal(50000))

	prices := map[string]primitives.Price{
		"ETH/USD": ethPrice,
		"BTC/USD": btcPrice,
	}

	snapshot := NewSimpleSnapshot(primitives.Time{}, prices)

	t.Run("Price existing pair", func(t *testing.T) {
		price, err := snapshot.Price("ETH/USD")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !price.Equal(ethPrice) {
			t.Errorf("price = %v, want %v", price, ethPrice)
		}
	})

	t.Run("Price non-existent pair", func(t *testing.T) {
		_, err := snapshot.Price("XRP/USD")
		if err == nil {
			t.Fatal("expected error for non-existent pair")
		}
		if !errors.Is(err, ErrPriceNotAvailable) {
			t.Errorf("error = %v, want %v", err, ErrPriceNotAvailable)
		}
	})

	t.Run("Prices", func(t *testing.T) {
		allPrices := snapshot.Prices()
		if len(allPrices) != 2 {
			t.Errorf("got %d prices, want 2", len(allPrices))
		}
	})

	t.Run("Get/Set metadata", func(t *testing.T) {
		snapshot.Set("test-key", "test-value")
		val, ok := snapshot.Get("test-key")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if val != "test-value" {
			t.Errorf("value = %v, want %q", val, "test-value")
		}

		_, ok = snapshot.Get("non-existent")
		if ok {
			t.Error("expected key to not exist")
		}
	})
}

// mockStrategyImpl is a test implementation of Strategy interface
type mockStrategyImpl struct{}

func (s *mockStrategyImpl) Rebalance(ctx context.Context, p *Portfolio, m MarketSnapshot) ([]Action, error) {
	return []Action{}, nil
}

// TestStrategyInterface tests that we can implement the Strategy interface
func TestStrategyInterface(t *testing.T) {
	s := &mockStrategyImpl{}

	// Verify it implements Strategy
	var _ Strategy = s

	actions, err := s.Rebalance(context.Background(), NewPortfolio(primitives.ZeroAmount()), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if actions == nil {
		t.Error("expected non-nil actions")
	}
}

// TestActionStrings tests String() methods on actions
func TestActionStrings(t *testing.T) {
	pos := &mockPosition{id: "test-pos", posType: PositionTypeSpot}

	tests := []struct {
		name   string
		action Action
		want   string
	}{
		{
			name:   "AddPositionAction",
			action: NewAddPositionAction(pos),
			want:   "test-pos",
		},
		{
			name:   "AddPositionAction nil",
			action: NewAddPositionAction(nil),
			want:   "nil",
		},
		{
			name:   "RemovePositionAction",
			action: NewRemovePositionAction("pos-id"),
			want:   "pos-id",
		},
		{
			name:   "ReplacePositionAction",
			action: NewReplacePositionAction("old-id", pos),
			want:   "old-id",
		},
		{
			name:   "AdjustCashAction with reason",
			action: NewAdjustCashAction(primitives.NewDecimal(1000), "deposit"),
			want:   "deposit",
		},
		{
			name:   "AdjustCashAction without reason",
			action: NewAdjustCashAction(primitives.NewDecimal(1000), ""),
			want:   "1000",
		},
		{
			name:   "BatchAction",
			action: NewBatchAction(NewAddPositionAction(pos)),
			want:   "1 actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.action.String()
			if !contains(s, tt.want) {
				t.Errorf("String() = %q, want substring %q", s, tt.want)
			}
		})
	}
}

// TestCashDecimal tests CashDecimal() method
func TestCashDecimal(t *testing.T) {
	t.Run("positive cash", func(t *testing.T) {
		p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
		cash := p.CashDecimal()
		if !cash.Equal(primitives.NewDecimal(10000)) {
			t.Errorf("CashDecimal() = %v, want %v", cash, primitives.NewDecimal(10000))
		}
	})

	t.Run("negative cash", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		_ = p.AdjustCash(primitives.NewDecimal(-5000))
		cash := p.CashDecimal()
		if !cash.Equal(primitives.NewDecimal(-5000)) {
			t.Errorf("CashDecimal() = %v, want %v", cash, primitives.NewDecimal(-5000))
		}
	})
}

// TestPortfolioSummary tests Summary() method
func TestPortfolioSummary(t *testing.T) {
	p := NewPortfolio(primitives.MustAmount(primitives.NewDecimal(10000)))
	_ = p.AddPosition(&mockPosition{
		id:      "pos1",
		posType: PositionTypeSpot,
		value:   primitives.MustAmount(primitives.NewDecimal(5000)),
	})

	t.Run("without snapshot", func(t *testing.T) {
		summary := p.Summary(nil)
		if !contains(summary, "1 positions") {
			t.Errorf("summary = %q, want '1 positions'", summary)
		}
		if !contains(summary, "10000") {
			t.Errorf("summary = %q, want cash '10000'", summary)
		}
	})

	t.Run("with snapshot", func(t *testing.T) {
		snapshot := NewSimpleSnapshot(primitives.Time{}, nil)
		summary := p.Summary(snapshot)
		if !contains(summary, "1 positions") {
			t.Errorf("summary = %q, want '1 positions'", summary)
		}
		if !contains(summary, "15000") {
			t.Errorf("summary = %q, want total value '15000'", summary)
		}
	})
}

// TestSnapshotTime tests Time() method on SimpleSnapshot
func TestSnapshotTime(t *testing.T) {
	now := primitives.Time{}
	snapshot := NewSimpleSnapshot(now, nil)

	gotTime := snapshot.Time()
	if gotTime != now {
		t.Errorf("Time() = %v, want %v", gotTime, now)
	}
}

// TestPortfolioEdgeCases tests edge cases and error conditions
func TestPortfolioEdgeCases(t *testing.T) {
	t.Run("negative cash from Cash()", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		_ = p.AdjustCash(primitives.NewDecimal(-1000))

		// Cash() should return zero for negative balance
		cash := p.Cash()
		if !cash.Equal(primitives.ZeroAmount()) {
			t.Errorf("Cash() with negative balance = %v, want zero", cash)
		}

		// But CashDecimal() should show negative
		cashDecimal := p.CashDecimal()
		if !cashDecimal.IsNegative() {
			t.Error("CashDecimal() should be negative")
		}
	})

	t.Run("value with negative total", func(t *testing.T) {
		p := NewPortfolio(primitives.ZeroAmount())
		_ = p.AdjustCash(primitives.NewDecimal(-5000))

		snapshot := NewSimpleSnapshot(primitives.Time{}, nil)
		value, err := p.Value(snapshot)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return zero for negative total value
		if !value.Equal(primitives.ZeroAmount()) {
			t.Errorf("Value() with negative total = %v, want zero", value)
		}
	})

	t.Run("action apply errors", func(t *testing.T) {
		// RemovePositionAction on empty ID
		p := NewPortfolio(primitives.ZeroAmount())
		action := NewRemovePositionAction("")
		if err := action.Apply(p); err == nil {
			t.Error("expected error for empty position ID")
		}

		// ReplacePositionAction with empty old ID
		action2 := NewReplacePositionAction("", &mockPosition{id: "new"})
		if err := action2.Apply(p); err == nil {
			t.Error("expected error for empty old position ID")
		}

		// ReplacePositionAction with nil new position
		_ = p.AddPosition(&mockPosition{id: "old", posType: PositionTypeSpot})
		action3 := NewReplacePositionAction("old", nil)
		if err := action3.Apply(p); err == nil {
			t.Error("expected error for nil new position")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsCheck(s, substr)))
}

func containsCheck(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
