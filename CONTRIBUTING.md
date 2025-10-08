# Contributing to go-crypto-quant-toolkit

Thank you for your interest in contributing to go-crypto-quant-toolkit! This framework is designed to be extended by the community with new market mechanisms, strategies, and improvements.

## Table of Contents
- [Ways to Contribute](#ways-to-contribute)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing Requirements](#testing-requirements)
- [Documentation Requirements](#documentation-requirements)
- [Pull Request Process](#pull-request-process)
- [Community Guidelines](#community-guidelines)

## Ways to Contribute

### 1. Add New Mechanism Implementations

The most valuable contributions are new market mechanism implementations that others can use:

- **AMM implementations**: Curve stable pools, Balancer weighted pools, constant-product pools, etc.
- **Derivative implementations**: Exotic options, structured products, interest rate swaps, etc.
- **Order book implementations**: CEX-style limit order books, Dutch auctions, batch auctions
- **Novel mechanisms**: Flash loan pools, MEV strategies, intent-based systems, cross-chain bridges

**See [EXTENDING.md](docs/EXTENDING.md) for a step-by-step guide.**

### 2. Improve Documentation

- Clarify existing documentation
- Add more examples and use cases
- Create tutorials for specific strategies
- Translate documentation (if multilingual support is added)

### 3. Add Reference Strategies

Implement common strategies demonstrating framework usage:

- Market making strategies
- Arbitrage strategies (DEX-DEX, CEX-DEX)
- Yield farming optimizers
- Risk-adjusted portfolio strategies

### 4. Enhance Testing

- Add property-based tests for interface contracts
- Improve test coverage (target >80%)
- Add benchmark tests for performance-critical code
- Create integration tests for complex scenarios

### 5. Report Bugs

Found a bug? Please open an issue with:

- Clear description of the problem
- Steps to reproduce
- Expected vs. actual behavior
- Go version and OS
- Minimal code example demonstrating the issue

### 6. Suggest Improvements

Have ideas for improvements? Open an issue with:

- Clear description of the proposal
- Rationale (why is this valuable?)
- Potential implementation approach
- Any breaking changes or compatibility concerns

## Getting Started

### Prerequisites

- **Go 1.21+** installed
- Familiarity with Go interfaces and the standard library
- Understanding of financial calculations and crypto market mechanisms
- Git for version control

### Setup Development Environment

1. **Fork the repository** on GitHub

2. **Clone your fork**:
```bash
git clone https://github.com/YOUR_USERNAME/go-crypto-quant-toolkit.git
cd go-crypto-quant-toolkit
```

3. **Add upstream remote**:
```bash
git remote add upstream https://github.com/johnayoung/go-crypto-quant-toolkit.git
```

4. **Install dependencies**:
```bash
go mod download
```

5. **Verify everything works**:
```bash
go test ./...
go run examples/simple_lp/main.go
```

### Project Structure

```
go-crypto-quant-toolkit/
├── docs/                     # Documentation
│   ├── EXTENDING.md         # How to add mechanisms
│   ├── ARCHITECTURE.md      # Design philosophy
│   ├── SPEC.md              # Technical specification
│   └── ROADMAP.md           # Development history
├── examples/                # Complete working examples
│   ├── simple_lp/          # Basic LP strategy
│   ├── delta_neutral/      # Multi-mechanism strategy
│   └── custom_mechanism/   # Custom implementation example
├── pkg/                     # Public library code
│   ├── primitives/         # Type-safe financial primitives
│   ├── mechanisms/         # Mechanism interfaces
│   ├── strategy/           # Strategy framework
│   ├── backtest/           # Backtest engine
│   └── implementations/    # Reference implementations
│       ├── concentrated_liquidity/
│       ├── blackscholes/
│       └── perpetual/
├── CONTRIBUTING.md          # This file
├── README.md               # Project overview
└── LICENSE                 # MIT License
```

## Development Workflow

### 1. Create a Feature Branch

```bash
# Update your local main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/my-new-mechanism
```

Use descriptive branch names:
- `feature/curve-stable-pool` - New feature
- `fix/decimal-precision-bug` - Bug fix
- `docs/improve-extending-guide` - Documentation
- `test/liquidity-pool-contracts` - Testing improvements

### 2. Make Your Changes

Follow the [Code Standards](#code-standards) and [Testing Requirements](#testing-requirements).

### 3. Commit Your Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add Curve stable pool implementation

- Implement mechanisms.LiquidityPool interface
- Add StableSwap formula calculations
- Include comprehensive tests with >85% coverage
- Validate against Curve protocol reference values
"
```

**Commit message format**:
- First line: Brief summary (50 chars or less)
- Blank line
- Detailed description with bullet points
- Reference issues if applicable (#123)

### 4. Push and Create Pull Request

```bash
git push origin feature/my-new-mechanism
```

Then create a Pull Request on GitHub with:
- Clear title describing the change
- Description explaining what and why
- Link to any related issues
- Checklist of completed items (see PR template)

## Code Standards

### Go Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go).

**Key points**:
- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common mistakes
- Use `golangci-lint run` for comprehensive linting (if available)
- Keep functions small and focused
- Use meaningful variable names (not `x`, `y`, `temp`)

### Financial Calculations

**CRITICAL**: Always use `primitives.Decimal` for financial math.

```go
// ❌ NEVER DO THIS
floatPrice := float64(reserveB) / float64(reserveA)

// ✅ ALWAYS DO THIS
priceDecimal, err := reserveB.Decimal().Div(reserveA.Decimal())
if err != nil {
    return fmt.Errorf("failed to calculate price: %w", err)
}
price := primitives.MustPrice(priceDecimal)
```

**Rules**:
1. Never use `float64` or `float32` for money, prices, or quantities
2. Always handle errors from `Decimal` operations (e.g., division by zero)
3. Document precision and rounding behavior
4. Validate inputs (check for zero, negative, out-of-bounds values)

### Interface Implementation

When implementing framework interfaces:

```go
// ✅ Compile-time interface check
var _ mechanisms.LiquidityPool = (*YourPool)(nil)

// ✅ Document interface compliance
// YourPool implements mechanisms.LiquidityPool for [protocol name].
type YourPool struct { ... }

// ✅ Implement ALL methods with proper signatures
func (p *YourPool) Mechanism() mechanisms.MechanismType {
    return mechanisms.MechanismTypeLiquidityPool
}

func (p *YourPool) Venue() string {
    return "your-protocol"
}

// ... implement other required methods
```

### Error Handling

```go
// ✅ Define custom errors
var (
    ErrInvalidReserves = errors.New("reserves must be positive")
    ErrInsufficientLiquidity = errors.New("insufficient liquidity")
)

// ✅ Validate inputs early
func (p *Pool) Calculate(ctx context.Context, params PoolParams) (PoolState, error) {
    if params.ReserveA.IsZero() {
        return PoolState{}, ErrInvalidReserves
    }
    // ... rest of function
}

// ✅ Wrap errors with context
result, err := someOperation()
if err != nil {
    return PoolState{}, fmt.Errorf("operation failed: %w", err)
}
```

### Context Usage

```go
// ✅ Accept context for cancellation
func (p *Pool) Calculate(ctx context.Context, params PoolParams) (PoolState, error) {
    // Check context if doing expensive work
    select {
    case <-ctx.Done():
        return PoolState{}, ctx.Err()
    default:
    }
    
    // ... calculations
}
```

### Documentation

**Every public type and function must have godoc comments**:

```go
// Pool implements mechanisms.LiquidityPool for Uniswap V3-style concentrated liquidity.
//
// The pool allows liquidity providers to concentrate capital within specific price ranges,
// improving capital efficiency compared to full-range liquidity provision.
//
// Mathematical Model:
//   The pool uses the x*y=k constant product formula within each tick range,
//   with virtual reserves calculated based on the current price tick.
//
// Thread Safety:
//   Pool methods are safe for concurrent reads. Concurrent writes require
//   external synchronization.
//
// Precision:
//   All calculations use github.com/shopspring/decimal with full precision.
//   Rounding uses ROUND_HALF_UP strategy.
type Pool struct {
    poolID      string
    tokenA      *core.Token
    tokenB      *core.Token
    fee         constants.FeeAmount
    tickSpacing int
}

// NewPool creates a new concentrated liquidity pool.
//
// Parameters:
//   - poolID: Unique identifier for this pool instance
//   - tokenAAddress: Ethereum address of token A
//   - tokenADecimals: Number of decimals for token A (e.g., 6 for USDC)
//   - tokenBAddress: Ethereum address of token B
//   - tokenBDecimals: Number of decimals for token B (e.g., 18 for WETH)
//   - fee: Fee tier in basis points (500=0.05%, 3000=0.3%, 10000=1%)
//
// Returns an error if the poolID is empty or the fee tier is invalid.
func NewPool(
    poolID string,
    tokenAAddress common.Address,
    tokenADecimals uint,
    tokenBAddress common.Address,
    tokenBDecimals uint,
    fee constants.FeeAmount,
) (*Pool, error) {
    // ... implementation
}
```

## Testing Requirements

### Minimum Requirements

- **Test Coverage**: >80% for new code (aim for >85%)
- **Table-Driven Tests**: Use test tables for multiple cases
- **Edge Cases**: Test zero values, very large/small numbers, negative values
- **Error Paths**: Verify errors are returned correctly
- **Interface Compliance**: Verify your type implements interfaces correctly

### Test File Organization

```go
package yourpackage

import (
    "context"
    "testing"
    
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
    "github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// TestYourType_Method tests specific functionality
func TestYourType_Method(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr error
    }{
        {
            name: "basic case",
            // ... test data
        },
        {
            name: "edge case: zero input",
            // ... test data
        },
        {
            name: "error case: invalid input",
            // ... test data
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

// TestYourType_ImplementsInterface verifies interface compliance
func TestYourType_ImplementsInterface(t *testing.T) {
    var _ mechanisms.LiquidityPool = (*YourType)(nil)
}
```

### Validation Against References

If implementing a known protocol, validate against reference values:

```go
func TestYourPool_AgainstReferenceValues(t *testing.T) {
    pool := NewYourPool(...)
    
    // Use known state from protocol documentation or on-chain data
    params := mechanisms.PoolParams{
        ReserveA: primitives.MustAmount(primitives.MustDecimal("1000000")),
        ReserveB: primitives.MustAmount(primitives.MustDecimal("500")),
    }
    
    result, err := pool.Calculate(context.Background(), params)
    if err != nil {
        t.Fatalf("Calculate failed: %v", err)
    }
    
    // Validate with tolerance (e.g., 0.01%)
    expectedPrice := primitives.MustDecimal("2000")
    tolerance := expectedPrice.Mul(primitives.MustDecimal("0.0001"))
    
    diff := result.SpotPrice.Decimal().Sub(expectedPrice).Abs()
    if diff.GreaterThan(tolerance) {
        t.Errorf("Price %v outside tolerance of %v",
            result.SpotPrice.Decimal(), expectedPrice)
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests with race detector
go test ./... -race

# Run specific package tests
go test ./pkg/implementations/yourpackage/...

# Verbose output
go test -v ./...
```

### Coverage Requirements

Aim for >80% coverage, focusing on:
- ✅ All public methods
- ✅ Error handling paths
- ✅ Edge cases (zero, negative, very large/small values)
- ✅ Context cancellation (if applicable)

You can skip coverage for:
- ❌ Generated code
- ❌ Trivial getters/setters
- ❌ Unreachable error paths (e.g., infallible operations)

## Documentation Requirements

### Code Documentation

1. **Package-level docs** at the top of main file:
```go
// Package yourpackage implements [protocol] for go-crypto-quant-toolkit.
//
// [Brief description of what the package does]
//
// Usage Example:
//   // Code example here
package yourpackage
```

2. **Type documentation** for all exported types:
```go
// YourType implements mechanisms.Interface for [purpose].
//
// Detailed description including:
// - What problem it solves
// - Key features or algorithms
// - Thread safety guarantees
// - Precision/rounding behavior
type YourType struct { ... }
```

3. **Method documentation** for all exported methods:
```go
// MethodName does [what] and returns [what].
//
// Parameters explain each input.
// Returns section explains outputs and errors.
// Example section shows usage (optional but recommended).
func (t *YourType) MethodName(...) (..., error) { ... }
```

### Additional Documentation

For substantial contributions (new implementations), add a `README.md`:

```markdown
# Your Implementation

## Overview
Brief description of what this implements

## Features
- Feature 1
- Feature 2

## Usage
```go
// Code example
```

## Testing
How to run tests and validate

## References
- Links to protocol documentation
- Academic papers if applicable
```

### Update Existing Docs

If your contribution affects existing documentation:
- Update relevant sections
- Keep documentation accurate and up-to-date
- Fix typos or unclear explanations you find

## Pull Request Process

### Before Submitting

- [ ] Code follows Go style guidelines (`go fmt`, `go vet`)
- [ ] All tests pass (`go test ./...`)
- [ ] Test coverage >80% for new code
- [ ] Race detector passes (`go test -race ./...`)
- [ ] All public APIs have godoc comments
- [ ] Added tests for new functionality
- [ ] Updated documentation if needed
- [ ] Validated against reference values (if applicable)
- [ ] No breaking changes to existing interfaces (unless discussed)

### PR Description Template

```markdown
## Description
Brief summary of changes

## Type of Change
- [ ] New mechanism implementation
- [ ] Bug fix
- [ ] Documentation improvement
- [ ] Test enhancement
- [ ] Other (describe)

## Checklist
- [ ] Tests added/updated
- [ ] Documentation added/updated
- [ ] Code follows project style
- [ ] All tests pass
- [ ] >80% test coverage

## Related Issues
Fixes #123 (if applicable)

## Additional Context
Any other information reviewers should know
```

### Review Process

1. **Automated checks** run on all PRs (tests, linting)
2. **Maintainer review** - at least one maintainer will review
3. **Feedback addressed** - respond to review comments
4. **Approval** - maintainer approves after addressing feedback
5. **Merge** - maintainer merges the PR

### After Merge

- Your contribution will be included in the next release
- You'll be credited in release notes
- Consider writing a blog post or tutorial about your contribution!

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Assume good intentions
- No harassment, discrimination, or inappropriate behavior

### Communication

- **GitHub Issues**: Bug reports, feature requests, questions
- **Pull Requests**: Code contributions and discussions
- **Discussions**: General questions and ideas (if enabled)

### Best Practices for Contributors

1. **Start small**: First contribution? Fix docs, add tests, or improve examples
2. **Ask questions**: Not sure about something? Open an issue and ask!
3. **Share your use case**: Explain why you need a feature
4. **Be patient**: Maintainers may be busy; we'll respond when we can
5. **Learn from feedback**: Code reviews are learning opportunities

### Getting Help

Stuck or need help?

1. Check [EXTENDING.md](docs/EXTENDING.md) for mechanism implementation guide
2. Review [examples/](examples/) for working examples
3. Search existing issues for similar problems
4. Open a new issue with your question

## Recognition

All contributors will be recognized in:
- Release notes
- Project documentation
- GitHub contributors list

Significant contributions may be featured in:
- Blog posts or case studies
- Conference presentations
- Community highlights

## Questions?

If you have questions about contributing, please:

1. Check this guide and other documentation
2. Search existing issues
3. Open a new issue with your question

Thank you for contributing to go-crypto-quant-toolkit! 🚀
