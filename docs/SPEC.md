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

## Architecture: Position-First Design

**Key Principle:** Mechanisms are stateless calculators, Positions are stateful containers, Strategies are composers.

**Architecture Evolution:**
- **v1.0 (MVP)**: Mechanism-first - strategies directly use mechanism interfaces
- **v2.0 (Commits 9-15)**: Position-first - positions wrap mechanisms and manage state

The position-first architecture emerged from analyzing complex strategies (see CLASSIC delta-neutral in `docs/strategies/delta-neutral/CLASSIC.md`). Real strategies require:
1. **State management** - Track collateral, debt, fees, PnL across protocols
2. **Lifecycle operations** - Open, modify, close with validation
3. **Protocol-specific logic** - Health factors, liquidation prices, rewards
4. **Natural composition** - Combine positions like building blocks

```
┌─────────────────────────────────────────────────────────────┐
│                        USER CODE                             │
│  • Implements Strategy interface                            │
│  • Composes positions from building blocks                  │
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
│  │         Portfolio (Position Container)                │  │
│  │  Aggregates positions, calculates total value        │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│           POSITION LAYER (Stateful Building Blocks)          │
│                    Commits 9-12                              │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │UniV3Position │  │LendingPosition│ │PerpetualPos  │     │
│  │              │  │               │  │              │     │
│  │• Mint()      │  │• Supply()     │  │• Open()      │     │
│  │• Burn()      │  │• Borrow()     │  │• Close()     │     │
│  │• Fees        │  │• Health()     │  │• PnL         │     │
│  │• IL          │  │• Rewards      │  │• Funding     │     │
│  └──────┬───────┘  └──────┬────────┘  └──────┬───────┘     │
│         │uses              │uses               │uses         │
└─────────┼──────────────────┼───────────────────┼─────────────┘
          │                  │                   │
          ▼                  ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│     MARKET MECHANISM INTERFACES (Stateless Calculators)      │
│         (Framework defines, users implement)                 │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              MarketMechanism                          │  │
│  │  Common interface for all tradeable mechanisms       │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────┐   │
│  │LiquidityPool │  │  Derivative  │  │LendingProtocol │   │
│  │  Interface   │  │  Interface   │  │  Interface     │   │
│  │              │  │              │  │  (Commit 11)   │   │
│  │ • Calculate  │  │ • Price      │  │ • SupplyAPY   │   │
│  │ • AddLiq     │  │ • Greeks     │  │ • BorrowAPY   │   │
│  │ • RemoveLiq  │  │ • Settle     │  │ • Health      │   │
│  └──────────────┘  └──────────────┘  └────────────────┘   │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  OrderBook   │  │BiCurrencyVlt │  • More to come...    │
│  │  Interface   │  │  (Swaps)     │                        │
│  └──────────────┘  └──────────────┘                        │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│         REFERENCE IMPLEMENTATIONS                            │
│         (Calculators used by Positions)                      │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  concentrated_liquidity/calculator.go                 │  │
│  │    → Implements LiquidityPool (stateless)            │  │
│  │    → Used by: UniV3Position                          │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  aave/calculator.go (Commit 11)                       │  │
│  │    → Implements LendingProtocol (stateless)          │  │
│  │    → Used by: LendingPosition                        │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  perpetual/future.go                                  │  │
│  │    → Implements Derivative (stateless)               │  │
│  │    → Used by: PerpetualPosition                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  Community adds: More calculators + position types...       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                  PRIMITIVES (Shared Types)                   │
│  Price, Amount, Decimal, Time - used by all layers         │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                   BACKTEST ENGINE                            │
│  Event loop: Data → Strategy.Rebalance() → Portfolio       │
│  (Position-agnostic, works with any position types)         │
└─────────────────────────────────────────────────────────────┘
```

**Key Architectural Features:**

1. **Separation of Concerns**: Mechanisms calculate, Positions manage state
2. **Composability**: Build complex strategies by combining position types
3. **Extensibility**: Add new position types or mechanisms independently
4. **Type Safety**: All financial math uses primitives.Decimal

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

All components live under `pkg/` following Go best practices for library projects.

### pkg/primitives/
**Purpose:** Shared types used across all layers
**Exports:**
- `Price`, `Amount`, `Decimal` (financial types)
- `Time`, `Duration` (temporal types)
- Type-safe arithmetic preventing invalid operations

### pkg/mechanisms/
**Purpose:** Define mechanism interface contracts (stateless calculators)
**Exports:**
- `MarketMechanism` (base interface)
- `LiquidityPool` (AMM contract)
- `Derivative` (options, perps contract)
- `LendingProtocol` (supply/borrow contract) - Commit 11
- `OrderBook` (CEX-style contract)
**Post-MVP:** Community adds BatchAuction, FlashLoan, etc.

### pkg/positions/ (Commits 9-12)
**Purpose:** Composable position implementations with state management
**Exports:**
- `AbstractPosition` interface (extends strategy.Position)
- `Portfolio` (moved from strategy, backward compatible)
- `UniV3Position` (concentrated liquidity with mint/burn/fees)
- `LendingPosition` (supply/borrow with health factor)
- `PerpetualPosition` (long/short with funding/PnL)
- `OptionPosition` (calls/puts with Greeks)
- `SpotPosition` (simple holdings)
- `BiCurrencyPosition` (token pair with swaps)
- `PositionManager` (coordinates position operations)

### pkg/implementations/
**Purpose:** Reference implementations of mechanism interfaces (calculators)
**Exports (v1.0-v2.0):**
- `concentrated_liquidity/calculator.go` (implements LiquidityPool) - Commit 13
- `aave/calculator.go` (implements LendingProtocol) - Commit 11
- `blackscholes/option.go` (implements Derivative)
- `perpetual/future.go` (implements Derivative)
**Note:** Commit 13 refactors to separate calculator (stateless) from position (stateful)
**Post-MVP:** Community adds more

### pkg/strategy/
**Purpose:** Strategy framework (strategy interface, actions, market data)
**Exports:**
- `Strategy` interface
- `Position` interface (base, extended by positions package)
- `Action` interface
- `MarketSnapshot` (market data abstraction)
**Note:** Portfolio moved to `pkg/positions/` in Commit 9 (backward compatible shim maintained)

### pkg/backtest/
**Purpose:** Event-driven backtesting engine
**Exports:**
- `Engine` (runs backtests)
- `Result` (performance metrics)
**Key:** Mechanism-agnostic, works with any Position types

## File Structure

Following the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) best practices:

```
go-crypto-quant-toolkit/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
│
├── docs/                      # Design and user documentation
│   ├── BRIEF.md
│   ├── SPEC.md               # Technical spec & architecture (this file)
│   ├── ROADMAP.md            # Implementation roadmap
│   ├── EXTENDING.md          # How to add mechanisms & positions
│   └── strategies/
│       └── delta-neutral/
│           └── CLASSIC.md    # Complex strategy reference
│
├── pkg/                       # Public library code (importable by external projects)
│   ├── primitives/
│   │   ├── types.go          # Price, Amount, Decimal
│   │   ├── time.go           # Time types
│   │   └── primitives_test.go
│   │
│   ├── mechanisms/            # Interface definitions (stateless calculators)
│   │   ├── mechanism.go      # Base MarketMechanism
│   │   ├── liquidity_pool.go # LiquidityPool interface
│   │   ├── derivative.go     # Derivative interface
│   │   ├── lending.go        # LendingProtocol interface (Commit 11)
│   │   ├── orderbook.go      # OrderBook interface
│   │   └── mechanisms_test.go # Interface contract tests
│   │
│   ├── positions/             # Position implementations (Commits 9-12)
│   │   ├── base.go           # AbstractPosition interface
│   │   ├── portfolio.go      # Portfolio (moved from strategy)
│   │   ├── spot.go           # SpotPosition
│   │   ├── bicurrency.go     # BiCurrencyPosition (swaps)
│   │   ├── concentrated_lp.go # UniV3Position
│   │   ├── perpetual.go      # PerpetualPosition
│   │   ├── option.go         # OptionPosition
│   │   ├── lending.go        # LendingPosition
│   │   ├── manager.go        # PositionManager
│   │   ├── events.go         # PositionEvent tracking
│   │   └── positions_test.go
│   │
│   ├── implementations/       # Reference implementations (calculators)
│   │   ├── concentrated_liquidity/
│   │   │   ├── calculator.go # LiquidityPool impl (Commit 13)
│   │   │   └── pool_test.go
│   │   ├── aave/             # (Commit 11)
│   │   │   ├── calculator.go # LendingProtocol impl
│   │   │   └── aave_test.go
│   │   ├── blackscholes/
│   │   │   ├── option.go     # Derivative impl
│   │   │   └── option_test.go
│   │   └── perpetual/
│   │       ├── future.go     # Derivative impl
│   │       └── future_test.go
│   │
│   ├── strategy/
│   │   ├── strategy.go       # Strategy interface
│   │   ├── position.go       # Position interface (base)
│   │   ├── action.go         # Action interface
│   │   ├── market.go         # MarketSnapshot
│   │   └── strategy_test.go
│   │   # Note: Portfolio moved to pkg/positions/ (Commit 9)
│   │
│   └── backtest/
│       ├── engine.go         # Backtest orchestration
│       ├── result.go         # Performance metrics
│       └── backtest_test.go
│
└── examples/                  # Example applications demonstrating library usage
    ├── simple_lp/
    │   ├── main.go           # Concentrated LP example
    │   └── README.md
    ├── delta_neutral/
    │   ├── main.go           # LP + perp hedge
    │   └── README.md
    ├── classic_delta_neutral/ # (Commit 14)
    │   ├── main.go           # Full CLASSIC strategy (lending+LP+perp)
    │   └── README.md
    └── custom_mechanism/
        ├── main.go           # Shows adding new mechanism
        └── README.md
```

**Directory Structure Rationale:**
- **`pkg/`**: Contains all public library code following Go best practices. This signals to users that these packages are safe to import and use in their own projects.
- **`docs/`**: Project documentation at root level as per golang-standards.
- **`examples/`**: Example applications demonstrating framework usage (not library code).
- No `cmd/` or `internal/` yet as this is a pure library framework without private implementation details or command-line tools.

## Integration Patterns

### MVP Usage Pattern

**1. Use existing mechanism:**
```go
import (
    "github.com/yourorg/go-crypto-quant-toolkit/pkg/implementations/concentrated_liquidity"
    "github.com/yourorg/go-crypto-quant-toolkit/pkg/strategy"
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