# Project Brief: go-crypto-quant-toolkit

## Vision
An open-source Go framework providing composable primitives and extensible interfaces for crypto trading strategy development. Enables researchers and traders to rapidly prototype and backtest strategies across CEX, DEX, and hybrid venues by composing pluggable market mechanisms without framework refactoring.

## User Personas
### Primary User: Quantitative Researcher
- **Role:** Researcher or trader developing crypto strategies in Go, validating strategy performance before capital deployment
- **Needs:** Extensible framework to compose strategies from pluggable components, ability to add new market mechanisms without framework changes, confidence that abstractions won't limit strategy complexity
- **Pain Points:** Existing libraries are either protocol-locked or require framework modifications for new strategy types, can't easily mix CEX order book logic with DEX AMM logic, adding new venue types breaks existing code
- **Success:** Implements complex multi-venue strategy using framework interfaces, adds custom market mechanism (e.g., batch auction) without touching framework code, backtests with confidence in extensibility

### Secondary User: Production Trading Firm
- **Role:** Trading firm building automated crypto strategies requiring extensible, maintainable foundation for Go-based execution systems
- **Needs:** Type-safe framework that won't break when adding new strategies, clear extension points for proprietary mechanisms, ability to contribute implementations back to open-source
- **Pain Points:** Rigid frameworks require forking and maintaining custom versions, unclear how to extend for proprietary venue types, fear of technical debt from framework limitations
- **Success:** Adopts framework as foundation for all strategies, extends with proprietary implementations, contributes generic improvements upstream

## Core Requirements
- [MVP] The system should provide clear interfaces for market mechanisms (liquidity pools, derivatives, order books) that strategies compose without implementation coupling
- [MVP] The system should enable adding new market mechanism implementations without modifying framework code or existing strategies
- [MVP] The system should provide 2-3 reference implementations (concentrated liquidity, options, perpetuals) demonstrating interface usage
- [MVP] The system should define composable strategy framework where strategies coordinate multiple mechanisms across venues
- [MVP] The system should provide event-driven backtesting that works regardless of strategy complexity or mechanism types
- [Post MVP] The system should accumulate community-contributed market mechanism implementations (order books, AMM variants, exotic derivatives)
- [Post MVP] The system should provide optimization and risk analytics as optional packages built on framework primitives
- [Post MVP] The system should support advanced features (MEV tooling, ML hooks, ZK integration) through clear extension points

## Success Metrics
1. Researcher adds custom market mechanism (e.g., batch auction) by implementing interface without modifying framework code
2. Reference implementations demonstrate composability: multi-venue strategy combining 3+ mechanism types in <500 lines
3. Three distinct mechanism implementations contributed by community within 6 months (validates extensibility)
4. Framework supports strategies from market making list without architectural changes (order books, MEV, flash loans, etc.)
5. Repository achieves 100+ GitHub stars within 12 months, with forks extending (not replacing) framework