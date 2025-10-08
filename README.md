# go-crypto-quant-toolkit

**An extensible, interface-first framework for building cryptocurrency quantitative trading strategies**

## Vision

`go-crypto-quant-toolkit` provides composable primitives and clear interfaces for implementing complex crypto trading strategies without coupling to specific market mechanisms. The framework is designed to be extended by the community with new market mechanisms, venues, and strategy types.

### Core Philosophy

- **Interface-First Design**: Define contracts, not implementations
- **Extensibility**: Add new market mechanisms without modifying framework code
- **Type Safety**: Prevent financial calculation errors at compile time
- **Precision**: Use decimal arithmetic for all monetary calculations
- **Composability**: Coordinate multiple mechanisms (AMMs, derivatives, order books) in a single strategy

## Why This Framework?

Modern crypto trading involves diverse market mechanisms:
- Automated Market Makers (AMMs) with various pricing curves
- Concentrated liquidity pools (Uniswap V3, etc.)
- Perpetual futures with funding rates
- Options markets with complex Greeks
- Order books across centralized and decentralized venues
- Novel mechanisms (batch auctions, intent pools, MEV strategies)

**The Problem**: Existing frameworks are either too specific (single venue/mechanism) or too general (no financial domain modeling).

**Our Solution**: Provide clear interfaces for market mechanisms that strategies compose, enabling:
- Multi-venue strategies (e.g., LP position hedged with perpetual)
- Novel mechanism implementations without framework changes
- Type-safe financial calculations preventing common bugs

## Features

### ✅ Type-Safe Financial Primitives
- `Price`, `Amount`, and `Decimal` types with compile-time safety
- Prevents invalid operations (e.g., `Price + Amount` won't compile)
- Decimal precision using `github.com/shopspring/decimal` (no float64 for money)

### 🔌 Extensible Market Mechanism Interfaces
- `LiquidityPool` interface for AMM-style mechanisms
- `Derivative` interface for options, perpetuals, futures
- `OrderBook` interface for limit order book trading
- Add new mechanism types by implementing interfaces

### 📊 Strategy Framework
- Portfolio management with position tracking
- Mechanism-agnostic strategy composition
- Action-based portfolio modifications
- Market data abstraction layer

### 🔄 Event-Driven Backtesting
- Test strategies across any combination of mechanisms
- Performance metrics (returns, Sharpe ratio, drawdown)
- Context-aware execution with cancellation support

### 🎯 Reference Implementations (Included)
- Concentrated Liquidity Pool (Uniswap V3-style)
- Black-Scholes Options Pricing
- Perpetual Futures with Funding Rates

## Installation

```bash
go get github.com/johnayoung/go-crypto-quant-toolkit
```

## Quick Start

Get started in 5 minutes with these examples.

### 1. Install

```bash
go get github.com/johnayoung/go-crypto-quant-toolkit
```

### 2. Use Type-Safe Primitives

```go
import "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"

// Create type-safe financial values
price := primitives.MustPrice(primitives.MustDecimal("1999.50"))
amount := primitives.MustAmount(primitives.MustDecimal("10.5"))

// Type-safe operations
totalValue := amount.MulPrice(price) // Amount × Price = Amount
// price.Add(amount) // ❌ Won't compile - cannot add Price and Amount

// Precise decimal arithmetic (no float64!)
newPrice := price.Mul(primitives.MustDecimal("1.05")) // 5% increase
```

### 3. Use Existing Mechanisms

```go
import (
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/implementations/concentrated_liquidity"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
)

// Create a Uniswap V3-style concentrated liquidity pool
pool, _ := concentrated_liquidity.NewPool(
    "usdc-eth-pool",
    common.HexToAddress("0xA0b8..."), // USDC address
    6,  // USDC decimals
    common.HexToAddress("0xC02a..."), // WETH address
    18, // WETH decimals
    constants.FeeAmount(3000), // 0.3% fee
)

// Calculate pool state
state, _ := pool.Calculate(ctx, mechanisms.PoolParams{
    ReserveA: primitives.MustAmount(primitives.MustDecimal("1000000")),
    ReserveB: primitives.MustAmount(primitives.MustDecimal("500")),
    Metadata: map[string]interface{}{
        "current_tick": 85176,
        "sqrt_price_x96": sqrtPriceX96,
    },
})

fmt.Printf("Spot price: %s\n", state.SpotPrice.Decimal())
```

### 4. Build a Strategy

```go
import (
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
)

type MyStrategy struct {
    pool mechanisms.LiquidityPool
}

func (s *MyStrategy) Rebalance(
    ctx context.Context,
    portfolio *strategy.Portfolio,
    market strategy.MarketSnapshot,
) ([]strategy.Action, error) {
    // Get current portfolio value
    value, _ := portfolio.TotalValue(market)
    
    // Your strategy logic here
    // Return actions to modify positions
    
    return []strategy.Action{}, nil
}
```

### 5. Run a Backtest

```go
import "github.com/johnayoung/go-crypto-quant-toolkit/pkg/backtest"

// Create engine with initial cash
engine := backtest.NewEngine(
    primitives.MustAmount(primitives.MustDecimal("100000")), // $100k
    backtest.WithLogging(true),
)

// Run strategy over historical data
result, _ := engine.Run(ctx, myStrategy, marketDataEvents)

// Analyze results
fmt.Printf("Total Return: %s%%\n", result.TotalReturn.Mul(primitives.MustDecimal("100")))
fmt.Printf("Sharpe Ratio: %s\n", result.SharpeRatio)
fmt.Printf("Max Drawdown: %s%%\n", result.MaxDrawdown.Mul(primitives.MustDecimal("100")))
```

### 6. Add Your Own Mechanism

```go
// 1. Implement the interface
type MyCustomPool struct {
    poolID   string
    tokenA   string
    tokenB   string
    feeRate  primitives.Decimal
}

// 2. Implement required methods
func (p *MyCustomPool) Mechanism() mechanisms.MechanismType {
    return mechanisms.MechanismTypeLiquidityPool
}

func (p *MyCustomPool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Your custom pricing logic
    price, _ := params.ReserveB.Decimal().Div(params.ReserveA.Decimal())
    
    return mechanisms.PoolState{
        SpotPrice: primitives.MustPrice(price),
        Liquidity: params.ReserveA,
        // ... other fields
    }, nil
}

// 3. Use it anywhere - framework automatically supports it!
```

See **[examples/custom_mechanism](examples/custom_mechanism/)** for a complete working example (~365 lines).

### Learn More

- **[EXTENDING.md](docs/EXTENDING.md)** - Complete guide to adding new mechanisms
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - Understand the design philosophy
- **[examples/](examples/)** - Three complete working examples

## Project Status

**Current Phase**: ✅ MVP Complete

### Completed
- ✅ Core primitives (Price, Amount, Decimal, Time, Duration) - 94.2% coverage
- ✅ Mechanism interface definitions (LiquidityPool, Derivative, OrderBook)
- ✅ Strategy framework (Portfolio, Position, Action, Strategy) - 94.8% coverage
- ✅ Reference implementations:
  - Concentrated Liquidity (Uniswap V3-style) - 78% coverage
  - Black-Scholes Options - 82.9% coverage
  - Perpetual Futures - 82.9% coverage
- ✅ Event-driven backtest engine - 83.6% coverage
- ✅ Complete examples (simple LP, delta-neutral, custom mechanism)
- ✅ Integration tests validating multi-mechanism strategies
- ✅ Comprehensive documentation

See [ROADMAP.md](docs/ROADMAP.md) for detailed implementation history.

## Documentation

- **[Quick Start Guide](#quick-start)** - Get started in 5 minutes
- **[Extending the Framework](docs/EXTENDING.md)** - Step-by-step guide for adding new mechanisms
- **[Architecture Guide](docs/ARCHITECTURE.md)** - Design philosophy and component relationships
- [Technical Specification](docs/SPEC.md) - Complete technical specification
- [Implementation Roadmap](docs/ROADMAP.md) - Development history and progress
- [Project Brief](docs/BRIEF.md) - Original project vision

### Examples

All examples are fully functional and can be run directly:

- **[Simple LP Strategy](examples/simple_lp/)** - Basic liquidity pool strategy with backtesting
- **[Delta-Neutral Strategy](examples/delta_neutral/)** - Multi-mechanism strategy (LP + perpetual hedge)
- **[Custom Mechanism](examples/custom_mechanism/)** - Adding a new mechanism without framework changes

```bash
# Run any example
go run examples/simple_lp/main.go
go run examples/delta_neutral/main.go
go run examples/custom_mechanism/main.go
```

## Design Principles

### 1. **Never use `float64` for money**
All financial values use `decimal.Decimal` to prevent precision errors.

### 2. **Accept interfaces, return structs**
Framework defines small, focused interfaces. Implementations return concrete types.

### 3. **Zero framework coupling in implementations**
Reference implementations should work without knowing about the framework.

### 4. **Fail fast with clear errors**
Type system prevents invalid operations at compile time. Runtime errors include context.

### 5. **Standard library first**
Minimize dependencies. Only add external libraries when necessary.

## Contributing

We welcome contributions! The framework is specifically designed to be extended by the community with:

- **New market mechanism implementations** - See [EXTENDING.md](docs/EXTENDING.md) for step-by-step guide
- Additional reference strategies
- Enhanced analytics and risk metrics
- Performance optimizations

### How to Contribute

1. **Read the guides**:
   - [EXTENDING.md](docs/EXTENDING.md) - Adding new mechanisms (4-step process)
   - [ARCHITECTURE.md](docs/ARCHITECTURE.md) - Understanding the design

2. **Follow best practices**:
   - Use `primitives.Decimal` for all financial calculations (never `float64`)
   - Achieve >80% test coverage for new implementations
   - Validate against reference values when implementing known protocols
   - Document all public types and methods with godoc

3. **Submit your work**:
   - Fork the repository
   - Create a feature branch
   - Add comprehensive tests
   - Submit a pull request with clear description

See [examples/custom_mechanism](examples/custom_mechanism/) for a complete example of adding a new mechanism.

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -cover

# Run with race detector
go test ./... -race

# Lint
golangci-lint run
```

## Requirements

- Go 1.21 or higher
- No external dependencies beyond `github.com/shopspring/decimal`

## License

MIT License - See [LICENSE](LICENSE) for details

## Related Projects

This framework complements but does not replace:
- Data collection tools (on-chain indexers, exchange APIs)
- Order execution systems (exchange integrations)
- Infrastructure (monitoring, deployment)

Focus is purely on **strategy modeling and backtesting** with extensible mechanism abstractions.

## Contact

- **Repository**: https://github.com/johnayoung/go-crypto-quant-toolkit
- **Issues**: https://github.com/johnayoung/go-crypto-quant-toolkit/issues

---

Built with ❤️ for the crypto quant community
