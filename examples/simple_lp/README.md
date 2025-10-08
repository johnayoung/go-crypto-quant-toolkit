# Simple Liquidity Provision Strategy

This example demonstrates a basic passive liquidity provision strategy using concentrated liquidity pools (Uniswap V3 style).

## Overview

The strategy:
1. Provides liquidity to a WETH/USDC concentrated liquidity pool
2. Holds the position passively throughout the backtest period
3. Tracks portfolio value including LP position changes

This serves as a template for more sophisticated LP strategies that might include:
- Dynamic range rebalancing
- Fee collection and reinvestment
- Impermanent loss monitoring
- Multiple pool positions

## Strategy Components

### LPPosition
A wrapper that adapts `mechanisms.PoolPosition` to the `strategy.Position` interface. This demonstrates how to bridge mechanism-specific types with the strategy framework.

### SimpleLPStrategy
A passive LP strategy that:
- Provides initial liquidity on first rebalance
- Holds the position without modifications
- Allows the backtest engine to track value changes

## Running the Example

```bash
# From the repository root
go run examples/simple_lp/main.go
```

Expected output:
```
=== Simple LP Strategy Backtest ===

Generated 30 days of historical data
Running backtest...

=== Backtest Results ===
Initial Value: $100,000.00
Final Value: $XXX,XXX.XX
Total Return: X.XX%
...
```

## Key Concepts Demonstrated

1. **Position Wrapping**: Shows how to wrap mechanism-specific position types to implement the `strategy.Position` interface

2. **Market Snapshots**: Demonstrates storing pool-specific metadata (current tick, sqrt price) in market snapshots

3. **Position Valuation**: Calculates LP position value by determining what tokens would be received if withdrawing liquidity

4. **Action Composition**: Uses multiple actions (add position + adjust cash) to properly reflect capital allocation

## Extending This Example

To build a more sophisticated LP strategy:

1. **Add Rebalancing Logic**: Check if position is out of range and create new position with updated range
   ```go
   if currentTick < lp.tickLower || currentTick > lp.tickUpper {
       // Remove old position, add new one with updated range
   }
   ```

2. **Fee Collection**: Track and collect accumulated fees periodically
   ```go
   fees := calculateAccumulatedFees(lp.poolPosition)
   return []strategy.Action{
       strategy.NewAdjustCashAction(fees, "LP fees collected"),
   }
   ```

3. **Impermanent Loss Monitoring**: Compare LP value to holding underlying tokens
   ```go
   hodlValue := calculateHodlValue(initialTokens, currentPrices)
   lpValue := position.Value(snapshot)
   il := hodlValue.Sub(lpValue)
   ```

4. **Multi-Pool Strategy**: Provide liquidity to multiple pools with different characteristics
   ```go
   for _, pool := range pools {
       // Add position to each pool based on strategy logic
   }
   ```

## Integration with Real Data

To use real historical data:

1. Replace `createHistoricalSnapshots()` with actual data loading
2. Fetch pool state from blockchain or data provider
3. Use real tick values, sqrt prices, and liquidity metrics

Example data sources:
- The Graph (Uniswap V3 subgraph)
- Dune Analytics
- Direct blockchain queries via go-ethereum

## Related Examples

- `examples/delta_neutral` - Shows how to hedge LP exposure with derivatives
- `examples/custom_mechanism` - Demonstrates adding custom pool implementations

## Framework Concepts

This example validates:
- ✅ Position interface extensibility
- ✅ Strategy framework composition
- ✅ Backtest engine mechanism-agnostic design
- ✅ Market snapshot metadata storage
