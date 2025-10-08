# Development Guidelines

## Context7 Usage
Auto-invoke Context7 MCP tools for up-to-date docs when working with external APIs, frameworks, or libraries.

## Go Standards
- Interface-first design: small interfaces (1-3 methods), accept interfaces/return structs
- Stdlib first; minimize external dependencies
- context.Context for cancellation/timeouts in I/O operations
- Error wrapping with fmt.Errorf + %w
- Minimal exported APIs; document with godoc

## Financial Math
- **NEVER float64** for money/prices; always use github.com/shopspring/decimal via primitives.Decimal
- Validate against known values with explicit tolerance (e.g., 0.01% for AMM math)
- Document precision/rounding; handle division by zero

## Interface Design
- Define contracts before implementations
- Framework never depends on concrete types
- Embedded interfaces for composition
- Document behaviors, errors, thread-safety
- Property-based tests for interface contracts
- Never break compatibility; add new interfaces instead

## Code Quality
- Run `go vet ./...` and `go fmt ./...` before committing
- Table-driven tests with descriptive names
- >80% coverage for core packages (primitives, mechanisms, strategy)
- Test error paths explicitly
- Validate implementations against reference systems (e.g., Uniswap V3)

## Project Layout (golang-standards/project-layout)
- **Structure:** `pkg/primitives/`, `pkg/mechanisms/`, `pkg/strategy/`, `pkg/implementations/`, `pkg/backtest/`
- **Imports:** Full path with pkg: `github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives`
- **Naming:** packages by purpose (strategy not strategies); Position/Action/Snapshot conventions
- Acyclic dependencies; stdlib/external/internal import grouping
