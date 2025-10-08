package strategy

import "errors"

var (
	// ErrPriceNotAvailable indicates the requested price is not in the snapshot
	ErrPriceNotAvailable = errors.New("price not available for pair")

	// ErrPositionNotFound indicates the position was not found in the portfolio
	ErrPositionNotFound = errors.New("position not found")

	// ErrInsufficientCash indicates insufficient cash for an operation
	ErrInsufficientCash = errors.New("insufficient cash")

	// ErrInvalidAction indicates an action cannot be applied
	ErrInvalidAction = errors.New("invalid action")

	// ErrNilPortfolio indicates a nil portfolio was provided
	ErrNilPortfolio = errors.New("portfolio cannot be nil")

	// ErrNilPosition indicates a nil position was provided
	ErrNilPosition = errors.New("position cannot be nil")
)
