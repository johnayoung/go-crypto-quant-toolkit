# Architecture & Design Philosophy

This document explains the design decisions behind go-crypto-quant-toolkit and how its architecture enables extensibility without framework modifications.

## Table of Contents
- [Core Philosophy](#core-philosophy)
- [Design Principles](#design-principles)
- [Architecture Overview](#architecture-overview)
- [Component Relationships](#component-relationships)
- [Extension Points](#extension-points)
- [Why Interface-First Design](#why-interface-first-design)
- [Trade-offs and Design Decisions](#trade-offs-and-design-decisions)

## Core Philosophy

**"Define contracts, not implementations."**

go-crypto-quant-toolkit is built on the principle that a quantitative trading framework should provide clear interfaces and composition patterns, not prescribe specific market mechanisms or strategies. Users should be able to:

1. Add new market mechanisms without modifying framework code
2. Compose strategies from any combination of mechanisms
3. Extend the framework in ways the original authors didn't anticipate

This philosophy is embodied in three key architectural decisions:

1. **Interface-first design** - All core abstractions are interfaces
2. **Dependency inversion** - Framework depends on abstractions, not concrete types
3. **Composition over inheritance** - Strategies compose mechanisms, not extend classes

## Design Principles

### 1. Minimal Interfaces

Interfaces are small (3-5 methods) and focused on a single responsibility:

```go
// ❌ BAD: Monolithic interface
type MarketMechanism interface {
    Price() Price
    AddLiquidity() Position
    RemoveLiquidity() TokenAmounts
    Greeks() Greeks
    Settle() Amount
    PlaceOrder() OrderID
    CancelOrder() error
    // ... 20+ methods
}

// ✅ GOOD: Multiple focused interfaces
type LiquidityPool interface {
    MarketMechanism  // Base identification
    Calculate(PoolParams) (PoolState, error)
    AddLiquidity(TokenAmounts) (PoolPosition, error)
    RemoveLiquidity(PoolPosition) (TokenAmounts, error)
}

type Derivative interface {
    MarketMechanism  // Base identification
    Price(PriceParams) (Price, error)
    Greeks(PriceParams) (Greeks, error)
    Settle(PriceParams) (Amount, error)
}
```

**Why:** Small interfaces are easier to implement, test, and compose. Users implement only what they need.

### 2. Accept Interfaces, Return Structs

The framework follows Go's proverb: "Accept interfaces, return concrete types."

```go
// ✅ Framework accepts interfaces
type Portfolio struct {
    positions []Position  // Position is an interface
}

func (p *Portfolio) AddPosition(pos Position) error {
    // Works with ANY type implementing Position
}

// ✅ Framework returns concrete types
type PoolState struct {
    SpotPrice          Price
    Liquidity          Amount
    EffectiveLiquidity Amount
    // ... concrete fields
}

func (pool *Pool) Calculate(params PoolParams) (PoolState, error) {
    // Returns concrete struct, not interface
}
```

**Why:** This maximizes flexibility for callers while maintaining type safety. Users can pass any implementation, and framework code is clear about what it returns.

### 3. Zero Framework Dependencies

Implementations depend on interfaces, not concrete framework types:

```go
// ✅ GOOD: Implementation only imports interfaces
package mypool

import (
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"  // Interfaces only
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"  // Shared types only
)

type MyPool struct { ... }

func (p *MyPool) Calculate(params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // No dependency on other implementations
}
```

**Why:** This prevents coupling between implementations. New mechanisms can be added without importing or understanding other mechanisms.

### 4. Type Safety for Financial Math

All financial calculations use `primitives.Decimal` (wrapping `shopspring/decimal`):

```go
// Financial primitive types
type Price struct { decimal.Decimal }     // Represents price per unit
type Amount struct { decimal.Decimal }    // Represents quantity
type Decimal struct { decimal.Decimal }   // Generic precise decimal

// ❌ Compile error: Cannot add Price + Amount
price := primitives.MustPrice(d1)
amount := primitives.MustAmount(d2)
result := price + amount  // Type error!

// ✅ Explicit conversions required
priceValue := price.Decimal()
amountValue := amount.Decimal()
result := priceValue.Mul(amountValue)  // OK: Decimal × Decimal
```

**Why:** Prevents financial calculation errors at compile time. Forces developers to think about units and precision.

## Architecture Overview

The framework is organized in layers, each depending only on layers below:

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER CODE                                 │
│  • Custom strategies (implement Strategy interface)             │
│  • Custom mechanisms (implement Mechanism interfaces)           │
│  • Market data providers                                        │
│  • Position types                                               │
└───────────────────────┬─────────────────────────────────────────┘
                        │ uses
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                   STRATEGY FRAMEWORK                             │
│                   pkg/strategy/                                  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Strategy Interface                                       │  │
│  │   Rebalance(ctx, portfolio, market) → Actions           │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Portfolio (Concrete)                                     │  │
│  │   Manages Position interface instances                  │  │
│  │   Provides value queries, cash management               │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Position Interface                                       │  │
│  │   Value(market) → Amount                                │  │
│  │   Risk(market) → RiskMetrics                            │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ Action Interface                                         │  │
│  │   Apply(portfolio) → error                              │  │
│  │   Types: AddPosition, RemovePosition, AdjustCash, etc.  │  │
│  └──────────────────────────────────────────────────────────┘  │
└───────────────────────┬─────────────────────────────────────────┘
                        │ uses
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                 MECHANISM INTERFACES                             │
│                 pkg/mechanisms/                                  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ MarketMechanism (Base Interface)                         │  │
│  │   Mechanism() → MechanismType                           │  │
│  │   Venue() → string                                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────────┐  ┌─────────────┐  ┌──────────────────┐    │
│  │ LiquidityPool  │  │ Derivative  │  │   OrderBook      │    │
│  │  • Calculate   │  │  • Price    │  │  • BestBid/Ask  │    │
│  │  • AddLiq      │  │  • Greeks   │  │  • PlaceOrder   │    │
│  │  • RemoveLiq   │  │  • Settle   │  │  • CancelOrder  │    │
│  └────────────────┘  └─────────────┘  └──────────────────┘    │
│                                                                  │
│  Extension: Users can define new interfaces (FlashLoan, etc.)  │
└───────────────────────┬─────────────────────────────────────────┘
                        │ implemented by
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│              REFERENCE IMPLEMENTATIONS                           │
│              pkg/implementations/                                │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ concentrated_liquidity/                                  │  │
│  │   Pool (implements LiquidityPool)                       │  │
│  │   Wraps daoleno/uniswapv3-sdk for accuracy             │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ blackscholes/                                            │  │
│  │   Option (implements Derivative)                        │  │
│  │   Black-Scholes pricing and Greeks                      │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │ perpetual/                                               │  │
│  │   Future (implements Derivative)                        │  │
│  │   Perpetual futures with funding rates                  │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  Extension: Users add implementations in their own packages     │
└───────────────────────┬─────────────────────────────────────────┘
                        │ uses
                        ▼
┌─────────────────────────────────────────────────────────────────┐
│                      PRIMITIVES                                  │
│                      pkg/primitives/                             │
│                                                                  │
│  Type-safe financial primitives:                                │
│  • Price, Amount, Decimal (wrapping shopspring/decimal)         │
│  • Time, Duration                                               │
│  • Arithmetic operations with compile-time type checking        │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     BACKTEST ENGINE                              │
│                     pkg/backtest/                                │
│                                                                  │
│  Event-driven backtesting:                                      │
│  • Data event → Strategy.Rebalance() → Apply Actions           │
│  • Performance metrics (returns, Sharpe, drawdown)             │
│  • Context cancellation support                                │
│                                                                  │
│  Works with ANY Strategy and Position implementations          │
└─────────────────────────────────────────────────────────────────┘
```

### Key Architectural Properties

1. **Acyclic Dependencies** - No circular imports. Dependencies flow downward.
2. **Interface Boundaries** - Each layer depends on interfaces, not concrete types.
3. **Extensibility Points** - Users can add code at any layer without modifying framework.
4. **Testability** - Each component can be tested in isolation with mocks.

## Component Relationships

### Strategy ↔ Portfolio ↔ Position

```go
// Strategy decides what to do
type Strategy interface {
    Rebalance(ctx context.Context, p *Portfolio, m MarketSnapshot) ([]Action, error)
}

// Portfolio tracks what we have
type Portfolio struct {
    positions []Position  // Position is an interface
    cash      Decimal
}

// Position represents any tradeable asset
type Position interface {
    Value(m MarketSnapshot) (Amount, error)
    Risk(m MarketSnapshot) (RiskMetrics, error)
    Type() PositionType
}
```

**Relationship:**
1. Backtest engine calls `Strategy.Rebalance(portfolio, market)`
2. Strategy examines portfolio positions and market data
3. Strategy returns `Action` objects (AddPosition, RemovePosition, etc.)
4. Engine applies actions to portfolio
5. Portfolio tracks positions using `Position` interface

**Why this works:**
- Strategy never knows concrete position types (LPPosition, OptionPosition, etc.)
- Portfolio stores any type implementing Position
- New position types integrate without framework changes

### Mechanism Interfaces ↔ Implementations

```go
// Framework defines interface
type LiquidityPool interface {
    MarketMechanism
    Calculate(params PoolParams) (PoolState, error)
    AddLiquidity(amounts TokenAmounts) (PoolPosition, error)
    RemoveLiquidity(position PoolPosition) (TokenAmounts, error)
}

// User implements in their own package
package myamm

import "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"

type MyPool struct { ... }

func (p *MyPool) Calculate(params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Custom implementation
}

// ... implement other methods
```

**Relationship:**
1. Framework defines what a LiquidityPool can do (interface)
2. Users implement the interface with their protocol logic
3. Strategies use the interface without knowing which implementation
4. Multiple implementations can coexist (Uniswap V2, V3, Curve, etc.)

**Why this works:**
- Interface is stable (doesn't change once defined)
- Implementations are isolated (don't affect each other)
- Users can wrap existing SDKs (see concentrated_liquidity/pool.go)

## Extension Points

The framework provides multiple extension points where users can add functionality:

### 1. Custom Mechanism Implementations

**What:** Implement existing mechanism interfaces with new protocols.

**Where:** `pkg/implementations/` or your own package

**Example:**
```go
// Implement mechanisms.LiquidityPool for Curve stable pools
type CurveStablePool struct { ... }
func (p *CurveStablePool) Calculate(...) { ... }
```

**See:** `examples/custom_mechanism/main.go` for complete example

### 2. New Mechanism Types

**What:** Define new interface types for novel mechanisms.

**Where:** Your own package (define interface, implement it)

**Example:**
```go
// Define new mechanism type for flash loans
type FlashLoan interface {
    mechanisms.MarketMechanism
    MaxBorrow(ctx context.Context, token string) (primitives.Amount, error)
    Fee(ctx context.Context, amount primitives.Amount) (primitives.Amount, error)
    Execute(ctx context.Context, params FlashLoanParams) (FlashLoanResult, error)
}

// Implement it
type AaveFlashLoan struct { ... }
func (f *AaveFlashLoan) MaxBorrow(...) { ... }
```

**Key:** Your interface embeds `mechanisms.MarketMechanism` for consistency.

### 3. Custom Position Types

**What:** Create position wrappers for your mechanisms.

**Where:** Your mechanism package

**Example:**
```go
// Position wrapper for your custom mechanism
type MyPosition struct {
    pool      *MyPool
    liquidity primitives.Amount
    // ... other fields
}

func (p *MyPosition) Value(m strategy.MarketSnapshot) (primitives.Amount, error) {
    // Calculate value using market data
}

func (p *MyPosition) Risk(m strategy.MarketSnapshot) (strategy.RiskMetrics, error) {
    // Calculate risk metrics
}

func (p *MyPosition) Type() strategy.PositionType {
    return strategy.PositionTypeLiquidityPool
}
```

**Integration:** Portfolio accepts any type implementing `Position` interface.

### 4. Custom Strategies

**What:** Implement trading logic coordinating multiple mechanisms.

**Where:** Your strategy package

**Example:**
```go
type MyStrategy struct {
    lpPool      mechanisms.LiquidityPool
    derivative  mechanisms.Derivative
    targetDelta primitives.Decimal
}

func (s *MyStrategy) Rebalance(
    ctx context.Context,
    p *strategy.Portfolio,
    m strategy.MarketSnapshot,
) ([]strategy.Action, error) {
    // Your custom logic
    // Can use ANY mechanism implementations
}
```

**See:** `examples/delta_neutral/main.go` for multi-mechanism strategy

### 5. Market Data Providers

**What:** Provide market data in the `MarketSnapshot` format.

**Where:** Your data package

**Example:**
```go
func FetchMarketData(timestamp time.Time) (strategy.MarketSnapshot, error) {
    // Fetch from exchange APIs, databases, etc.
    return strategy.MarketSnapshot{
        Timestamp: timestamp,
        Prices:    prices,
        Metadata:  metadata,
    }, nil
}
```

**Integration:** Backtest engine accepts any function returning `MarketSnapshot`.

### 6. Performance Analytics

**What:** Add custom performance metrics beyond built-in ones.

**Where:** Your analytics package

**Example:**
```go
func CalculateSortino(result *backtest.Result) primitives.Decimal {
    // Calculate Sortino ratio from backtest results
}

func CalculateMaxConsecutiveLosses(result *backtest.Result) int {
    // Custom metric
}
```

**Integration:** Backtest results are public structs you can analyze.

## Why Interface-First Design

### Problem: Framework Lock-in

Traditional quantitative frameworks often suffer from tight coupling:

```go
// ❌ BAD: Concrete type dependencies
type Strategy struct {
    uniswapPool *UniswapV3Pool  // Locked to specific implementation
    gmxPerp     *GMXPerpetual
}

func (s *Strategy) Execute() {
    // Can only use these specific types
    // Adding new protocol requires modifying Strategy
}
```

**Issues:**
- Adding new protocols requires framework modifications
- Testing requires real protocol implementations
- Can't compose mechanisms from different sources
- Framework becomes bloated with every protocol addition

### Solution: Interface Boundaries

```go
// ✅ GOOD: Interface dependencies
type Strategy struct {
    pool mechanisms.LiquidityPool  // Works with ANY LiquidityPool
    perp mechanisms.Derivative     // Works with ANY Derivative
}

func (s *Strategy) Execute() {
    // Works with any implementation
    // No framework changes needed for new protocols
}
```

**Benefits:**
- Users add protocols without touching framework code
- Easy to test with mock implementations
- Strategies compose any mechanism types
- Framework stays minimal and focused

### Real-World Example: Adding Curve Pools

With interface-first design, adding Curve stable pools requires:

1. ✅ Create `pkg/implementations/curve/pool.go`
2. ✅ Implement `mechanisms.LiquidityPool` interface
3. ✅ Add tests
4. ✅ Use in strategies

**No framework modifications needed.**

With concrete-type design, you'd need to:

1. ❌ Modify `pkg/mechanisms/` to add Curve-specific methods
2. ❌ Update strategy framework to support Curve
3. ❌ Modify backtest engine to recognize Curve
4. ❌ Update all existing code that depends on mechanisms

**Framework modifications required everywhere.**

### Interface Stability

Once an interface is defined, it should never change (only add new interfaces):

```go
// ✅ Version 1.0
type LiquidityPool interface {
    Calculate(params PoolParams) (PoolState, error)
    AddLiquidity(amounts TokenAmounts) (PoolPosition, error)
    RemoveLiquidity(position PoolPosition) (TokenAmounts, error)
}

// ✅ Version 2.0 - Add new interface, don't modify existing
type LiquidityPoolV2 interface {
    LiquidityPool  // Embed existing interface
    EstimateImpermanentLoss(params PoolParams) (Decimal, error)  // New method
}

// Existing implementations still work
// New implementations can optionally implement V2
```

**Why:** Interface stability is critical for extensibility. Breaking changes force all implementations to update.

## Trade-offs and Design Decisions

### Trade-off 1: Flexibility vs. Convenience

**Decision:** Prioritize flexibility over convenience helpers.

```go
// Framework provides: Interface + minimal helpers
type LiquidityPool interface {
    Calculate(params PoolParams) (PoolState, error)
    // ... other methods
}

// Framework does NOT provide:
// - Pre-built strategies for every protocol
// - Automatic parameter optimization
// - Built-in data fetching
```

**Rationale:**
- Convenience features become opinionated and limit flexibility
- Users have different needs (live trading vs. research vs. backtesting)
- Helpers would bloat the framework and create maintenance burden
- Users can build their own helpers using the interfaces

**Consequence:** Slightly more code for users, but complete control over behavior.

### Trade-off 2: Type Safety vs. Runtime Flexibility

**Decision:** Type safety for financial primitives, flexibility for mechanisms.

```go
// ✅ Type-safe primitives (compile-time checking)
type Price struct { decimal.Decimal }
type Amount struct { decimal.Decimal }
// Cannot add Price + Amount (compile error)

// ✅ Flexible mechanisms (runtime polymorphism)
type Position interface {
    Value(m MarketSnapshot) (Amount, error)
}
// Any type can implement Position
```

**Rationale:**
- Financial math errors are catastrophic → enforce at compile time
- Mechanism variety is infinite → allow runtime flexibility
- Type system prevents common bugs without limiting extensibility

**Consequence:** More verbose primitive operations, but safer financial code.

### Trade-off 3: Batteries-Included vs. Minimal Core

**Decision:** Minimal core with reference implementations.

**Core (Always Included):**
- Interface definitions (mechanisms, strategy, position)
- Primitives (Price, Amount, Decimal)
- Portfolio management
- Backtest engine

**Reference (Included but Optional):**
- Concentrated liquidity pool (Uniswap V3)
- Black-Scholes options
- Perpetual futures

**Not Included (User Adds):**
- Protocol-specific implementations beyond references
- Data fetching/storage
- Risk management rules
- Optimization algorithms
- Live trading execution

**Rationale:**
- Core is stable and well-tested
- References demonstrate patterns
- Users customize based on their needs
- Reduces framework bloat and maintenance

**Consequence:** Users write more code, but framework stays focused and maintainable.

### Trade-off 4: Abstraction vs. Performance

**Decision:** Favor abstraction with escape hatches.

```go
// Framework uses interface for flexibility
type Position interface {
    Value(m MarketSnapshot) (Amount, error)
}

// But users can type-assert for performance-critical code
func optimizedPath(p Position) {
    if lpPos, ok := p.(*LPPosition); ok {
        // Fast path for specific type
    } else {
        // General path using interface
    }
}
```

**Rationale:**
- Most code benefits from abstraction (clarity, flexibility)
- Performance-critical sections can optimize selectively
- Interface overhead is negligible for I/O-bound operations (typical in quant)

**Consequence:** Slightly slower than monomorphic code, but difference is negligible in practice.

### Trade-off 5: Documentation vs. Code Generation

**Decision:** Comprehensive documentation over code generation.

**We provide:**
- Detailed godoc for all interfaces
- Step-by-step extension guide (EXTENDING.md)
- Multiple complete examples
- Architecture explanation (this document)

**We don't provide:**
- Code generators for mechanisms
- CLI scaffolding tools
- IDE plugins

**Rationale:**
- Documentation scales better than code generation
- Code generation creates maintenance burden
- Examples are more flexible than templates
- Users understand what they build from scratch

**Consequence:** More initial learning curve, but deeper understanding.

## Summary

go-crypto-quant-toolkit's architecture is designed for **extensibility through composition**:

1. **Minimal interfaces** define what components can do
2. **Users implement interfaces** in their own packages
3. **Framework composes implementations** without knowing concrete types
4. **Zero coupling** between implementations

This design enables:
- ✅ Adding new protocols without framework changes
- ✅ Testing with mocks and stubs
- ✅ Composing mechanisms from any source
- ✅ Evolving the framework without breaking existing code

The trade-off is more initial code for users, but the payoff is a framework that never limits what you can build.

For practical guidance on adding mechanisms, see [EXTENDING.md](EXTENDING.md).
