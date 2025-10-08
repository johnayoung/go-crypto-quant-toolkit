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

### Using Primitives

```go
import "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"

// Create type-safe financial values
price := primitives.MustPrice(primitives.NewDecimal(1999))
amount := primitives.MustAmount(primitives.NewDecimal(10))

// Type-safe operations
totalValue := amount.MulPrice(price) // Amount * Price = Amount (in currency units)
// price.Add(amount) // ❌ Won't compile - cannot add Price and Amount
```

### Implementing a Strategy

```go
import (
    "context"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

type MyStrategy struct {
    // Your strategy state
}

func (s *MyStrategy) Rebalance(
    ctx context.Context,
    portfolio *strategy.Portfolio,
    market strategy.MarketSnapshot,
) ([]strategy.Action, error) {
    // Analyze market conditions
    // Calculate desired positions
    // Return actions to modify portfolio
    return []strategy.Action{}, nil
}
```

### Adding a Custom Mechanism

The framework is designed for extensibility. You can add new market mechanisms without modifying any framework code:

```go
// 1. Define your mechanism implementing a framework interface
type MyCustomPool struct {
    // Your fields
}

// 2. Implement the LiquidityPool interface (or define your own)
func (p *MyCustomPool) Calculate(params mechanisms.PoolParams) (mechanisms.PoolState, error) {
    // Your custom logic
}

// 3. Use it in strategies - framework automatically supports it
```

See `examples/custom_mechanism/` for a complete working example.

## Project Status

**Current Phase**: MVP Development

### Completed
- ✅ Core primitives (Price, Amount, Decimal, Time, Duration)
- ✅ Type-safe arithmetic with >94% test coverage
- ✅ Go module setup with decimal dependency

### In Progress
- 🚧 Mechanism interface definitions
- 🚧 Strategy framework core
- 🚧 Reference implementations

### Planned
- ⏳ Backtest engine
- ⏳ Examples and integration tests
- ⏳ Comprehensive documentation

See [ROADMAP.md](docs/ROADMAP.md) for detailed implementation plan.

## Documentation

- [Technical Specification](docs/SPEC.md) - Architecture and design decisions
- [Implementation Roadmap](docs/ROADMAP.md) - Development plan and progress
- [Project Brief](docs/BRIEF.md) - Original project vision

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

- New market mechanism implementations
- Additional reference strategies
- Enhanced analytics and risk metrics
- Performance optimizations

Please see `CONTRIBUTING.md` (coming soon) for guidelines.

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
