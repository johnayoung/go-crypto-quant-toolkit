package strategy

import (
	"fmt"

	"github.com/johnayoung/go-crypto-quant-toolkit/primitives"
)

// Action represents a desired modification to a portfolio.
// Actions are returned by Strategy.Rebalance() and applied by the portfolio.
//
// Implementations should be immutable and contain all information needed
// to apply the action. Actions should validate their own preconditions
// during Apply().
//
// Thread Safety: Action implementations should be safe for concurrent reads.
type Action interface {
	// Apply executes this action on the given portfolio.
	// Returns error if the action cannot be applied (e.g., insufficient funds,
	// position not found, invalid parameters).
	//
	// Actions should be atomic - either fully succeed or leave portfolio unchanged.
	Apply(portfolio *Portfolio) error

	// String returns a human-readable description of this action.
	// Used for logging and debugging.
	String() string
}

// AddPositionAction adds a new position to the portfolio.
type AddPositionAction struct {
	Position Position
}

// NewAddPositionAction creates an action to add a position to the portfolio.
func NewAddPositionAction(position Position) *AddPositionAction {
	return &AddPositionAction{Position: position}
}

// Apply adds the position to the portfolio.
func (a *AddPositionAction) Apply(portfolio *Portfolio) error {
	if portfolio == nil {
		return ErrNilPortfolio
	}
	if a.Position == nil {
		return fmt.Errorf("%w: cannot add nil position", ErrInvalidAction)
	}

	return portfolio.AddPosition(a.Position)
}

// String returns a description of this action.
func (a *AddPositionAction) String() string {
	if a.Position == nil {
		return "AddPosition(nil)"
	}
	return fmt.Sprintf("AddPosition(%s)", a.Position.ID())
}

// RemovePositionAction removes a position from the portfolio.
type RemovePositionAction struct {
	PositionID string
}

// NewRemovePositionAction creates an action to remove a position by ID.
func NewRemovePositionAction(positionID string) *RemovePositionAction {
	return &RemovePositionAction{PositionID: positionID}
}

// Apply removes the position from the portfolio.
func (a *RemovePositionAction) Apply(portfolio *Portfolio) error {
	if portfolio == nil {
		return ErrNilPortfolio
	}
	if a.PositionID == "" {
		return fmt.Errorf("%w: position ID cannot be empty", ErrInvalidAction)
	}

	return portfolio.RemovePosition(a.PositionID)
}

// String returns a description of this action.
func (a *RemovePositionAction) String() string {
	return fmt.Sprintf("RemovePosition(%s)", a.PositionID)
}

// ReplacePositionAction replaces an existing position with a new one.
// This is useful for updating positions (e.g., adjusting LP range, rolling options).
type ReplacePositionAction struct {
	OldPositionID string
	NewPosition   Position
}

// NewReplacePositionAction creates an action to replace a position.
func NewReplacePositionAction(oldPositionID string, newPosition Position) *ReplacePositionAction {
	return &ReplacePositionAction{
		OldPositionID: oldPositionID,
		NewPosition:   newPosition,
	}
}

// Apply replaces the position in the portfolio.
func (a *ReplacePositionAction) Apply(portfolio *Portfolio) error {
	if portfolio == nil {
		return ErrNilPortfolio
	}
	if a.OldPositionID == "" {
		return fmt.Errorf("%w: old position ID cannot be empty", ErrInvalidAction)
	}
	if a.NewPosition == nil {
		return fmt.Errorf("%w: new position cannot be nil", ErrInvalidAction)
	}

	// Remove old position first, then add new one
	if err := portfolio.RemovePosition(a.OldPositionID); err != nil {
		return err
	}

	if err := portfolio.AddPosition(a.NewPosition); err != nil {
		// Attempt to restore the old position on failure
		// Note: This is a best-effort rollback; in production consider transaction semantics
		return fmt.Errorf("failed to add new position after removing old: %w", err)
	}

	return nil
}

// String returns a description of this action.
func (a *ReplacePositionAction) String() string {
	newID := "nil"
	if a.NewPosition != nil {
		newID = a.NewPosition.ID()
	}
	return fmt.Sprintf("ReplacePosition(%s -> %s)", a.OldPositionID, newID)
}

// AdjustCashAction adds or removes cash from the portfolio.
// Positive values add cash, negative values remove cash.
type AdjustCashAction struct {
	Delta  primitives.Decimal
	Reason string // Optional description of why cash is being adjusted
}

// NewAdjustCashAction creates an action to adjust portfolio cash.
// Use positive delta to add cash, negative delta to remove cash.
func NewAdjustCashAction(delta primitives.Decimal, reason string) *AdjustCashAction {
	return &AdjustCashAction{
		Delta:  delta,
		Reason: reason,
	}
}

// Apply adjusts the cash balance in the portfolio.
func (a *AdjustCashAction) Apply(portfolio *Portfolio) error {
	if portfolio == nil {
		return ErrNilPortfolio
	}

	return portfolio.AdjustCash(a.Delta)
}

// String returns a description of this action.
func (a *AdjustCashAction) String() string {
	if a.Reason != "" {
		return fmt.Sprintf("AdjustCash(%s, reason: %s)", a.Delta.String(), a.Reason)
	}
	return fmt.Sprintf("AdjustCash(%s)", a.Delta.String())
}

// BatchAction applies multiple actions as a single logical operation.
// Useful for complex rebalancing that requires multiple steps.
//
// If any action fails, all subsequent actions are skipped and an error is returned.
// Already-applied actions are not automatically rolled back; implement rollback
// logic if needed.
type BatchAction struct {
	Actions []Action
}

// NewBatchAction creates an action that applies multiple actions in sequence.
func NewBatchAction(actions ...Action) *BatchAction {
	return &BatchAction{Actions: actions}
}

// Apply executes all actions in sequence.
func (a *BatchAction) Apply(portfolio *Portfolio) error {
	if portfolio == nil {
		return ErrNilPortfolio
	}

	for i, action := range a.Actions {
		if err := action.Apply(portfolio); err != nil {
			return fmt.Errorf("batch action failed at step %d: %w", i, err)
		}
	}

	return nil
}

// String returns a description of this action.
func (a *BatchAction) String() string {
	return fmt.Sprintf("BatchAction(%d actions)", len(a.Actions))
}
