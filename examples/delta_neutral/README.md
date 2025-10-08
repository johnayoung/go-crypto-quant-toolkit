# Delta-Neutral Strategy Example

This example demonstrates a delta-neutral strategy that combines a concentrated liquidity position with a perpetual futures hedge to minimize directional price risk while capturing fees and funding rates.

## Overview

The strategy:
1. **Provides liquidity** to a WETH/USDC concentrated liquidity pool to earn trading fees
2. **Opens a short perpetual** position to hedge the ETH price exposure from the LP position
3. **Maintains delta neutrality** to reduce sensitivity to ETH price movements
4. **Captures multiple revenue streams**: LP fees + potential funding rate profits

## Delta-Neutral Concept

A delta-neutral position has minimal exposure to price movements in the underlying asset. For an LP position holding ETH and USDC:

- **LP Position**: Has positive delta (gains when ETH price increases)
- **Short Perpetual**: Has negative delta (gains when ETH price decreases)
- **Combined Position**: Net delta ≈ 0 (insensitive to small price changes)

## Multi-Mechanism Composition

This example demonstrates the framework's key value proposition: **composing multiple market mechanisms in a single strategy**.

### Mechanisms Used

1. **Concentrated Liquidity Pool** (`pkg/implementations/concentrated_liquidity`)
   - Uniswap V3 style liquidity provision
   - Earns trading fees
   - Creates directional exposure

2. **Perpetual Futures** (`pkg/implementations/perpetual`)
   - Perpetual swap contract
   - Provides hedge against LP exposure
   - May earn/pay funding rates

### Integration Points

Both mechanisms integrate through the `strategy.Position` interface:
```go
type Position interface {
    Value(snapshot MarketSnapshot) (primitives.Amount, error)
    Type() PositionType
    ID() string
}
```

The backtest engine and portfolio manager work with both position types without any mechanism-specific knowledge, validating the framework's extensibility design.

## Running the Example

```bash
# From the repository root
go run examples/delta_neutral/main.go
```

Expected output:
```
=== Delta-Neutral Strategy Backtest ===
Combining LP position with perpetual hedge

Generated 30 days of market data
Running backtest...

=== Backtest Results ===
...

=== Position Analysis ===
Final positions: 2

Position: eth-usdc-pool
  Type: liquidity_pool
  Final Value: $XXX,XXX.XX

Position: perp-eth-hedge
  Type: perpetual
  Final Value: $XXX,XXX.XX

=== Delta Analysis ===
LP Position Value: $XXX,XXX.XX
Perpetual Position Value: $XXX,XXX.XX
Net Position Value: $XXX,XXX.XX
```

## Code Structure (<350 lines)

The example includes:
- **Position wrappers** (60 lines): `LPPosition`, `PerpPosition`
- **Strategy implementation** (80 lines): `DeltaNeutralStrategy`
- **Data generation** (40 lines): `createHistoricalSnapshots`
- **Main execution** (80 lines): Setup, backtest, analysis
- **Documentation** (90 lines): Comments explaining logic

Total: ~350 lines including comments, demonstrating that complex multi-mechanism strategies remain concise and readable.

## Key Concepts Demonstrated

### 1. Multi-Mechanism Strategy
Shows how a single strategy can coordinate positions across different mechanism types without coupling to implementations.

### 2. Position Composition
Both LP and perpetual positions implement `strategy.Position`, allowing uniform portfolio management.

### 3. Delta Management
Calculates net position delta by combining individual position values, enabling risk monitoring.

### 4. Market Data Integration
Uses `MarketSnapshot.Get()` for mechanism-specific metadata (pool state, funding rates) without extending the core interface.

## Extending This Strategy

### Dynamic Rebalancing
Monitor delta drift and rebalance when it exceeds thresholds:
```go
currentDelta := calculateNetDelta(portfolio, snapshot)
if currentDelta.Abs().GreaterThan(threshold) {
    adjustHedgeSize := calculateRequiredAdjustment(currentDelta)
    return []strategy.Action{
        strategy.NewReplacePositionAction("perp-eth-hedge", newPerpPosition),
    }
}
```

### Multiple Pools
Provide liquidity across multiple pools with different characteristics:
```go
for _, poolConfig := range configs {
    lpPos := createLPPosition(poolConfig)
    hedgePos := calculateHedge(lpPos)
    actions = append(actions,
        strategy.NewAddPositionAction(lpPos),
        strategy.NewAddPositionAction(hedgePos),
    )
}
```

### Advanced Hedging
Use options instead of perpetuals for asymmetric payoffs:
```go
// Replace short perpetual with put option
putOption := createPutOption(strike, expiry)
return []strategy.Action{
    strategy.NewRemovePositionAction("perp-eth-hedge"),
    strategy.NewAddPositionAction(putOption),
}
```

### Fee Optimization
Track and collect fees from both mechanisms:
```go
lpFees := calculateLPFees(lpPosition, snapshot)
fundingPaid := calculateFundingPayments(perpPosition, snapshot)
netIncome := lpFees.Sub(fundingPaid)
// Reinvest if profitable
if netIncome.IsPositive() {
    return []strategy.Action{
        strategy.NewAdjustCashAction(netIncome, "fees collected"),
    }
}
```

## Real-World Considerations

### Capital Efficiency
- Use leverage on perpetual side to reduce capital requirements
- Optimize tick ranges for LP position based on historical volatility
- Consider cross-margin opportunities

### Risk Management
- Monitor liquidation risk on leveraged perpetual
- Set maximum delta deviation thresholds
- Implement emergency exit procedures

### Cost Analysis
- Trading fees for rebalancing
- Funding rate costs/revenues
- Gas costs for on-chain operations
- Slippage on larger position sizes

### Market Conditions
- Strategy performs best in range-bound markets (earns fees, no IL)
- May underperform in strong trending markets (IL despite hedge)
- Funding rates can be favorable or unfavorable depending on market sentiment

## Framework Validation

This example validates critical framework requirements:

✅ **Multi-mechanism composition**: LP + Perpetual work together seamlessly  
✅ **Mechanism-agnostic design**: Backtest engine treats both position types uniformly  
✅ **Extensibility**: No framework modifications required  
✅ **Type safety**: All financial calculations use primitives.Decimal  
✅ **Conciseness**: Complete strategy in <500 lines (requirement met)

## Related Examples

- `examples/simple_lp` - Basic LP strategy without hedging
- `examples/custom_mechanism` - Creating custom mechanism implementations

## Further Reading

- **Delta-neutral strategies**: [Investopedia - Delta Neutral](https://www.investopedia.com/terms/d/deltaneutral.asp)
- **Impermanent loss**: Understanding LP risk
- **Funding rates**: Perpetual futures mechanics
- **Greeks**: Option sensitivity measures
