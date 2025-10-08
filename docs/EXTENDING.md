# Extending the Framework: Adding New Mechanisms

This guide shows you how to extend go-crypto-quant-toolkit by adding new market mechanisms. The framework's interface-first design means you can add support for new protocols, venues, or trading mechanisms without modifying any framework code.

## Table of Contents
- [Overview](#overview)
- [The 4-Step Process](#the-4-step-process)
- [Step 1: Choose or Define Interface](#step-1-choose-or-define-interface)
- [Step 2: Implement the Mechanism](#step-2-implement-the-mechanism)
- [Step 3: Test Your Implementation](#step-3-test-your-implementation)
- [Step 4: Document Your Implementation](#step-4-document-your-implementation)
- [Complete Examples](#complete-examples)
- [Best Practices](#best-practices)
- [Interface Contract Testing](#interface-contract-testing)

## Overview

The framework provides three core mechanism interfaces in `pkg/mechanisms/`:
- **`LiquidityPool`** - For AMM-style liquidity pools (Uniswap, Curve, Balancer, etc.)
- **`Derivative`** - For derivative instruments (options, perpetuals, futures)
- **`OrderBook`** - For order book based trading (CEX-style limit order books)

Each interface extends the base `MarketMechanism` interface, which provides identification and context.

You can:
1. Implement one of the existing interfaces (most common)
2. Define a new interface for a novel mechanism category
3. Create position types that work with the framework's `Portfolio`

## The 4-Step Process

### Step 1: Choose or Define Interface

**If your mechanism fits an existing category**, implement that interface directly.

**Examples of mechanism → interface mapping:**
- Uniswap V2/V3, Curve, Balancer → `LiquidityPool`
- Black-Scholes options, Perpetual futures → `Derivative`
- Centralized exchange order books → `OrderBook`
- Constant-product AMM → `LiquidityPool` (see examples/custom_mechanism)

**If you're adding a novel mechanism type**, define a new interface in your code that embeds `MarketMechanism`:

```go
// FlashLoan represents flash loan providers
type FlashLoan interface {
    mechanisms.MarketMechanism
    
    // MaxBorrow returns maximum borrowable amount for a token
    MaxBorrow(ctx context.Context, token string) (primitives.Amount, error)
    
    // Fee returns the flash loan fee for a given amount
    Fee(ctx context.Context, amount primitives.Amount) (primitives.Amount, error)
    
    // Execute simulates a flash loan execution
    Execute(ctx context.Context, params FlashLoanParams) (FlashLoanResult, error)
}
```

**Key principle**: Interfaces should be minimal (3-5 methods), composable, and document their contracts clearly.

### Step 2: Implement the Mechanism

Create your implementation following these guidelines:

#### 2.1 Structure Your Code

Place implementations in a logical package structure:
```
pkg/implementations/
  your_mechanism/
    mechanism.go          # Core implementation
    mechanism_test.go     # Comprehensive tests
    README.md             # Usage documentation (optional but recommended)
```

#### 2.2 Implement Required Methods

Here's a template for implementing `LiquidityPool` (the most common case):

```go
package your_mechanism

import (
    "context"
    "fmt"
    
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// YourPool implements mechanisms.LiquidityPool for [protocol name].
//
// Describe what makes this pool unique:
// - Mathematical formula (e.g., constant product, stable swap)
// - Key parameters and their effects
// - Any assumptions or limitations
type YourPool struct {
    poolID string
    // Add your pool-specific fields
    tokenA string
    tokenB string
    feeRate primitives.Decimal
}

// NewYourPool creates a new pool instance.
func NewYourPool(poolID, tokenA, tokenB string, feeRate primitives.Decimal) *YourPool {
    return &YourPool{
        poolID: poolID,
        tokenA: tokenA,
        tokenB: tokenB,
        feeRate: feeRate,
    }
}

// Mechanism returns the mechanism type (implements mechanisms.MarketMechanism).
func (p *YourPool) Mechanism() mechanisms.MechanismType {
    return mechanisms.MechanismTypeLiquidityPool
}

// Venue returns the venue identifier.
func (p *YourPool) Venue() string {
    return "your-protocol-name"
}

// Calculate computes pool state using your protocol's formula.
//
// This is the core method that implements your pricing/state logic.
// Use primitives.Decimal for all financial calculations.
func (p *YourPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Validate inputs
    if params.ReserveA.IsZero() || params.ReserveB.IsZero() {
        return mechanisms.PoolState{}, fmt.Errorf("reserves cannot be zero")
    }
    
    // Implement your pricing formula using primitives.Decimal
    // Example: spot price = reserveB / reserveA
    priceDecimal, err := params.ReserveB.Decimal().Div(params.ReserveA.Decimal())
    if err != nil {
        return mechanisms.PoolState{}, fmt.Errorf("failed to calculate price: %w", err)
    }
    
    spotPrice := primitives.MustPrice(priceDecimal)
    
    // Calculate other pool metrics (liquidity, fees, etc.)
    // ...
    
    return mechanisms.PoolState{
        SpotPrice:          spotPrice,
        Liquidity:          params.ReserveA, // Simplified
        EffectiveLiquidity: params.ReserveA,
        AccumulatedFeesA:   primitives.ZeroAmount(),
        AccumulatedFeesB:   primitives.ZeroAmount(),
        Metadata:           make(map[string]interface{}),
    }, nil
}

// AddLiquidity simulates adding liquidity to the pool.
func (p *YourPool) AddLiquidity(ctx context.Context, amounts mechanisms.TokenAmounts) (mechanisms.PoolPosition, error) {
    // Implement your liquidity addition logic
    // Calculate LP tokens, optimal ratios, etc.
    
    return mechanisms.PoolPosition{
        AmountA:  amounts.AmountA,
        AmountB:  amounts.AmountB,
        Metadata: make(map[string]interface{}),
    }, nil
}

// RemoveLiquidity simulates removing liquidity from the pool.
func (p *YourPool) RemoveLiquidity(ctx context.Context, position mechanisms.PoolPosition) (mechanisms.TokenAmounts, error) {
    // Implement your liquidity removal logic
    // Calculate token amounts to return based on position
    
    return mechanisms.TokenAmounts{
        AmountA: position.AmountA,
        AmountB: position.AmountB,
    }, nil
}
```

#### 2.3 Use Decimal Arithmetic

**CRITICAL**: Always use `primitives.Decimal` for financial calculations. Never use `float64` for money or prices.

```go
// ✅ CORRECT: Use Decimal for all financial math
price, err := reserveB.Decimal().Div(reserveA.Decimal())
if err != nil {
    return err
}
spotPrice := primitives.MustPrice(price)

// ❌ WRONG: Never use float64
floatPrice := float64(reserveB) / float64(reserveA)  // Precision loss!
```

#### 2.4 Handle Errors Explicitly

Document all error conditions and validate inputs:

```go
// Document error conditions
var (
    ErrInvalidReserves = errors.New("reserves must be positive")
    ErrInsufficientLiquidity = errors.New("insufficient liquidity")
    ErrInvalidFee = errors.New("fee rate must be between 0 and 1")
)

// Validate all inputs
func (p *YourPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    if params.ReserveA.IsZero() || params.ReserveB.IsZero() {
        return mechanisms.PoolState{}, ErrInvalidReserves
    }
    
    // Check context cancellation if doing expensive operations
    select {
    case <-ctx.Done():
        return mechanisms.PoolState{}, ctx.Err()
    default:
    }
    
    // Your calculations...
}
```

#### 2.5 Leverage Existing Libraries

For battle-tested protocols, prefer wrapping existing implementations:

```go
// ✅ GOOD: Wrap existing SDK for accuracy
import "github.com/protocol/official-sdk"

type YourPool struct {
    sdk *official_sdk.Pool
}

func (p *YourPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Adapt SDK types to framework types
    sdkResult := p.sdk.GetPrice(params.ReserveA, params.ReserveB)
    
    return mechanisms.PoolState{
        SpotPrice: primitives.MustPrice(sdkResult.Price),
        // Map other fields...
    }, nil
}
```

See `pkg/implementations/concentrated_liquidity/pool.go` for a real example wrapping `daoleno/uniswapv3-sdk`.

### Step 3: Test Your Implementation

#### 3.1 Write Comprehensive Unit Tests

Test files should be thorough and cover edge cases:

```go
package your_mechanism

import (
    "context"
    "testing"
    
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

func TestYourPool_Calculate(t *testing.T) {
    tests := []struct {
        name        string
        params      mechanisms.PoolParams
        wantPrice   string // Use string for exact decimal comparison
        wantErr     error
    }{
        {
            name: "basic calculation",
            params: mechanisms.PoolParams{
                ReserveA: primitives.MustAmount(primitives.MustDecimal("1000")),
                ReserveB: primitives.MustAmount(primitives.MustDecimal("2000")),
            },
            wantPrice: "2.0",  // reserveB / reserveA
            wantErr:   nil,
        },
        {
            name: "zero reserve A",
            params: mechanisms.PoolParams{
                ReserveA: primitives.ZeroAmount(),
                ReserveB: primitives.MustAmount(primitives.MustDecimal("1000")),
            },
            wantErr: ErrInvalidReserves,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pool := NewYourPool("test-pool", "TOKEN-A", "TOKEN-B", primitives.MustDecimal("0.003"))
            
            got, err := pool.Calculate(context.Background(), tt.params)
            
            if tt.wantErr != nil {
                if err == nil || err != tt.wantErr {
                    t.Errorf("Calculate() error = %v, wantErr %v", err, tt.wantErr)
                }
                return
            }
            
            if err != nil {
                t.Fatalf("Calculate() unexpected error = %v", err)
            }
            
            if tt.wantPrice != "" {
                gotPrice := got.SpotPrice.Decimal().String()
                if gotPrice != tt.wantPrice {
                    t.Errorf("Calculate() price = %v, want %v", gotPrice, tt.wantPrice)
                }
            }
        })
    }
}

// Test edge cases
func TestYourPool_EdgeCases(t *testing.T) {
    pool := NewYourPool("test", "A", "B", primitives.MustDecimal("0.003"))
    
    t.Run("very large reserves", func(t *testing.T) {
        params := mechanisms.PoolParams{
            ReserveA: primitives.MustAmount(primitives.MustDecimal("1e18")),
            ReserveB: primitives.MustAmount(primitives.MustDecimal("1e18")),
        }
        _, err := pool.Calculate(context.Background(), params)
        if err != nil {
            t.Errorf("should handle large values: %v", err)
        }
    })
    
    t.Run("very small reserves", func(t *testing.T) {
        params := mechanisms.PoolParams{
            ReserveA: primitives.MustAmount(primitives.MustDecimal("0.000001")),
            ReserveB: primitives.MustAmount(primitives.MustDecimal("0.000001")),
        }
        _, err := pool.Calculate(context.Background(), params)
        if err != nil {
            t.Errorf("should handle small values: %v", err)
        }
    })
    
    t.Run("context cancellation", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        cancel() // Cancel immediately
        
        params := mechanisms.PoolParams{
            ReserveA: primitives.MustAmount(primitives.MustDecimal("1000")),
            ReserveB: primitives.MustAmount(primitives.MustDecimal("2000")),
        }
        
        _, err := pool.Calculate(ctx, params)
        // Should respect context cancellation if implementation is long-running
        _ = err
    })
}
```

#### 3.2 Validate Against Reference Values

If implementing a known protocol, validate against published values:

```go
func TestYourPool_AgainstReferenceValues(t *testing.T) {
    // Use known values from protocol documentation or on-chain data
    pool := NewYourPool("test", "USDC", "ETH", primitives.MustDecimal("0.003"))
    
    // Example: Known state from block X
    params := mechanisms.PoolParams{
        ReserveA: primitives.MustAmount(primitives.MustDecimal("1000000")),  // 1M USDC
        ReserveB: primitives.MustAmount(primitives.MustDecimal("500")),      // 500 ETH
    }
    
    result, err := pool.Calculate(context.Background(), params)
    if err != nil {
        t.Fatalf("Calculate failed: %v", err)
    }
    
    // Expected price: 1000000 / 500 = 2000 USDC per ETH
    expectedPrice := primitives.MustDecimal("2000")
    
    // Allow small tolerance for rounding (e.g., 0.01%)
    diff := result.SpotPrice.Decimal().Sub(expectedPrice).Abs()
    tolerance := expectedPrice.Mul(primitives.MustDecimal("0.0001")) // 0.01%
    
    if diff.GreaterThan(tolerance) {
        t.Errorf("Price outside tolerance: got %v, want %v, diff %v",
            result.SpotPrice.Decimal(), expectedPrice, diff)
    }
}
```

#### 3.3 Achieve >80% Test Coverage

Run coverage analysis:

```bash
go test ./pkg/implementations/your_mechanism/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Focus on covering:
- All public methods
- Error paths
- Edge cases (zero values, very large/small numbers)
- Context cancellation (if applicable)

### Step 4: Document Your Implementation

#### 4.1 Add Package Documentation

At the top of your main file:

```go
// Package your_mechanism implements [protocol name] for go-crypto-quant-toolkit.
//
// [Protocol Name] is a [brief description of what it does and why it's useful].
//
// Key Features:
//   - Feature 1: Description
//   - Feature 2: Description
//   - Feature 3: Description
//
// Mathematical Model:
//   [Brief description of the core formula or algorithm]
//   Example: Constant product formula x * y = k
//
// Usage Example:
//
//	pool := your_mechanism.NewYourPool("pool-1", "USDC", "ETH", primitives.MustDecimal("0.003"))
//	
//	params := mechanisms.PoolParams{
//	    ReserveA: primitives.MustAmount(primitives.MustDecimal("1000000")),
//	    ReserveB: primitives.MustAmount(primitives.MustDecimal("500")),
//	}
//	
//	state, err := pool.Calculate(context.Background(), params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	
//	fmt.Printf("Spot price: %s\n", state.SpotPrice.Decimal())
//
// References:
//   - Protocol documentation: https://...
//   - Whitepaper: https://...
//   - Reference implementation: https://...
package your_mechanism
```

#### 4.2 Document All Public Types and Methods

Use godoc format for all exports:

```go
// YourPool implements mechanisms.LiquidityPool for [protocol].
//
// The pool uses [mathematical model] to determine prices and simulate
// liquidity operations.
//
// Thread Safety: YourPool methods are safe for concurrent reads but require
// external synchronization for concurrent modifications to pool state.
//
// Precision: All calculations use github.com/shopspring/decimal with 32 decimal
// places of precision. Rounding is performed using ROUND_HALF_UP strategy.
type YourPool struct {
    // poolID uniquely identifies this pool instance
    poolID string
    
    // Document non-obvious fields
    // ...
}

// Calculate computes the pool state based on current reserves and parameters.
//
// The calculation follows the [formula name] formula:
//   [mathematical formula or description]
//
// Parameters:
//   - ctx: Context for cancellation and deadlines
//   - params: Pool parameters including reserves and optional metadata
//
// Returns:
//   - PoolState: Current pool state including spot price and liquidity
//   - error: Returns ErrInvalidReserves if reserves are zero or negative
//
// Example:
//
//	state, err := pool.Calculate(ctx, mechanisms.PoolParams{
//	    ReserveA: primitives.MustAmount(primitives.MustDecimal("1000")),
//	    ReserveB: primitives.MustAmount(primitives.MustDecimal("2000")),
//	})
func (p *YourPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Implementation...
}
```

#### 4.3 Create a README (Optional but Recommended)

Add `pkg/implementations/your_mechanism/README.md`:

```markdown
# Your Mechanism Implementation

Implementation of [Protocol Name] for go-crypto-quant-toolkit.

## Overview

Brief description of what this mechanism does and why it's useful.

## Mathematical Model

Detailed explanation of the core algorithm/formula.

## Usage

```go
// Example code showing how to use your implementation
```

## Testing

```bash
go test ./pkg/implementations/your_mechanism/...
```

## Validation

Describe how you validated the implementation:
- Reference values used
- Tolerance levels
- Known limitations

## References

- [Protocol Documentation](https://...)
- [Whitepaper](https://...)
```

## Complete Examples

### Example 1: Custom Constant-Product AMM

See `examples/custom_mechanism/main.go` for a complete working example (~365 lines) that:
1. Implements `LiquidityPool` interface
2. Creates a custom position type
3. Uses it in a strategy
4. Runs a backtest

Key snippets from that example:

```go
// Custom pool implementation
type ConstantProductPool struct {
    poolID       string
    tokenASymbol string
    tokenBSymbol string
    feeRate      primitives.Decimal
}

// Implements mechanisms.LiquidityPool
func (p *ConstantProductPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Constant product: x * y = k
    // Price = reserveB / reserveA
    priceDecimal, err := params.ReserveB.Decimal().Div(params.ReserveA.Decimal())
    if err != nil {
        return mechanisms.PoolState{}, fmt.Errorf("failed to calculate price: %w", err)
    }
    
    spotPrice := primitives.MustPrice(priceDecimal)
    
    return mechanisms.PoolState{
        SpotPrice:          spotPrice,
        Liquidity:          params.ReserveA,
        EffectiveLiquidity: params.ReserveA,
        // ... other fields
    }, nil
}
```

Run it:
```bash
go run examples/custom_mechanism/main.go
```

### Example 2: Uniswap V3 Concentrated Liquidity

See `pkg/implementations/concentrated_liquidity/pool.go` for a production-grade example that:
1. Wraps the `daoleno/uniswapv3-sdk` library
2. Adapts SDK types to framework types
3. Validates against Uniswap V3 reference values
4. Includes comprehensive tests with >78% coverage

Key pattern:

```go
// Wrap existing SDK for battle-tested math
type Pool struct {
    poolID      string
    tokenA      *core.Token
    tokenB      *core.Token
    fee         constants.FeeAmount
    tickSpacing int
}

// Adapt SDK calls to framework interface
func (p *Pool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Extract metadata
    currentTick := params.Metadata["current_tick"].(int)
    sqrtPriceX96 := params.Metadata["sqrt_price_x96"].(*big.Int)
    
    // Use SDK for calculations
    sqrtRatioX96, _ := utils.GetSqrtRatioAtTick(currentTick)
    
    // Convert to framework types
    return mechanisms.PoolState{
        SpotPrice: primitives.MustPrice(calculatedPrice),
        // ... other fields
    }, nil
}
```

### Example 3: Black-Scholes Options

See `pkg/implementations/blackscholes/option.go` for implementing the `Derivative` interface:

```go
type Option struct {
    optionID   string
    underlying string
    strike     primitives.Price
    expiry     time.Time
    optionType OptionType
    style      Style
}

func (o *Option) Price(ctx context.Context, params mechanisms.PriceParams) (primitives.Price, error) {
    // Black-Scholes formula implementation using Decimal arithmetic
    d1 := // ... calculation
    d2 := // ... calculation
    
    var price primitives.Decimal
    if o.optionType == Call {
        price = // ... call option formula
    } else {
        price = // ... put option formula
    }
    
    return primitives.MustPrice(price), nil
}
```

## Best Practices

### Financial Calculations

1. **Always use `primitives.Decimal`** - Never use `float64` for money
2. **Validate inputs** - Check for zero, negative, or invalid values
3. **Handle division by zero** - Decimal.Div() returns an error
4. **Document precision** - State rounding behavior and precision guarantees
5. **Test with extreme values** - Very large and very small numbers

```go
// ✅ CORRECT
price, err := reserveB.Decimal().Div(reserveA.Decimal())
if err != nil {
    return err  // Handles division by zero
}

// ❌ WRONG
price := float64(reserveB) / float64(reserveA)  // Precision loss + panic on zero
```

### Error Handling

1. **Define custom errors** - Use `var Err... = errors.New(...)`
2. **Wrap errors** - Use `fmt.Errorf("context: %w", err)` for stack traces
3. **Document error conditions** - List all possible errors in godoc
4. **Validate early** - Check inputs before expensive operations

```go
var (
    ErrInvalidInput = errors.New("invalid input")
    ErrOutOfBounds = errors.New("value out of bounds")
)

func (p *Pool) Calculate(ctx context.Context, params PoolParams) (PoolState, error) {
    // Validate early
    if params.ReserveA.IsZero() {
        return PoolState{}, ErrInvalidInput
    }
    
    // Wrap errors
    result, err := p.complexCalculation(params)
    if err != nil {
        return PoolState{}, fmt.Errorf("calculation failed: %w", err)
    }
    
    return result, nil
}
```

### Context Usage

1. **Accept context** - All I/O or long-running operations should accept `context.Context`
2. **Check cancellation** - Use `select` on `ctx.Done()` for expensive operations
3. **Propagate context** - Pass context to downstream calls

```go
func (p *Pool) Calculate(ctx context.Context, params PoolParams) (PoolState, error) {
    // Check context before expensive operation
    select {
    case <-ctx.Done():
        return PoolState{}, ctx.Err()
    default:
    }
    
    // Expensive calculation...
    
    return state, nil
}
```

### Interface Compliance

1. **Implement all methods** - Even if some are not applicable, return appropriate errors
2. **Match signatures exactly** - Use the exact types from the interface
3. **Follow conventions** - Return types, error handling, method names
4. **Document behavior** - Clarify what each method does for your mechanism

### Testing Best Practices

1. **Table-driven tests** - Use test tables for multiple cases
2. **Test edge cases** - Zero, negative, very large, very small values
3. **Validate against references** - Use known values from protocol
4. **Test error paths** - Ensure errors are returned correctly
5. **Achieve >80% coverage** - Focus on core logic and error handling

## Interface Contract Testing

The framework provides patterns for testing interface contracts. While implementations should test their specific logic, you can also validate that your implementation satisfies the interface contract.

### Basic Contract Validation

```go
func TestYourPool_ImplementsInterface(t *testing.T) {
    // Compile-time check that your type implements the interface
    var _ mechanisms.LiquidityPool = (*YourPool)(nil)
}

func TestYourPool_MechanismContract(t *testing.T) {
    pool := NewYourPool("test", "A", "B", primitives.MustDecimal("0.003"))
    
    // Verify Mechanism() returns correct type
    if pool.Mechanism() != mechanisms.MechanismTypeLiquidityPool {
        t.Errorf("Mechanism() = %v, want %v", 
            pool.Mechanism(), mechanisms.MechanismTypeLiquidityPool)
    }
    
    // Verify Venue() returns non-empty string
    if pool.Venue() == "" {
        t.Error("Venue() returned empty string")
    }
}
```

### Property-Based Contract Tests

For more rigorous testing, implement property-based tests:

```go
func TestYourPool_Properties(t *testing.T) {
    pool := NewYourPool("test", "A", "B", primitives.MustDecimal("0.003"))
    
    // Property: Price should be inversely proportional to reserve ratio
    t.Run("price proportional to reserves", func(t *testing.T) {
        params1 := mechanisms.PoolParams{
            ReserveA: primitives.MustAmount(primitives.MustDecimal("1000")),
            ReserveB: primitives.MustAmount(primitives.MustDecimal("2000")),
        }
        
        params2 := mechanisms.PoolParams{
            ReserveA: primitives.MustAmount(primitives.MustDecimal("1000")),
            ReserveB: primitives.MustAmount(primitives.MustDecimal("4000")),
        }
        
        state1, err := pool.Calculate(context.Background(), params1)
        if err != nil {
            t.Fatal(err)
        }
        
        state2, err := pool.Calculate(context.Background(), params2)
        if err != nil {
            t.Fatal(err)
        }
        
        // Price should double when ReserveB doubles (for constant product)
        price1 := state1.SpotPrice.Decimal()
        price2 := state2.SpotPrice.Decimal()
        expectedRatio := primitives.MustDecimal("2")
        
        actualRatio, _ := price2.Div(price1)
        if !actualRatio.Equal(expectedRatio) {
            t.Errorf("Price ratio = %v, want %v", actualRatio, expectedRatio)
        }
    })
    
    // Property: Adding and removing liquidity should be reversible
    t.Run("add then remove liquidity", func(t *testing.T) {
        amounts := mechanisms.TokenAmounts{
            AmountA: primitives.MustAmount(primitives.MustDecimal("1000")),
            AmountB: primitives.MustAmount(primitives.MustDecimal("2000")),
        }
        
        position, err := pool.AddLiquidity(context.Background(), amounts)
        if err != nil {
            t.Fatal(err)
        }
        
        removed, err := pool.RemoveLiquidity(context.Background(), position)
        if err != nil {
            t.Fatal(err)
        }
        
        // Should get back approximately the same amounts (minus fees)
        if !removed.AmountA.Decimal().Equal(amounts.AmountA.Decimal()) {
            t.Errorf("AmountA = %v, want %v", removed.AmountA, amounts.AmountA)
        }
    })
}
```

## Next Steps

1. **Study existing implementations** - Look at `pkg/implementations/` for patterns
2. **Run examples** - Execute `examples/custom_mechanism/main.go` to see it in action
3. **Read ARCHITECTURE.md** - Understand the design philosophy
4. **Join the community** - Share your implementation or ask questions

## Common Pitfalls to Avoid

❌ **Using float64 for financial calculations**
```go
// WRONG
price := float64(reserveB) / float64(reserveA)
```

✅ **Use primitives.Decimal instead**
```go
// CORRECT
price, err := reserveB.Decimal().Div(reserveA.Decimal())
```

---

❌ **Ignoring errors from Decimal operations**
```go
// WRONG
result, _ := a.Div(b)  // Division by zero panics!
```

✅ **Always check errors**
```go
// CORRECT
result, err := a.Div(b)
if err != nil {
    return err
}
```

---

❌ **Not validating inputs**
```go
// WRONG
func Calculate(params PoolParams) (PoolState, error) {
    // Assume params are valid...
}
```

✅ **Validate early**
```go
// CORRECT
func Calculate(params PoolParams) (PoolState, error) {
    if params.ReserveA.IsZero() {
        return PoolState{}, ErrInvalidReserves
    }
    // ...
}
```

---

❌ **Modifying framework code**
```go
// WRONG
// Editing pkg/mechanisms/liquidity_pool.go to add your method
```

✅ **Implement interfaces in your own package**
```go
// CORRECT
// Create pkg/implementations/your_mechanism/pool.go
type YourPool struct { ... }
func (p *YourPool) Calculate(...) { ... }
```

---

❌ **Insufficient test coverage**
```go
// WRONG
func TestBasicCase(t *testing.T) {
    // Only tests one happy path
}
```

✅ **Test edge cases and errors**
```go
// CORRECT
func TestCalculate(t *testing.T) {
    tests := []struct{
        name string
        // Test: happy path, errors, edge cases, boundaries
    }{...}
}
```

## Summary

Adding a new mechanism to go-crypto-quant-toolkit requires:

1. ✅ **Choose/define interface** - Pick existing or create new
2. ✅ **Implement methods** - Use Decimal arithmetic, handle errors
3. ✅ **Write tests** - >80% coverage, validate against references
4. ✅ **Document** - Package docs, method docs, README

The framework's interface-first design ensures you can add any mechanism type without modifying framework code. Follow the patterns in existing implementations and examples, and your mechanism will integrate seamlessly.

For more details on the architecture and design philosophy, see [ARCHITECTURE.md](ARCHITECTURE.md).
