# Complete Delta-Neutral Automation Process

## Phase 1: Pre-Position Analysis (Every 5 Minutes)

### Step 1: Data Collection
```
1.1 Fetch current prices
    - ETH price from Chainlink oracles
    - BTC price from multiple sources
    - Cross-reference Binance, Coinbase, FalconX
    - Calculate TWAP over last 60 seconds
    - Identify any price discrepancies >0.5%

1.2 Scan lending rates across protocols
    - Aave V3 (Ethereum): USDC supply APY, ETH borrow APY
    - Aave V3 (Arbitrum): USDC supply APY, ETH borrow APY  
    - Aave V3 (Optimism): USDC supply APY, ETH borrow APY
    - Compound V3: USDC supply APY, ETH borrow APY
    - Morpho: Optimized Aave/Compound rates

1.3 Collect rewards/incentive data
    - ARB rewards on Arbitrum protocols
    - OP rewards on Optimism protocols
    - Native token emissions (RDNT, COMP, etc.)
    - Calculate dollar value of rewards
    - Factor in vesting schedules

1.4 Monitor perpetual funding rates
    CEX:
    - Coinbase International funding (via Prime)
    - Binance funding (if accessible)
    - FTX funding (via FalconX if available)
    
    DEX:
    - GMX (Arbitrum): ETH-USD funding
    - GMX (Avalanche): ETH-USD funding
    - Gains Network: ETH-USD funding
    - MUX Protocol: ETH-USD funding
    - Vertex Protocol: ETH-USD funding
    - HMX: ETH-USD funding
    - Hyperliquid: ETH-USD funding
    - dYdX v4: ETH-USD funding

1.5 Gas price monitoring
    - Current gas on each chain
    - Historical gas patterns (cheaper at 3 AM UTC)
    - Estimate transaction costs for rebalancing
    - L2 costs vs L1 costs
```

### Step 2: Opportunity Identification
```
2.1 Calculate net yields for each combination
    Example calculation:
    
    Aave Arbitrum + GMX:
    - Supply APY: 4.5%
    - ARB Rewards: 2.1%
    - Borrow APY: -2.8%
    - GMX Funding: -0.01% * 3 = -0.03% daily = -10.95% APY
    - Net Yield: 4.5% + 2.1% - 2.8% + 10.95% = 14.75%
    - Minus gas costs: ~$50/week = -0.25% APY
    - Final: 14.5% APY

2.2 Rank opportunities by risk-adjusted return
    - Calculate Sharpe ratio for each
    - Factor in protocol TVL (safety)
    - Consider liquidity for exit
    - Weight by historical reliability

2.3 Check position constraints
    - Maximum 30% in any protocol
    - Minimum $50k position size (gas efficiency)
    - Maximum 3x leverage
    - Correlation limits
```

## Phase 2: Position Opening (One-Time)

### Step 3: Capital Deployment
```
3.1 Determine optimal allocation
    Input: $1,000,000 USDC
    
    Allocation decision:
    - 50% ($500k) to Aave Arbitrum (highest TVL, safest)
    - 30% ($300k) to Compound (diversification)
    - 10% ($100k) to Morpho (optimized rates)
    - 10% ($100k) reserve for opportunities

3.2 Bridge preparation (if needed)
    - Check bridge liquidity (Stargate, Across, Native)
    - Compare bridge costs vs holding on each chain
    - Execute bridges during low gas periods
    - Verify receipt on destination chain

3.3 Execute supply transactions
    For each protocol via FordeFi:
    
    Aave Arbitrum:
    a. Approve USDC spending (if first time)
    b. Call pool.supply(USDC, 400000e6, user, 0)
    c. Verify aUSDC receipt tokens received
    d. Enable USDC as collateral if not already

3.4 Wait for confirmations
    - Arbitrum: Wait 1 block (~250ms)
    - Verify collateral is enabled
    - Confirm supply APY is being earned
```

### Step 4: Borrow Execution
```
4.1 Calculate safe borrow amounts
    For $400k USDC in Aave at 85% max LTV:
    - Maximum borrow: $340k
    - Safe borrow (50% LTV): $200k
    - Convert to ETH: $200k / $2000 = 100 ETH

4.2 Execute borrow transactions
    Aave Arbitrum:
    a. Call pool.borrow(WETH, 100e18, 2, 0, user)
    b. Verify WETH received in wallet
    c. Check health factor > 1.5

4.3 Verify positions
    - Total borrowed: 100 ETH ($200k)
    - Health factor: 2.0
    - Liquidation price: $1,176 (41% drop)
```

### Step 5: Spot Conversion
```
5.1 Choose execution venue
    Options for selling 175 ETH:
    
    CEX route (via FalconX):
    - Get quote from FalconX
    - Typically 0.01% spread
    - Instant settlement
    
    DEX route (via FordeFi):
    - Check Uniswap V3 liquidity
    - Check Curve liquidity
    - Use 1inch router for best path

5.2 Execute swap
    If FalconX better:
    a. Transfer ETH to FalconX settlement address
    b. Execute market sell
    c. Receive USDC in Coinbase Prime
    
    If DEX better:
    a. Approve WETH spending to router
    b. Execute swap via aggregator
    c. Receive USDC in wallet

5.3 Verify execution
    - Confirm USDC received: $350k
    - Calculate slippage: Target <0.1%
    - Log execution price for P&L
```

### Step 6: Perpetual Short
```
6.1 Select venues for shorting
    Split across venues for best rates:
    - 100 ETH short on GMX (good funding)
    - 50 ETH short on Coinbase (liquidity)
    - 25 ETH short on Gains (backup)

6.2 Execute GMX short (via FordeFi)
    a. Approve USDC for GMX vault
    b. Create position:
       - Collateral: $40k USDC
       - Size: $200k (100 ETH)
       - Leverage: 5x
    c. Verify position opened
    d. Check liquidation price: $2,400

6.3 Execute Coinbase short
    a. Transfer USDC collateral to Prime
    b. Open perpetual short via API
    c. Set position size: 50 ETH
    d. Verify margin requirements

6.4 Position verification
    - Total short: 175 ETH
    - Matches borrow amount exactly
    - Delta neutral achieved: ±0.01 ETH
```

## Phase 3: Continuous Monitoring (Every 30 Seconds)

### Step 7: Health Monitoring
```
7.1 Check lending health factors
    Every 30 seconds:
    - Query each protocol's health factor
    - If health < 1.3: ALERT
    - If health < 1.2: EMERGENCY DELEVERAGE
    - If health < 1.1: IMMEDIATE CLOSE ALL

7.2 Monitor liquidation prices
    Track distance to liquidation:
    
    Aave position:
    - Current ETH: $2,000
    - Liquidation at: $1,142
    - Buffer: 43% ($858)
    
    GMX short:
    - Current ETH: $2,000
    - Liquidation at: $2,400
    - Buffer: 20% ($400)

7.3 Delta calculation
    Current positions:
    - Long exposure: 0 ETH (sold spot)
    - Borrow exposure: -175 ETH
    - Short exposure: +175 ETH
    - Net delta: 0 ETH ✓

7.4 Funding rate monitoring
    GMX funding: -0.01% (earning)
    Coinbase funding: -0.008% (earning)
    
    If funding goes positive:
    - Calculate cost vs gas to close
    - If cost > $100/day, close position
    - Reopen on different venue
```

### Step 8: P&L Tracking
```
8.1 Real-time P&L calculation
    Revenue streams (per day):
    + USDC supply interest: $49.32
    + ARB rewards: $23.01
    + ETH borrow rewards: $5.75
    + Funding earned: $52.05
    = Gross revenue: $130.13/day

    Costs (per day):
    - ETH borrow interest: $26.85
    - Gas for monitoring: $5.00
    - Rebalancing gas: $10.00
    = Total costs: $41.85/day
    
    Net profit: $88.28/day
    APY: 11.2%

8.2 Track cumulative metrics
    - Total earned: $2,648.40 (30 days)
    - Gas spent: $450
    - Rebalances executed: 15
    - Near liquidations: 0
```

## Phase 4: Rebalancing Logic (Every 15 Minutes)

### Step 9: Rebalancing Triggers
```
9.1 Delta drift detection
    If abs(delta) > 2 ETH:
    - Calculate cost to rebalance
    - If cost < $50, execute immediately
    - If cost > $50, wait for 5 ETH drift

9.2 Rate optimization triggers
    If new_venue_APY > current_APY + 2%:
    - Calculate migration costs
    - If payback < 30 days, migrate
    - Execute during low gas window

9.3 Risk-based triggers
    If health factor < 1.5:
    - Reduce position by 20%
    - Add more collateral
    - Lower leverage

9.4 Funding rate triggers
    If funding > 0.01% for 3 periods:
    - Close perpetual position
    - Find negative funding venue
    - Reopen position
```

### Step 10: Rebalancing Execution
```
10.1 Small rebalance (delta drift)
    Example: ETH pumped to $2,100
    - Borrow value increased: 175 ETH = $367.5k
    - Short value: Still $350k at entry
    - Delta: -8.33 ETH ($17.5k exposed)
    
    Fix:
    a. Increase short by 8.33 ETH on GMX
    b. Or reduce borrow by 8.33 ETH
    c. Choose based on gas costs

10.2 Venue migration (better rates)
    Moving from GMX to Hyperliquid:
    
    a. Open new short on Hyperliquid first
    b. Close GMX position
    c. Never be without hedge
    d. Accept 5 minutes of double exposure

10.3 Emergency deleveraging
    If health factor critical:
    
    Priority order:
    1. Close lowest-yielding positions first
    2. Use flashloans if available
    3. Market sell if necessary
    4. Accept slippage to save position
```

## Phase 5: Daily Operations

### Step 11: Rewards Management
```
11.1 Claim rewards
    Daily at 00:00 UTC:
    - Claim ARB from Aave
    - Claim GMX rewards
    - Claim esGMX and stake

11.2 Rewards optimization
    If rewards > $1000:
    - Sell 50% for USDC
    - Compound into position
    - Keep 50% for accumulation

11.3 Vest management
    - Track vesting schedules
    - Auto-claim when vested
    - Reinvest or distribute
```

### Step 12: Reporting
```
12.1 Generate daily report
    - Starting NAV
    - Positions opened/closed
    - P&L breakdown
    - Gas costs
    - Health factors
    - Funding rates earned
    - Rewards claimed
    - Ending NAV

12.2 Risk report
    - Max drawdown
    - Liquidation distances
    - Correlation analysis
    - Stress test results

12.3 Send notifications
    - Telegram/Discord alerts
    - Email daily summary
    - Dashboard updates
```

## Phase 6: Edge Cases & Safety

### Step 13: Handle Failures
```
13.1 Protocol failures
    If Aave is down:
    - Stop new positions
    - Monitor existing only
    - Use backup protocols

13.2 Oracle failures
    If price feed diverges >2%:
    - Pause all operations
    - Use backup oracles
    - Manual intervention required

13.3 Gas spike handling
    If gas > 1000 gwei:
    - Only emergency transactions
    - Queue non-urgent rebalances
    - Execute when gas normalizes

13.4 Depeg handling
    If USDC depegs >1%:
    - Pause new positions
    - Consider closing to USDT
    - Wait for resolution
```

### Step 14: Circuit Breakers
```
14.1 Loss limits
    If daily loss > 2%:
    - Stop all operations
    - Close risky positions
    - Manual review required

14.2 Exposure limits
    If any protocol > 40% of portfolio:
    - Prevent new positions
    - Rebalance when possible

14.3 Correlation breaks
    If perp/spot spread > 1%:
    - Alert immediately
    - Close if spread > 2%
    - Investigate cause
```

## Complete Automation Flow

```python
while True:
    # Every 30 seconds
    health = check_all_health_factors()
    if health.critical:
        emergency_close_positions()
    
    # Every 5 minutes
    opportunities = scan_all_rates()
    if better_opportunity_exists():
        calculate_migration_cost()
        if profitable:
            migrate_position()
    
    # Every 15 minutes
    delta = calculate_portfolio_delta()
    if abs(delta) > threshold:
        rebalance_to_neutral()
    
    # Every hour
    funding = check_funding_rates()
    if funding.turned_positive:
        rotate_perpetual_venues()
    
    # Daily
    claim_all_rewards()
    generate_reports()
    compound_profits()
    
    # Continuous
    log_all_actions()
    update_dashboards()
    send_alerts_if_needed()
```

## Time Estimates for Automated Actions

- **Health check**: 50ms per protocol (parallelized)
- **Rate scanning**: 2 seconds (all venues)
- **Rebalancing decision**: 100ms
- **Transaction execution**: 1-12 seconds (chain dependent)
- **Full cycle**: <30 seconds

This replaces what would take a human:
- 2 hours/day of monitoring
- 30 minutes per rebalance
- High stress from liquidation risk
- Missed opportunities while sleeping