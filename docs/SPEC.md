# MVP Technical Specification: go-crypto-quant-toolkit

**Project Type:** Go Library (Extensible Framework)

**Core Philosophy:** Provide composable primitives and clear interfaces that enable strategy implementation without prescribing specific mechanisms. The framework should make it easy to add new market mechanisms, venues, and strategy types without refactoring core abstractions.

## Core Requirements (from Brief)

### MVP Requirements
1. Provide clear interfaces for market mechanisms that strategies compose without implementation coupling
2. Enable adding new market mechanism implementations without modifying framework code
3. Provide 2-3 reference implementations demonstrating interface usage
4. Define composable strategy framework for coordinating multiple mechanisms across venues
5. Provide event-driven backtesting working regardless of strategy complexity

### Post-MVP (Community-Driven)
- Community-contributed mechanism implementations
- Optimization and risk analytics packages
- Advanced features (MEV, ML, ZK) through extension points

## Technology Stack

**Language:** Go 1.21+
- Interface-based extensibility (key requirement)
- Type safety for financial calculations
- Performance for backtesting
- Standard library first

**Dependencies (Minimal):**
- `github.com/shopspring/decimal` - Precise decimal arithmetic
- **No hard dependencies on specific implementations**

**Testing:**
- Standard `testing` package
- Property-based tests for interface contracts
- Integration tests validating extensibility

## Architecture: Interface-First Design

**Key Principle:** Define contracts, not implementations. Users compose strategies from any implementations of these contracts.

```
┌─────────────────────────────────────────────────────────────┐
│                        USER CODE                             │
│  • Implements Strategy interface                            │
│  • Chooses market mechanism implementations                 │
│  • Provides market data                                     │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                  STRATEGY FRAMEWORK                          │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Strategy Interface (User Implements)          │  │
│  │  Rebalance(ctx, portfolio, market) → Actions         │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Portfolio (Framework-Provided)                │  │
│  │  Tracks positions, cash, queries values              │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │         Position Interface (User Implements)          │  │
│  │  Value(market) → Amount                              │  │
│  │  Risk() → RiskMetrics                                │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│            MARKET MECHANISM INTERFACES                       │
│         (Framework defines, users implement)                 │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              MarketMechanism                          │  │
│  │  Common interface for all tradeable mechanisms       │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │LiquidityPool │  │  Derivative  │  │  OrderBook     │   │
│  │  Interface   │  │  Interface   │  │  Interface     │   │
│  │              │  │              │  │                │   │
│  │ • Calculate  │  │ • Price      │  │ • BestBid/Ask │   │
│  │ • AddLiq     │  │ • Greeks     │  │ • Place Order │   │
│  │ • RemoveLiq  │  │ • Settle     │  │ • Cancel      │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
│                                                              │
│  Users can add: BatchAuction, FlashLoan, IntentPool, etc.  │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│         REFERENCE IMPLEMENTATIONS (MVP: 2-3)                 │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │    ConcentratedLiquidityPool (LiquidityPool impl)    │  │
│  │    • Demonstrates LiquidityPool interface            │  │
│  │    • Provides tick math, IL calculations             │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │    BlackScholesOption (Derivative impl)              │  │
│  │    • Demonstrates Derivative interface               │  │
│  │    • Provides pricing, Greeks                        │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │    PerpetualFuture (Derivative impl)                 │  │
│  │    • Demonstrates Derivative interface               │  │
│  │    • Provides funding rate calculations              │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Community adds: ConstantProduct, OrderBook, FlashLoan...   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  PRIMITIVES (Shared Types)                   │
│  Price, Amount, Decimal, Time - used by all layers         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   BACKTEST ENGINE                            │
│  Event loop: Data → Strategy.Rebalance() → Portfolio       │
└─────────────────────────────────────────────────────────────┘
```

## Core Interfaces (Framework Contract)

### strategy/strategy.go
```go
// Strategy makes decisions based on portfolio and market state
type Strategy interface {
    // Rebalance is called on each market data event
    // Returns desired position changes
    Rebalance(ctx context.Context, p *Portfolio, m MarketSnapshot) ([]Action, error)
}

// Action represents a desired position change
type Action interface {
    Apply(p *Portfolio) error
}
```

### strategy/position.go
```go
// Position represents any tradeable position
// Users implement this for their mechanism types
type Position interface {
    // Value returns current position value
    Value(m MarketSnapshot) (primitives.Amount, error)
    
    // Risk returns position-specific risk metrics
    Risk(m MarketSnapshot) (RiskMetrics, error)
    
    // Type returns position classification
    Type() PositionType
}

// PositionType enum: Spot, LP, Option, Perpetual, OrderBook, etc.
type PositionType string
```

### mechanisms/mechanism.go
```go
// MarketMechanism is the base interface all mechanisms implement
type MarketMechanism interface {
    // Mechanism identifies the market mechanism type
    Mechanism() MechanismType
    
    // Venue identifies where this mechanism exists (optional)
    Venue() string
}

// LiquidityPool represents AMM-style liquidity provision
type LiquidityPool interface {
    MarketMechanism
    
    // Calculate returns position state for given inputs
    Calculate(params PoolParams) (PoolState, error)
    
    // AddLiquidity simulates adding liquidity
    AddLiquidity(amounts TokenAmounts) (PoolPosition, error)
    
    // RemoveLiquidity simulates removing liquidity
    RemoveLiquidity(position PoolPosition) (TokenAmounts, error)
}

// Derivative represents any derivative instrument
type Derivative interface {
    MarketMechanism
    
    // Price returns current derivative price
    Price(params PriceParams) (primitives.Price, error)
    
    // Greeks returns risk sensitivities (if applicable)
    Greeks(params PriceParams) (Greeks, error)
    
    // Settle calculates settlement value
    Settle() (primitives.Amount, error)
}

// OrderBook represents order book trading
type OrderBook interface {
    MarketMechanism
    
    // BestBid returns top bid price and size
    BestBid() (primitives.Price, primitives.Amount, error)
    
    // BestAsk returns top ask price and size
    BestAsk() (primitives.Price, primitives.Amount, error)
    
    // PlaceOrder simulates order placement
    PlaceOrder(order Order) (OrderID, error)
    
    // CancelOrder simulates order cancellation
    CancelOrder(id OrderID) error
}

// Future mechanisms implement these interfaces:
// - BatchAuction
// - IntentPool
// - FlashLoan
// - CrossChainBridge
// etc.
```

## Data Flow

**Extensibility-Focused Flow:**

1. User implements `Strategy` interface
2. User selects mechanism implementations (our ref impls or custom)
3. User provides market data (any format → MarketSnapshot)
4. Backtest engine calls `Strategy.Rebalance()` per event
5. Strategy uses mechanism interfaces to calculate positions
6. Strategy returns `Actions` to modify portfolio
7. Portfolio applies actions (mechanism-agnostic)
8. Repeat until data exhausted

**Key Insight:** Strategy code never knows concrete mechanism types, only interfaces. This enables swapping implementations without code changes.

## System Components

### primitives/
**Purpose:** Shared types used across all layers
**Exports:**
- `Price`, `Amount`, `Decimal` (financial types)
- `Time`, `Duration` (temporal types)
- Type-safe arithmetic preventing invalid operations

### mechanisms/
**Purpose:** Define mechanism interface contracts
**Exports:**
- `MarketMechanism` (base interface)
- `LiquidityPool` (AMM contract)
- `Derivative` (options, perps contract)
- `OrderBook` (CEX-style contract)
**Post-MVP:** Community adds BatchAuction, FlashLoan, etc.

### implementations/ (or impls/)
**Purpose:** Reference implementations of mechanism interfaces
**Exports (MVP):**
- `ConcentratedLiquidityPool` (implements LiquidityPool)
- `BlackScholesOption` (implements Derivative)
- `PerpetualFuture` (implements Derivative)
**Post-MVP:** Community adds more

### strategy/
**Purpose:** Strategy framework (portfolio, positions, actions)
**Exports:**
- `Strategy` interface
- `Portfolio` (position tracking)
- `Position` interface
- `Action` interface
- `MarketSnapshot` (market data abstraction)

### backtest/
**Purpose:** Event-driven backtesting engine
**Exports:**
- `Engine` (runs backtests)
- `Result` (performance metrics)
**Key:** Mechanism-agnostic, works with any Position types

## File Structure

```
go-crypto-quant-toolkit/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── docs/
│   ├── BRIEF.md
│   ├── SPEC.md
│   ├── EXTENDING.md          # How to add mechanisms
│   └── ARCHITECTURE.md        # Design decisions
│
├── primitives/
│   ├── types.go              # Price, Amount, Decimal
│   ├── time.go               # Time types
│   └── primitives_test.go
│
├── mechanisms/                # Interface definitions only
│   ├── mechanism.go          # Base MarketMechanism
│   ├── liquidity_pool.go     # LiquidityPool interface
│   ├── derivative.go         # Derivative interface
│   ├── orderbook.go          # OrderBook interface
│   └── mechanisms_test.go    # Interface contract tests
│
├── implementations/           # Reference implementations
│   ├── concentrated_liquidity/
│   │   ├── pool.go           # Implements LiquidityPool
│   │   ├── tick_math.go      # Tick calculations
│   │   ├── il.go             # Impermanent loss
│   │   └── pool_test.go
│   ├── blackscholes/
│   │   ├── option.go         # Implements Derivative
│   │   ├── greeks.go         # Greeks calculations
│   │   └── option_test.go
│   └── perpetual/
│       ├── future.go         # Implements Derivative
│       ├── funding.go        # Funding rate logic
│       └── future_test.go
│
├── strategy/
│   ├── strategy.go           # Strategy interface
│   ├── portfolio.go          # Portfolio management
│   ├── position.go           # Position interface
│   ├── action.go             # Action interface
│   ├── market.go             # MarketSnapshot
│   └── strategy_test.go
│
├── backtest/
│   ├── engine.go             # Backtest orchestration
│   ├── result.go             # Performance metrics
│   └── backtest_test.go
│
└── examples/
    ├── simple_lp/
    │   ├── main.go           # Concentrated LP example
    │   └── README.md
    ├── delta_neutral/
    │   ├── main.go           # LP + perp hedge
    │   └── README.md
    └── custom_mechanism/
        ├── main.go           # Shows adding new mechanism
        └── README.md
```

## Integration Patterns

### MVP Usage Pattern

**1. Use existing mechanism:**
```go
import (
    "github.com/yourorg/go-crypto-quant-toolkit/implementations/concentrated_liquidity"
    "github.com/yourorg/go-crypto-quant-toolkit/strategy"
)

type MyStrategy struct {
    pool *concentrated_liquidity.Pool
}

func (s *MyStrategy) Rebalance(
    ctx context.Context,
    p *strategy.Portfolio,
    m strategy.MarketSnapshot,
) ([]strategy.Action, error) {
    // Use pool.Calculate() to determine position
    // Return actions
}
```

**2. Add custom mechanism (extensibility demo):**
```go
// User creates new mechanism type
type BatchAuction struct {
    // ... fields
}

// Implement mechanism interface
func (b *BatchAuction) Mechanism() mechanisms.MechanismType {
    return "batch_auction"
}

// Implement mechanism-specific methods
func (b *BatchAuction) SubmitOrder(order Order) error {
    // ... logic
}

// Use in strategy WITHOUT changing framework
type AuctionStrategy struct {
    auction *BatchAuction  // Custom mechanism
}

func (s *AuctionStrategy) Rebalance(...) ([]strategy.Action, error) {
    // Framework doesn't care about BatchAuction specifics
    // Works because framework only knows Position interface
}
```

**3. Run backtest:**
```go
engine := backtest.NewEngine()
results := engine.Run(
    ctx,
    myStrategy,
    marketData,
    initialPortfolio,
)
```

### Post-MVP Extensions

**Adding new mechanism category (e.g., MEV):**
```go
// 1. Define new interface in mechanisms/
type MEVOpportunity interface {
    MarketMechanism
    Simulate() (profit primitives.Amount, err error)
    Execute() error
}

// 2. Community implements
type SandwichAttack struct { ... }
func (s *SandwichAttack) Simulate() { ... }

// 3. Use in strategies - framework needs ZERO changes
```

## Extension Points (Documented)

### docs/EXTENDING.md will explain:

1. **Adding Market Mechanism:**
   - Define interface in `mechanisms/`
   - Implement in `implementations/yourname/`
   - Document interface contract
   - Add property-based tests

2. **Adding Position Type:**
   - Implement `Position` interface
   - Add to portfolio
   - Backtest engine handles automatically

3. **Adding Venue Support:**
   - Mechanism implementations specify venue
   - Framework doesn't care about venue differences

4. **Adding Analytics:**
   - Build on Portfolio interface
   - No framework changes needed

## Implementation Priority

### Phase 1: Core Interfaces (Week 1)
- Primitives types
- Mechanism interfaces (base + 3 categories)
- Strategy interfaces
- **Goal:** Framework contracts defined

### Phase 2: One Reference Implementation (Week 2)
- ConcentratedLiquidityPool
- Full tests demonstrating interface compliance
- **Goal:** Prove interface design works

### Phase 3: Two More Implementations (Week 3)
- BlackScholesOption
- PerpetualFuture
- **Goal:** Validate composability

### Phase 4: Strategy Framework (Week 4)
- Portfolio management
- Position tracking
- Action system
- **Goal:** Strategies can coordinate mechanisms

### Phase 5: Backtest Engine (Week 5)
- Event loop
- Basic metrics
- **Goal:** End-to-end backtesting works

### Phase 6: Documentation & Examples (Week 6)
- EXTENDING.md (how to add mechanisms)
- Three examples
- README polish
- **Goal:** Others can extend framework

## Testing Strategy

**Interface Contract Tests:**
- Every interface has property-based tests
- Any implementation must pass interface tests
- Validates substitutability

**Integration Tests:**
- Multi-mechanism strategies
- Prove extensibility (add mechanism without framework changes)

**Reference Implementation Tests:**
- Validate against known correct values
- Concentrated liquidity vs UniV3 contracts (0.01% tolerance)

## Success Criteria

**MVP is complete when:**
1. All core interfaces defined with clear contracts
2. Three reference implementations demonstrate different mechanism types
3. Example shows adding custom mechanism without framework modification
4. User can compose multi-mechanism strategy in <500 lines
5. Documentation explains extension patterns clearly
6. Property-based tests validate interface contracts

## Design Decisions (Rationale)

**Why Interface-First?**
- Market making strategies list shows massive variety
- Can't predict all mechanism types needed
- Users need extension without forking

**Why Minimal Reference Implementations?**
- Community will contribute mechanisms
- Framework shouldn't be bottleneck
- Quality over quantity for MVP

**Why No Hard Dependencies?**
- Users might prefer different math libraries
- Keeps framework lightweight
- Optional adapter pattern for popular libraries

**Why Mechanism Categories?**
- Groups related interfaces (pools, derivatives, orderbooks)
- But doesn't prevent new categories
- Users can add FlashLoan, MEV, etc.

## Non-Goals (Explicitly Out of Scope)

- ❌ Exhaustive mechanism implementations (community-driven)
- ❌ Data collection (users provide)
- ❌ Order execution (simulation only)
- ❌ Venue-specific clients
- ❌ UI/visualization
- ❌ Prescriptive strategy patterns (framework enables, doesn't constrain)