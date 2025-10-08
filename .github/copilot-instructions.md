# Development Guidelines

## Documentation and Context

Always use Context7 to retrieve current documentation when working with frameworks, libraries, or APIs rather than relying on training data. This applies to:
- Answering questions about APIs, frameworks, or libraries
- Implementing integrations with external APIs or services
- Writing code that uses third-party packages or SDKs
- Debugging or updating existing integrations

Automatically invoke the Context7 MCP tools without being asked to ensure you're using the most up-to-date documentation and best practices.

## Development Standards

### Go Best Practices
- Prioritize interface-based design with small, focused interfaces following Go idioms (accept interfaces, return structs)
- Use the standard library first; avoid external dependencies unless absolutely necessary for the framework's extensibility goals
- Leverage context.Context for cancellation and timeouts in all long-running or I/O operations
- Keep exported APIs minimal and stable; prefer internal packages for implementation details
- Use meaningful error wrapping with fmt.Errorf and %w to preserve error chains for debugging
- Avoid premature optimization; profile before optimizing, especially in backtesting hot paths

### Financial Computation Patterns
- Always use github.com/shopspring/decimal for monetary values and price calculations to prevent floating-point precision errors
- Never use float64 for financial amounts, prices, or percentages; use primitives.Decimal wrappers consistently
- Validate all financial calculations against known test cases with explicit tolerance levels (e.g., 0.01% for AMM math)
- Document precision requirements and rounding behavior explicitly in godoc comments
- Handle division by zero and overflow explicitly in all arithmetic operations

### Interface Design Guidelines
- Define contracts (interfaces) before implementations; framework should never depend on concrete types
- Keep interface methods minimal (1-3 methods ideal); compose larger behaviors from small interfaces
- Use embedded interfaces to build complex contracts from simpler ones (e.g., MarketMechanism in specialized interfaces)
- Document interface contracts thoroughly including expected behaviors, error conditions, and thread-safety guarantees
- Write property-based tests for interface contracts that any implementation must satisfy
- Never break interface compatibility once published; add new interfaces instead of modifying existing ones

### Code Quality Standards
- Run golangci-lint with strict settings before committing; configure with --fast-only flag for IDE integration
- Enable key linters: govet, staticcheck, errcheck, gosec, gofmt, goimports for standard code quality
- Use go vet for catching common mistakes; address all vet warnings before merging
- Write table-driven tests with clear test names describing the scenario being tested
- Ensure all exported functions, types, and methods have godoc comments following Go documentation conventions
- Maintain >80% test coverage for core framework packages (primitives, mechanisms, strategy)

### Testing Patterns
- Use property-based testing (github.com/leanovate/gopter or similar) for validating interface contracts and invariants
- Write integration tests demonstrating that new mechanism implementations work without framework changes
- Separate unit tests (single package) from integration tests (cross-package) using build tags if needed
- Test error paths explicitly; ensure errors are properly wrapped and contain actionable context
- Use golden files or known-correct values from production systems (e.g., Uniswap V3) for reference implementation validation

### Project Conventions
- Follow standard Go project layout: cmd/, internal/, pkg/ where pkg/ contains importable framework code
- Name packages by what they provide, not what they contain (strategy not strategies, mechanisms not interfaces)
- Keep package dependencies acyclic; use interfaces to invert dependencies when needed
- Use conventional Go import grouping: stdlib, external, internal (separated by blank lines)
- Prefix internal implementation packages with impl or internal to signal they're not part of the public API
- Use consistent naming: Position for any tradeable thing, Action for portfolio modifications, Snapshot for point-in-time state
