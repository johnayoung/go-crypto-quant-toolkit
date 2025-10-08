# Custom Mechanism Example

This example demonstrates the framework's extensibility by implementing a custom market mechanism without modifying any framework code.

## Overview

We implement a simple **constant-product AMM** (Uniswap V2 style) from scratch and use it in a strategy to prove that:
1. New mechanism types can be added without framework changes
2. Custom mechanisms work seamlessly with the backtest engine
3. The framework is truly mechanism-agnostic

## Custom Mechanism: ConstantProductPool

A simple constant-product automated market maker implementing the formula:

```
x * y = k (constant)
```

Where:
- `x` = reserve of token A
- `y` = reserve of token B  
- `k` = constant product

Price discovery: `price = y / x`

### Implementation

The `ConstantProductPool` struct implements two framework interfaces:

1. **mechanisms.MarketMechanism**
   ```go
   Mechanism() MechanismType
   Venue() string
   ```

2. **mechanisms.LiquidityPool**
   ```go
   Calculate(ctx, params) (PoolState, error)
   AddLiquidity(ctx, amounts) (PoolPosition, error)
   RemoveLiquidity(ctx, position) (TokenAmounts, error)
   ```

## Position Wrapper: CustomAMMPosition

Adapts the pool position to implement `strategy.Position`:

```go
type CustomAMMPosition struct {
    poolPosition mechanisms.PoolPosition
    pool         *ConstantProductPool
}

// Implements strategy.Position interface
func (cap *CustomAMMPosition) ID() string
func (cap *CustomAMMPosition) Type() PositionType
func (cap *CustomAMMPosition) Value(MarketSnapshot) (Amount, error)
```

## Strategy: CustomAMMStrategy

A simple strategy that:
1. Deposits liquidity into the custom AMM on first rebalance
2. Holds the position passively (can be extended with rebalancing logic)

## Running the Example

```bash
# From the repository root
go run examples/custom_mechanism/main.go
```

Expected output:
```
=== Custom Mechanism Example ===
Demonstrating framework extensibility

Step 1: Creating custom constant-product AMM...
  Pool ID: custom-eth-usdc-pool
  Type: Constant Product (x * y = k)
  Fee: 0.3%

Step 2: Verifying interface implementation...
  ✓ Implements MarketMechanism interface
  ✓ Implements LiquidityPool interface

Step 3: Creating strategy with custom mechanism...
  Initial deposit: 5 ETH + 10000 USDC

Step 4: Running backtest with framework engine...
  Generated 30 days of market data

Backtest Results:
  ...

=== Custom Mechanism Validation ===
Positions tracked by framework: 1

Position Details:
  ID: custom-eth-usdc-pool
  Type: liquidity_pool
  ✓ Successfully using CustomAMMPosition
  ✓ Pool Type: constant_product

=== Extensibility Demonstration Summary ===
✓ Created custom ConstantProductPool mechanism
✓ Implemented mechanisms.LiquidityPool interface
✓ Created CustomAMMPosition wrapper
✓ Implemented strategy.Position interface
✓ Used in strategy without framework modifications
✓ Backtest engine processed custom mechanism seamlessly
✓ Portfolio manager tracked custom position type

✅ Zero framework modifications required!
```

## Key Validation Points

### 1. Interface Implementation
The custom mechanism implements standard framework interfaces, ensuring compatibility:
- `mechanisms.MarketMechanism` - Base mechanism interface
- `mechanisms.LiquidityPool` - Pool-specific operations
- `strategy.Position` - Portfolio integration

### 2. Framework Integration
The custom mechanism works with all framework features:
- ✅ Backtest engine processes it
- ✅ Portfolio manager tracks it
- ✅ Actions apply to it
- ✅ Value calculations work
- ✅ Type system enforced

### 3. Zero Modifications
No changes to framework code:
- No edits to `pkg/mechanisms/`
- No edits to `pkg/strategy/`
- No edits to `pkg/backtest/`
- No edits to `pkg/primitives/`

## Extending This Example

### Add More Complexity

Implement advanced AMM features:
```go
// Implement concentrated liquidity ranges
func (p *ConstantProductPool) CalculateInRange(tickLower, tickUpper int) (PoolState, error)

// Add fee tier selection
func (p *ConstantProductPool) SetFeeTier(tier primitives.Decimal) error

// Implement price impact calculations
func (p *ConstantProductPool) GetPriceImpact(swapAmount primitives.Amount) (primitives.Decimal, error)
```

### Create Other Custom Mechanisms

The same pattern works for any mechanism type:

**Order Book**:
```go
type CustomOrderBook struct {
    bids []Order
    asks []Order
}

func (ob *CustomOrderBook) Mechanism() mechanisms.MechanismType {
    return mechanisms.MechanismTypeOrderBook
}
```

**Lending Protocol**:
```go
type CustomLendingPool struct {
    utilization primitives.Decimal
    interestRate primitives.Decimal
}

// Custom mechanism type
const MechanismTypeLending mechanisms.MechanismType = "lending"
```

**Batch Auction**:
```go
type BatchAuction struct {
    batchPeriod time.Duration
    clearingPrice primitives.Price
}

const MechanismTypeBatchAuction mechanisms.MechanismType = "batch_auction"
```

### Compose Multiple Custom Mechanisms

Combine custom mechanisms in a single strategy:
```go
type MultiCustomStrategy struct {
    customAMM     *ConstantProductPool
    customOrderBook *CustomOrderBook
}

func (s *MultiCustomStrategy) Rebalance(...) ([]strategy.Action, error) {
    // Arbitrage between custom AMM and custom order book
    ammPrice := s.customAMM.GetPrice()
    obPrice := s.customOrderBook.BestAsk()
    
    if ammPrice.LessThan(obPrice) {
        // Buy from AMM, sell on order book
        return []strategy.Action{...}, nil
    }
    
    return nil, nil
}
```

## Framework Extensibility Points

This example demonstrates the framework's extensibility philosophy:

### 1. Interface-First Design
Framework defines contracts, not implementations:
```go
// Framework provides the interface
type LiquidityPool interface {
    Calculate(ctx, params) (PoolState, error)
    // ...
}

// Users provide implementations
type YourCustomPool struct { ... }
func (p *YourCustomPool) Calculate(...) { ... }
```

### 2. Mechanism-Agnostic Core
Backtest engine and portfolio never reference concrete types:
```go
// Engine works with any Position implementation
func (e *Engine) Run(strat Strategy, snapshots []MarketSnapshot) (*Result, error) {
    // No knowledge of ConstantProductPool or CustomAMMPosition
}
```

### 3. Type Safety
All custom mechanisms benefit from primitives.Decimal precision:
```go
// Your calculations are precise
product := reserveA.Decimal().Mul(reserveB.Decimal())
// No floating-point errors
```

### 4. Composability
Mix framework implementations with custom ones:
```go
actions := []strategy.Action{
    NewAddPositionAction(frameworkLPPos),     // Framework's LP
    NewAddPositionAction(customAMMPos),        // Your custom AMM
    NewAddPositionAction(frameworkPerpPos),    // Framework's perpetual
}
```

## Comparison to Alternatives

| Approach             | Extensibility              | Type Safety | Complexity |
| -------------------- | -------------------------- | ----------- | ---------- |
| **This Framework**   | ✅ Interface-based          | ✅ Full      | Low        |
| Hardcoded mechanisms | ❌ Requires framework edits | ✅ Full      | Medium     |
| Plugin system        | ⚠️ Limited to plugin API    | ⚠️ Partial   | High       |
| Script-based         | ✅ Very flexible            | ❌ None      | High       |

## Real-World Use Cases

1. **Protocol-Specific Implementations**
   - Implement exact formulas from live protocols
   - Test strategies before deploying on-chain

2. **Research & Development**
   - Test novel mechanism designs
   - Compare mechanism performance

3. **Proprietary Mechanisms**
   - Keep custom logic private
   - Distribute framework publicly

4. **Academic Research**
   - Implement mechanisms from papers
   - Validate theoretical models

## Learning Path

To create your own custom mechanism:

1. **Choose mechanism type**
   - Decide: Pool, Derivative, OrderBook, or new type?

2. **Implement interfaces**
   - Start with `MarketMechanism`
   - Add mechanism-specific interface (e.g., `LiquidityPool`)

3. **Create position wrapper**
   - Implement `strategy.Position` interface
   - Bridge mechanism data to portfolio

4. **Write tests**
   - Test interface implementation
   - Verify calculations
   - Validate edge cases

5. **Use in strategy**
   - Create strategy using your mechanism
   - Run backtest
   - Analyze results

## Related Examples

- `examples/simple_lp` - Using framework's concentrated liquidity
- `examples/delta_neutral` - Multi-mechanism composition

## Framework Source

Study framework interfaces for guidance:
- `pkg/mechanisms/mechanism.go` - Base interfaces
- `pkg/mechanisms/liquidity_pool.go` - Pool interfaces
- `pkg/strategy/position.go` - Position interface
- `pkg/strategy/strategy.go` - Strategy interface
