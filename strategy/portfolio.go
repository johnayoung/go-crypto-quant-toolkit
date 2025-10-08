package strategy

import (
	"fmt"
	"sync"

	"github.com/johnayoung/go-crypto-quant-toolkit/primitives"
)

// Portfolio manages a collection of positions and cash holdings.
// It provides position tracking, value queries, and cash management
// while remaining mechanism-agnostic.
//
// Thread Safety: Portfolio is safe for concurrent reads but not concurrent writes.
// If multiple goroutines need to modify a portfolio, external synchronization
// is required. Read operations (Value, GetPosition, Positions, Cash) are safe
// when no writes are occurring.
//
// Design: Portfolio is intentionally simple and doesn't prescribe strategy logic.
// It's a data structure for tracking positions, not a strategy coordinator.
type Portfolio struct {
	// mu protects concurrent access to positions and cash
	mu sync.RWMutex

	// positions maps position ID to Position
	// Using a map allows O(1) lookups and prevents duplicate IDs
	positions map[string]Position

	// cash tracks the current cash balance in the portfolio's denomination currency as a Decimal
	// (can be negative to represent borrowed funds/leverage)
	cashDecimal primitives.Decimal
}

// NewPortfolio creates a new empty portfolio with the specified initial cash.
func NewPortfolio(initialCash primitives.Amount) *Portfolio {
	return &Portfolio{
		positions:   make(map[string]Position),
		cashDecimal: initialCash.Decimal(),
	}
}

// AddPosition adds a position to the portfolio.
// Returns error if a position with the same ID already exists.
func (p *Portfolio) AddPosition(position Position) error {
	if position == nil {
		return ErrNilPosition
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	id := position.ID()
	if _, exists := p.positions[id]; exists {
		return fmt.Errorf("position %s already exists", id)
	}

	p.positions[id] = position
	return nil
}

// RemovePosition removes a position from the portfolio by ID.
// Returns error if the position is not found.
func (p *Portfolio) RemovePosition(positionID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.positions[positionID]; !exists {
		return fmt.Errorf("%w: %s", ErrPositionNotFound, positionID)
	}

	delete(p.positions, positionID)
	return nil
}

// GetPosition retrieves a position by ID.
// Returns error if the position is not found.
func (p *Portfolio) GetPosition(positionID string) (Position, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	position, exists := p.positions[positionID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrPositionNotFound, positionID)
	}

	return position, nil
}

// HasPosition returns true if a position with the given ID exists.
func (p *Portfolio) HasPosition(positionID string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, exists := p.positions[positionID]
	return exists
}

// Positions returns all positions in the portfolio.
// The returned slice is a snapshot and safe to iterate over.
// Modifications to the slice do not affect the portfolio.
func (p *Portfolio) Positions() []Position {
	p.mu.RLock()
	defer p.mu.RUnlock()

	positions := make([]Position, 0, len(p.positions))
	for _, pos := range p.positions {
		positions = append(positions, pos)
	}

	return positions
}

// PositionsByType returns all positions of the given type.
func (p *Portfolio) PositionsByType(posType PositionType) []Position {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var positions []Position
	for _, pos := range p.positions {
		if pos.Type() == posType {
			positions = append(positions, pos)
		}
	}

	return positions
}

// PositionCount returns the number of positions in the portfolio.
func (p *Portfolio) PositionCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.positions)
}

// Cash returns the current cash balance as an Amount (always non-negative).
// If cash is negative (debt), this returns zero. Use CashDecimal() for signed value.
func (p *Portfolio) Cash() primitives.Amount {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.cashDecimal.IsNegative() {
		return primitives.ZeroAmount()
	}
	// Safe to use Must because we checked it's not negative
	return primitives.MustAmount(p.cashDecimal)
}

// CashDecimal returns the current cash balance as a Decimal (can be negative for debt).
func (p *Portfolio) CashDecimal() primitives.Decimal {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.cashDecimal
}

// AdjustCash adds or removes cash from the portfolio.
// Positive values add cash, negative values remove cash.
// Negative cash balance represents borrowed funds / leverage.
func (p *Portfolio) AdjustCash(delta primitives.Decimal) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cashDecimal = p.cashDecimal.Add(delta)
	return nil
}

// SetCash sets the cash balance to a specific amount.
// Useful for initialization and testing.
func (p *Portfolio) SetCash(amount primitives.Amount) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.cashDecimal = amount.Decimal()
}

// Value returns the total value of the portfolio (positions + cash)
// using prices from the provided market snapshot.
//
// If any position fails to calculate its value, the error is returned
// and the total value calculation is aborted.
func (p *Portfolio) Value(snapshot MarketSnapshot) (primitives.Amount, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalValueDecimal := p.cashDecimal

	for id, position := range p.positions {
		posValue, err := position.Value(snapshot)
		if err != nil {
			return primitives.Amount{}, fmt.Errorf("failed to value position %s: %w", id, err)
		}
		totalValueDecimal = totalValueDecimal.Add(posValue.Decimal())
	}

	// Convert to Amount - if negative total, return zero (portfolio is underwater)
	if totalValueDecimal.IsNegative() {
		return primitives.ZeroAmount(), nil
	}
	return primitives.MustAmount(totalValueDecimal), nil
}

// PositionsValue returns the total value of all positions (excluding cash).
func (p *Portfolio) PositionsValue(snapshot MarketSnapshot) (primitives.Amount, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalValue := primitives.ZeroAmount()

	for id, position := range p.positions {
		posValue, err := position.Value(snapshot)
		if err != nil {
			return primitives.Amount{}, fmt.Errorf("failed to value position %s: %w", id, err)
		}
		totalValue = totalValue.Add(posValue)
	}

	return totalValue, nil
}

// Clone creates a deep copy of the portfolio.
// The cloned portfolio has independent position and cash state.
// Note: Positions themselves are not cloned (they should be immutable).
func (p *Portfolio) Clone() *Portfolio {
	p.mu.RLock()
	defer p.mu.RUnlock()

	positions := make(map[string]Position, len(p.positions))
	for id, pos := range p.positions {
		positions[id] = pos
	}

	return &Portfolio{
		positions:   positions,
		cashDecimal: p.cashDecimal,
	}
}

// Clear removes all positions and resets cash to zero.
// Useful for testing and resetting portfolio state.
func (p *Portfolio) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.positions = make(map[string]Position)
	p.cashDecimal = primitives.Zero()
}

// Summary returns a human-readable summary of the portfolio.
// Includes position count, cash balance, and total value if snapshot provided.
func (p *Portfolio) Summary(snapshot MarketSnapshot) string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	summary := fmt.Sprintf("Portfolio: %d positions, Cash: %s",
		len(p.positions), p.cashDecimal.String())

	if snapshot != nil {
		// Unlock for Value calculation (which also takes lock)
		p.mu.RUnlock()
		totalValue, err := p.Value(snapshot)
		p.mu.RLock()

		if err == nil {
			summary += fmt.Sprintf(", Total Value: %s", totalValue.String())
		}
	}

	return summary
}
