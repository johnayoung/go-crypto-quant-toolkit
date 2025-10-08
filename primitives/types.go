// Package primitives provides type-safe financial and temporal primitives
// used across all framework layers. All financial calculations use decimal
// arithmetic to prevent floating-point precision errors.
package primitives

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var (
	// ErrNegativePrice indicates an invalid negative price value
	ErrNegativePrice = errors.New("price cannot be negative")
	// ErrNegativeAmount indicates an invalid negative amount value
	ErrNegativeAmount = errors.New("amount cannot be negative")
	// ErrDivisionByZero indicates attempted division by zero
	ErrDivisionByZero = errors.New("division by zero")
	// ErrInvalidDecimal indicates an invalid decimal value
	ErrInvalidDecimal = errors.New("invalid decimal value")
)

// Decimal wraps shopspring/decimal.Decimal for precise arithmetic.
// Used as the base type for all financial calculations.
type Decimal struct {
	value decimal.Decimal
}

// NewDecimal creates a Decimal from an int64 value.
func NewDecimal(value int64) Decimal {
	return Decimal{value: decimal.NewFromInt(value)}
}

// NewDecimalFromFloat creates a Decimal from a float64 value.
// Note: Use this sparingly; prefer NewDecimalFromString for external data.
func NewDecimalFromFloat(value float64) Decimal {
	return Decimal{value: decimal.NewFromFloat(value)}
}

// NewDecimalFromString creates a Decimal from a string representation.
// Returns error if the string is not a valid decimal number.
func NewDecimalFromString(value string) (Decimal, error) {
	d, err := decimal.NewFromString(value)
	if err != nil {
		return Decimal{}, fmt.Errorf("%w: %s", ErrInvalidDecimal, err)
	}
	return Decimal{value: d}, nil
}

// MustDecimalFromString creates a Decimal from a string, panicking on error.
// Only use for known-valid constants in tests or initialization.
func MustDecimalFromString(value string) Decimal {
	d, err := NewDecimalFromString(value)
	if err != nil {
		panic(err)
	}
	return d
}

// Zero returns a Decimal representing zero.
func Zero() Decimal {
	return Decimal{value: decimal.Zero}
}

// One returns a Decimal representing one.
func One() Decimal {
	return Decimal{value: decimal.NewFromInt(1)}
}

// Add returns the sum of two Decimals.
func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{value: d.value.Add(other.value)}
}

// Sub returns the difference of two Decimals.
func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{value: d.value.Sub(other.value)}
}

// Mul returns the product of two Decimals.
func (d Decimal) Mul(other Decimal) Decimal {
	return Decimal{value: d.value.Mul(other.value)}
}

// Div returns the quotient of two Decimals.
// Returns error if dividing by zero.
func (d Decimal) Div(other Decimal) (Decimal, error) {
	if other.value.IsZero() {
		return Decimal{}, ErrDivisionByZero
	}
	return Decimal{value: d.value.Div(other.value)}, nil
}

// Abs returns the absolute value of the Decimal.
func (d Decimal) Abs() Decimal {
	return Decimal{value: d.value.Abs()}
}

// Neg returns the negation of the Decimal.
func (d Decimal) Neg() Decimal {
	return Decimal{value: d.value.Neg()}
}

// IsZero returns true if the Decimal is zero.
func (d Decimal) IsZero() bool {
	return d.value.IsZero()
}

// IsNegative returns true if the Decimal is negative.
func (d Decimal) IsNegative() bool {
	return d.value.IsNegative()
}

// IsPositive returns true if the Decimal is positive.
func (d Decimal) IsPositive() bool {
	return d.value.IsPositive()
}

// GreaterThan returns true if d > other.
func (d Decimal) GreaterThan(other Decimal) bool {
	return d.value.GreaterThan(other.value)
}

// LessThan returns true if d < other.
func (d Decimal) LessThan(other Decimal) bool {
	return d.value.LessThan(other.value)
}

// Equal returns true if d == other.
func (d Decimal) Equal(other Decimal) bool {
	return d.value.Equal(other.value)
}

// Float64 returns the float64 representation of the Decimal.
// Use only for display or external APIs; not for calculations.
func (d Decimal) Float64() float64 {
	f, _ := d.value.Float64()
	return f
}

// String returns the string representation of the Decimal.
func (d Decimal) String() string {
	return d.value.String()
}

// Price represents a unit price of an asset.
// Prices cannot be negative and support specific arithmetic operations
// that maintain type safety (e.g., Price * Amount = Amount, not Price).
type Price struct {
	value Decimal
}

// NewPrice creates a Price from a Decimal value.
// Returns error if the value is negative.
func NewPrice(value Decimal) (Price, error) {
	if value.IsNegative() {
		return Price{}, ErrNegativePrice
	}
	return Price{value: value}, nil
}

// MustPrice creates a Price from a Decimal, panicking if invalid.
// Only use for known-valid constants in tests or initialization.
func MustPrice(value Decimal) Price {
	p, err := NewPrice(value)
	if err != nil {
		panic(err)
	}
	return p
}

// ZeroPrice returns a Price representing zero.
func ZeroPrice() Price {
	return Price{value: Zero()}
}

// Decimal returns the underlying Decimal value.
func (p Price) Decimal() Decimal {
	return p.value
}

// Mul returns the product of a Price and a Decimal (e.g., price adjustment).
func (p Price) Mul(factor Decimal) Price {
	return Price{value: p.value.Mul(factor)}
}

// Div returns the quotient of a Price and a Decimal.
// Returns error if dividing by zero.
func (p Price) Div(divisor Decimal) (Price, error) {
	result, err := p.value.Div(divisor)
	if err != nil {
		return Price{}, err
	}
	return Price{value: result}, nil
}

// Add returns the sum of two Prices (e.g., combining quotes).
func (p Price) Add(other Price) Price {
	return Price{value: p.value.Add(other.value)}
}

// Sub returns the difference of two Prices.
// Returns error if the result would be negative.
func (p Price) Sub(other Price) (Price, error) {
	result := p.value.Sub(other.value)
	if result.IsNegative() {
		return Price{}, ErrNegativePrice
	}
	return Price{value: result}, nil
}

// GreaterThan returns true if p > other.
func (p Price) GreaterThan(other Price) bool {
	return p.value.GreaterThan(other.value)
}

// LessThan returns true if p < other.
func (p Price) LessThan(other Price) bool {
	return p.value.LessThan(other.value)
}

// Equal returns true if p == other.
func (p Price) Equal(other Price) bool {
	return p.value.Equal(other.value)
}

// IsZero returns true if the Price is zero.
func (p Price) IsZero() bool {
	return p.value.IsZero()
}

// String returns the string representation of the Price.
func (p Price) String() string {
	return p.value.String()
}

// Amount represents a quantity of an asset (tokens, shares, etc.).
// Amounts cannot be negative and support specific arithmetic operations.
type Amount struct {
	value Decimal
}

// NewAmount creates an Amount from a Decimal value.
// Returns error if the value is negative.
func NewAmount(value Decimal) (Amount, error) {
	if value.IsNegative() {
		return Amount{}, ErrNegativeAmount
	}
	return Amount{value: value}, nil
}

// MustAmount creates an Amount from a Decimal, panicking if invalid.
// Only use for known-valid constants in tests or initialization.
func MustAmount(value Decimal) Amount {
	a, err := NewAmount(value)
	if err != nil {
		panic(err)
	}
	return a
}

// ZeroAmount returns an Amount representing zero.
func ZeroAmount() Amount {
	return Amount{value: Zero()}
}

// Decimal returns the underlying Decimal value.
func (a Amount) Decimal() Decimal {
	return a.value
}

// Add returns the sum of two Amounts.
func (a Amount) Add(other Amount) Amount {
	return Amount{value: a.value.Add(other.value)}
}

// Sub returns the difference of two Amounts.
// Returns error if the result would be negative.
func (a Amount) Sub(other Amount) (Amount, error) {
	result := a.value.Sub(other.value)
	if result.IsNegative() {
		return Amount{}, ErrNegativeAmount
	}
	return Amount{value: result}, nil
}

// Mul returns the product of an Amount and a Decimal (e.g., scaling quantity).
func (a Amount) Mul(factor Decimal) Amount {
	return Amount{value: a.value.Mul(factor)}
}

// Div returns the quotient of an Amount and a Decimal.
// Returns error if dividing by zero.
func (a Amount) Div(divisor Decimal) (Amount, error) {
	result, err := a.value.Div(divisor)
	if err != nil {
		return Amount{}, err
	}
	return Amount{value: result}, nil
}

// MulPrice returns the value of multiplying an Amount by a Price (Amount * Price = Amount in different units).
func (a Amount) MulPrice(price Price) Amount {
	return Amount{value: a.value.Mul(price.value)}
}

// DivPrice returns the quotient of an Amount divided by a Price (Amount / Price = Amount in different units).
// Returns error if dividing by zero.
func (a Amount) DivPrice(price Price) (Amount, error) {
	if price.value.IsZero() {
		return Amount{}, ErrDivisionByZero
	}
	result, err := a.value.Div(price.value)
	if err != nil {
		return Amount{}, err
	}
	return Amount{value: result}, nil
}

// GreaterThan returns true if a > other.
func (a Amount) GreaterThan(other Amount) bool {
	return a.value.GreaterThan(other.value)
}

// LessThan returns true if a < other.
func (a Amount) LessThan(other Amount) bool {
	return a.value.LessThan(other.value)
}

// Equal returns true if a == other.
func (a Amount) Equal(other Amount) bool {
	return a.value.Equal(other.value)
}

// IsZero returns true if the Amount is zero.
func (a Amount) IsZero() bool {
	return a.value.IsZero()
}

// String returns the string representation of the Amount.
func (a Amount) String() string {
	return a.value.String()
}
