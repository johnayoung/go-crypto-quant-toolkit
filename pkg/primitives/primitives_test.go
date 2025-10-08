package primitives

import (
	"testing"
	"time"
)

// TestDecimal tests basic Decimal operations
func TestDecimal(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		d1 := NewDecimal(100)
		if d1.String() != "100" {
			t.Errorf("expected 100, got %s", d1.String())
		}

		d2 := NewDecimalFromFloat(123.45)
		if d2.Float64() != 123.45 {
			t.Errorf("expected 123.45, got %f", d2.Float64())
		}

		d3, err := NewDecimalFromString("999.99")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if d3.String() != "999.99" {
			t.Errorf("expected 999.99, got %s", d3.String())
		}

		_, err = NewDecimalFromString("invalid")
		if err == nil {
			t.Error("expected error for invalid string")
		}
	})

	t.Run("zero and one", func(t *testing.T) {
		z := Zero()
		if !z.IsZero() {
			t.Error("Zero() should return zero")
		}

		one := One()
		if one.String() != "1" {
			t.Errorf("One() should return 1, got %s", one.String())
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		a := NewDecimal(10)
		b := NewDecimal(3)

		sum := a.Add(b)
		if sum.String() != "13" {
			t.Errorf("10 + 3 should be 13, got %s", sum.String())
		}

		diff := a.Sub(b)
		if diff.String() != "7" {
			t.Errorf("10 - 3 should be 7, got %s", diff.String())
		}

		prod := a.Mul(b)
		if prod.String() != "30" {
			t.Errorf("10 * 3 should be 30, got %s", prod.String())
		}

		quot, err := a.Div(b)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		quotStr := quot.String()
		if len(quotStr) < 4 || quotStr[:4] != "3.33" { // Approximate check
			t.Errorf("10 / 3 should be ~3.333, got %s", quot.String())
		}

		_, err = a.Div(Zero())
		if err != ErrDivisionByZero {
			t.Error("dividing by zero should return ErrDivisionByZero")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		a := NewDecimal(10)
		b := NewDecimal(5)
		c := NewDecimal(10)

		if !a.GreaterThan(b) {
			t.Error("10 should be greater than 5")
		}

		if !b.LessThan(a) {
			t.Error("5 should be less than 10")
		}

		if !a.Equal(c) {
			t.Error("10 should equal 10")
		}
	})

	t.Run("sign operations", func(t *testing.T) {
		positive := NewDecimal(10)
		if !positive.IsPositive() {
			t.Error("10 should be positive")
		}

		negative := NewDecimal(-10)
		if !negative.IsNegative() {
			t.Error("-10 should be negative")
		}

		abs := negative.Abs()
		if !abs.Equal(positive) {
			t.Error("|-10| should equal 10")
		}

		neg := positive.Neg()
		if !neg.Equal(negative) {
			t.Error("-(10) should equal -10")
		}
	})
}

// TestPrice tests Price type operations
func TestPrice(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		p, err := NewPrice(NewDecimal(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.String() != "100" {
			t.Errorf("expected 100, got %s", p.String())
		}

		_, err = NewPrice(NewDecimal(-10))
		if err != ErrNegativePrice {
			t.Error("negative price should return ErrNegativePrice")
		}

		zp := ZeroPrice()
		if !zp.IsZero() {
			t.Error("ZeroPrice() should return zero price")
		}
	})

	t.Run("must price panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustPrice with negative should panic")
			}
		}()
		MustPrice(NewDecimal(-100))
	})

	t.Run("arithmetic", func(t *testing.T) {
		p1 := MustPrice(NewDecimal(100))
		p2 := MustPrice(NewDecimal(50))

		// Price + Price
		sum := p1.Add(p2)
		if sum.String() != "150" {
			t.Errorf("100 + 50 should be 150, got %s", sum.String())
		}

		// Price - Price
		diff, err := p1.Sub(p2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff.String() != "50" {
			t.Errorf("100 - 50 should be 50, got %s", diff.String())
		}

		// Price - Price (negative result)
		_, err = p2.Sub(p1)
		if err != ErrNegativePrice {
			t.Error("50 - 100 should return ErrNegativePrice")
		}

		// Price * Decimal
		doubled := p1.Mul(NewDecimal(2))
		if doubled.String() != "200" {
			t.Errorf("100 * 2 should be 200, got %s", doubled.String())
		}

		// Price / Decimal
		halved, err := p1.Div(NewDecimal(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if halved.String() != "50" {
			t.Errorf("100 / 2 should be 50, got %s", halved.String())
		}

		_, err = p1.Div(Zero())
		if err != ErrDivisionByZero {
			t.Error("dividing price by zero should return ErrDivisionByZero")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		p1 := MustPrice(NewDecimal(100))
		p2 := MustPrice(NewDecimal(50))
		p3 := MustPrice(NewDecimal(100))

		if !p1.GreaterThan(p2) {
			t.Error("100 should be greater than 50")
		}

		if !p2.LessThan(p1) {
			t.Error("50 should be less than 100")
		}

		if !p1.Equal(p3) {
			t.Error("100 should equal 100")
		}
	})
}

// TestAmount tests Amount type operations
func TestAmount(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		a, err := NewAmount(NewDecimal(100))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.String() != "100" {
			t.Errorf("expected 100, got %s", a.String())
		}

		_, err = NewAmount(NewDecimal(-10))
		if err != ErrNegativeAmount {
			t.Error("negative amount should return ErrNegativeAmount")
		}

		za := ZeroAmount()
		if !za.IsZero() {
			t.Error("ZeroAmount() should return zero amount")
		}
	})

	t.Run("must amount panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustAmount with negative should panic")
			}
		}()
		MustAmount(NewDecimal(-100))
	})

	t.Run("arithmetic", func(t *testing.T) {
		a1 := MustAmount(NewDecimal(100))
		a2 := MustAmount(NewDecimal(50))

		// Amount + Amount
		sum := a1.Add(a2)
		if sum.String() != "150" {
			t.Errorf("100 + 50 should be 150, got %s", sum.String())
		}

		// Amount - Amount
		diff, err := a1.Sub(a2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if diff.String() != "50" {
			t.Errorf("100 - 50 should be 50, got %s", diff.String())
		}

		// Amount - Amount (negative result)
		_, err = a2.Sub(a1)
		if err != ErrNegativeAmount {
			t.Error("50 - 100 should return ErrNegativeAmount")
		}

		// Amount * Decimal
		doubled := a1.Mul(NewDecimal(2))
		if doubled.String() != "200" {
			t.Errorf("100 * 2 should be 200, got %s", doubled.String())
		}

		// Amount / Decimal
		halved, err := a1.Div(NewDecimal(2))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if halved.String() != "50" {
			t.Errorf("100 / 2 should be 50, got %s", halved.String())
		}

		_, err = a1.Div(Zero())
		if err != ErrDivisionByZero {
			t.Error("dividing amount by zero should return ErrDivisionByZero")
		}
	})

	t.Run("price operations", func(t *testing.T) {
		amount := MustAmount(NewDecimal(10))
		price := MustPrice(NewDecimal(100))

		// Amount * Price
		value := amount.MulPrice(price)
		if value.String() != "1000" {
			t.Errorf("10 * 100 should be 1000, got %s", value.String())
		}

		// Amount / Price
		result, err := value.DivPrice(price)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "10" {
			t.Errorf("1000 / 100 should be 10, got %s", result.String())
		}

		// Division by zero price
		_, err = amount.DivPrice(ZeroPrice())
		if err != ErrDivisionByZero {
			t.Error("dividing by zero price should return ErrDivisionByZero")
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		a1 := MustAmount(NewDecimal(100))
		a2 := MustAmount(NewDecimal(50))
		a3 := MustAmount(NewDecimal(100))

		if !a1.GreaterThan(a2) {
			t.Error("100 should be greater than 50")
		}

		if !a2.LessThan(a1) {
			t.Error("50 should be less than 100")
		}

		if !a1.Equal(a3) {
			t.Error("100 should equal 100")
		}
	})
}

// TestTypeSafety validates that invalid type operations cause compile errors
// Note: These are commented out to show what SHOULD NOT compile
func TestTypeSafety(t *testing.T) {
	// The following would cause compile errors (intentionally):
	// price := MustPrice(NewDecimal(100))
	// amount := MustAmount(NewDecimal(10))
	//
	// _ = price.Add(amount)      // COMPILE ERROR: cannot add Price and Amount
	// _ = amount.Add(price)      // COMPILE ERROR: cannot add Amount and Price
	// _ = price.Sub(amount)      // COMPILE ERROR: cannot subtract Amount from Price
	// _ = amount.Mul(price)      // COMPILE ERROR: must use MulPrice

	// This test passes by existing; the type safety is enforced at compile time
	t.Log("Type safety is enforced at compile time")
}

// TestTime tests Time type operations
func TestTime(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		now := Now()
		if now.Unix() == 0 {
			t.Error("Now() should return non-zero time")
		}

		epoch := Unix(0, 0)
		if epoch.Unix() != 0 {
			t.Error("Unix(0, 0) should return epoch time")
		}

		goTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		wrapped := NewTime(goTime)
		if !wrapped.Time().Equal(goTime) {
			t.Error("wrapped time should equal original")
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		t1 := Unix(1000, 0)
		d := Seconds(100)

		t2 := t1.Add(d)
		if t2.Unix() != 1100 {
			t.Errorf("1000 + 100 should be 1100, got %d", t2.Unix())
		}

		diff := t2.Sub(t1)
		if diff.Seconds() != 100 {
			t.Errorf("difference should be 100 seconds, got %f", diff.Seconds())
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		t1 := Unix(1000, 0)
		t2 := Unix(2000, 0)
		t3 := Unix(1000, 0)

		if !t1.Before(t2) {
			t.Error("t1 should be before t2")
		}

		if !t2.After(t1) {
			t.Error("t2 should be after t1")
		}

		if !t1.Equal(t3) {
			t.Error("t1 should equal t3")
		}
	})

	t.Run("formatting", func(t *testing.T) {
		goTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
		wrapped := NewTime(goTime)

		formatted := wrapped.Format("2006-01-02")
		if formatted != "2024-01-01" {
			t.Errorf("expected 2024-01-01, got %s", formatted)
		}
	})
}

// TestDuration tests Duration type operations
func TestDuration(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		d1 := Seconds(60)
		if d1.Seconds() != 60 {
			t.Errorf("expected 60 seconds, got %f", d1.Seconds())
		}

		d2 := Minutes(1)
		if d2.Minutes() != 1 {
			t.Errorf("expected 1 minute, got %f", d2.Minutes())
		}

		d3 := Hours(1)
		if d3.Hours() != 1 {
			t.Errorf("expected 1 hour, got %f", d3.Hours())
		}

		d4 := Days(1)
		if d4.Hours() != 24 {
			t.Errorf("expected 24 hours, got %f", d4.Hours())
		}

		goDur := time.Second * 30
		wrapped := NewDuration(goDur)
		if wrapped.Seconds() != 30 {
			t.Errorf("expected 30 seconds, got %f", wrapped.Seconds())
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		d1 := Seconds(60)
		d2 := Seconds(30)

		sum := d1.Add(d2)
		if sum.Seconds() != 90 {
			t.Errorf("60 + 30 should be 90, got %f", sum.Seconds())
		}

		diff := d1.Sub(d2)
		if diff.Seconds() != 30 {
			t.Errorf("60 - 30 should be 30, got %f", diff.Seconds())
		}

		mul := d1.Mul(2)
		if mul.Seconds() != 120 {
			t.Errorf("60 * 2 should be 120, got %f", mul.Seconds())
		}

		div, err := d1.Div(2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if div.Seconds() != 30 {
			t.Errorf("60 / 2 should be 30, got %f", div.Seconds())
		}

		_, err = d1.Div(0)
		if err != ErrDivisionByZero {
			t.Error("dividing duration by zero should return ErrDivisionByZero")
		}
	})

	t.Run("sign operations", func(t *testing.T) {
		negative := Seconds(-60)

		if !NewDuration(0).IsZero() {
			t.Error("zero duration should be zero")
		}

		abs := negative.Abs()
		if abs.Seconds() != 60 {
			t.Errorf("|-60s| should be 60s, got %f", abs.Seconds())
		}
	})

	t.Run("comparisons", func(t *testing.T) {
		d1 := Seconds(60)
		d2 := Seconds(30)
		d3 := Seconds(60)

		if !d1.GreaterThan(d2) {
			t.Error("60s should be greater than 30s")
		}

		if !d2.LessThan(d1) {
			t.Error("30s should be less than 60s")
		}

		if !d1.Equal(d3) {
			t.Error("60s should equal 60s")
		}
	})
}

// TestDecimalPrecision tests decimal precision for financial calculations
func TestDecimalPrecision(t *testing.T) {
	t.Run("no float precision loss", func(t *testing.T) {
		// This would fail with float64: 0.1 + 0.2 != 0.3
		a := MustDecimalFromString("0.1")
		b := MustDecimalFromString("0.2")
		c := MustDecimalFromString("0.3")

		sum := a.Add(b)
		if !sum.Equal(c) {
			t.Errorf("0.1 + 0.2 should equal 0.3, got %s", sum.String())
		}
	})

	t.Run("price calculations", func(t *testing.T) {
		price := MustPrice(MustDecimalFromString("1999.99"))
		amount := MustAmount(MustDecimalFromString("3.5"))

		total := amount.MulPrice(price)
		expected := MustDecimalFromString("6999.965")

		if !total.Decimal().Equal(expected) {
			t.Errorf("expected %s, got %s", expected.String(), total.String())
		}
	})
}

// TestMustFunctions tests panic behavior of Must* functions
func TestMustFunctions(t *testing.T) {
	t.Run("MustDecimalFromString panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDecimalFromString with invalid string should panic")
			}
		}()
		MustDecimalFromString("invalid")
	})

	t.Run("MustDecimalFromString success", func(t *testing.T) {
		d := MustDecimalFromString("123.45")
		if d.String() != "123.45" {
			t.Errorf("expected 123.45, got %s", d.String())
		}
	})
}
