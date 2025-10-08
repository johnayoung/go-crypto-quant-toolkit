# Complete ETH-Denominated Farming Automation Process

## Strategy Overview
**Objective**: Maximize ETH-denominated yields through fixed-rate lending, liquidity provision rewards, and strategic leverage while maintaining delta neutrality.
**Target Return**: 15-25% APY in ETH terms
**Risk Profile**: Medium-High (complex position management, multiple protocol risks)

## Phase 1: Pre-Position Analysis (Every 5 Minutes)

### Step 1: Market Data Collection
```
1.1 Fetch current pricing data
    Core assets:
    - ETH spot price from multiple sources
    - stETH/ETH ratio monitoring
    - PT-stETH discount/premium to maturity
    - YT-stETH implied yield
    - wstETH conversion rate
    
    Oracle verification:
    - Chainlink primary feed
    - Binance/Coinbase backup
    - Pendle's internal TWAP oracle
    - Calculate 5-minute VWAP
    - Alert if divergence >0.5%

1.2 Scan Pendle fixed yield opportunities
    PT (Principal Token) markets:
    - PT-stETH maturity dates available
    - Fixed APY for each maturity
    - Liquidity depth per market
    - Days to maturity
    - Premium/discount to underlying
    
    YT (Yield Token) markets:
    - Implied APY calculations
    - Volume and liquidity
    - Historical yield variance
    
    LP opportunities:
    - PT/SY pool APYs
    - PENDLE incentives
    - Vesting schedules
    - Impermanent loss estimates

1.3 Monitor lending market parameters
    Primary venues:
    - Aave V3: ETH/stETH collateral factors
    - Compound V3: ETH supply/borrow rates
    - Morpho: P2P matching rates
    - Euler V2: Risk-adjusted rates
    - Silo: Isolated pool opportunities
    
    Key metrics:
    - Maximum LTV ratios
    - Liquidation thresholds
    - Available liquidity
    - Utilization rates
    - Supply caps remaining

1.4 GMX reward mechanics analysis
    GLP composition monitoring:
    - Current ETH weight in index
    - BTC, USDC, DAI allocations
    - Target vs actual weights
    - Rebalancing incentives
    
    Reward streams:
    - esGMX emission rate
    - ETH fee distribution (30% of fees)
    - Multiplier points accrual
    - Vesting schedules
    - Boost mechanisms

1.5 Perpetual funding rate scanning
    CEX venues:
    - Binance ETH-USDT funding
    - Coinbase International rates
    - Bybit ETH-USD funding
    - Deribit options-adjusted funding
    
    DEX venues:
    - GMX V2 funding mechanism
    - Synthetix Perps V3
    - Vertex Protocol rates
    - Level Finance funding
    - HMX funding rates
    - Gains Network rates

1.6 Gas and execution cost analysis
    Transaction cost estimation:
    - Pendle position entry (complex)
    - Multiple lending operations
    - GMX GLP minting costs
    - Perpetual position management
    - Claim and compound costs
    - Emergency exit scenarios
```

### Step 2: Yield Optimization Matrix
```
2.1 Calculate Pendle fixed yields
    PT-stETH analysis example (Dec 2025 maturity):
    - Current PT price: 0.92 ETH
    - Maturity value: 1.0 ETH
    - Days to maturity: 365
    - Fixed yield: (1/0.92 - 1) * 365/365 = 8.7% APY
    - Plus underlying stETH yield: 3.5%
    - Total fixed return: 12.2% APY
    
    Optimal maturity selection:
    - Short-term (30-90 days): Lower yield, less risk
    - Medium-term (180-365 days): Balanced yield/risk
    - Long-term (365+ days): Higher yield, duration risk
    
2.2 Calculate leveraged returns
    Starting with 1000 ETH capital:
    
    Pendle LP position (1000 ETH):
    - Supply to PT-stETH/SY-stETH pool
    - Fixed yield: 8.7% on PT side
    - Variable yield: 3.5% on SY side
    - LP fees: 0.3% of volume
    - PENDLE rewards: 4.2% APY
    - Base position return: ~16.7% APY
    
    Leverage addition (50% LTV):
    - Borrow 500 ETH against position
    - Deploy to GMX GLP
    - ETH rewards from fees: 15-25% APY
    - esGMX rewards: 10-15% APY
    - Multiplier points: 100% APR on staked
    
2.3 Net position calculations
    Revenue streams:
    + Pendle fixed yield: 87 ETH/year
    + Pendle LP fees: 3 ETH/year
    + PENDLE rewards: 42 ETH value/year
    + GMX ETH distribution: 100 ETH/year
    + esGMX value: 60 ETH equivalent/year
    
    Costs:
    - Borrow cost (3%): -15 ETH/year
    - Funding (if positive): Variable
    - Gas and operations: -5 ETH/year
    
    Net return: ~272 ETH/year (27.2% APY)
    
2.4 Risk-adjusted scoring
    Position ranking factors:
    - Sharpe ratio calculation
    - Maximum drawdown potential
    - Liquidity for exit (can unwind $10M?)
    - Smart contract risk score
    - Correlation to ETH price
    - Complexity penalty
```

## Phase 2: Position Construction (One-Time Setup)

### Step 3: Pendle PT Position Entry
```
3.1 Pre-entry validation
    Check before deploying 1000 ETH:
    - PT market liquidity (>$20M depth)
    - Current discount to maturity
    - Days until next maturity
    - Gas costs vs expected profit
    - Slippage estimates for size

3.2 Convert ETH to position assets
    Path optimization:
    
    Option A - Direct PT purchase:
    a. Swap ETH to stETH via Curve
    b. Wrap stETH to wstETH if needed
    c. Swap wstETH to PT-stETH on Pendle
    d. Calculate effective fixed rate locked
    
    Option B - LP position entry:
    a. Split ETH 50/50
    b. Convert half to PT-stETH
    c. Keep half as SY-stETH
    d. Add liquidity to PT/SY pool
    e. Receive LP tokens

3.3 Execute Pendle LP provision
    Via FordeFi or direct interaction:
    
    a. Approve token spending to Pendle router
    b. Calculate optimal LP ratio
    c. Call addLiquidityDual with parameters:
       - PT token address
       - SY token address
       - Amount of each token
       - Minimum LP tokens expected
       - Deadline for execution
    d. Receive and verify LP tokens
    e. Check position value and APY

3.4 Enable LP staking (if available)
    For additional rewards:
    a. Stake LP tokens in Pendle gauge
    b. Enable PENDLE reward streaming
    c. Set up vePENDLE boost if holding
    d. Configure auto-compound settings
```

### Step 4: Collateralized Borrowing
```
4.1 Select optimal lending venue
    Evaluation criteria:
    - Accept Pendle LP as collateral?
    - If not, use stETH/ETH directly
    - Best LTV ratios available
    - Lowest borrow rates
    - Liquidation parameters
    
    Likely venues:
    - Aave V3 (stETH collateral)
    - Compound V3 (ETH markets)
    - Morpho (optimized rates)
    - Silo (isolated pools)

4.2 Collateral deployment strategy
    If Pendle LP not accepted:
    a. Keep 500 ETH liquid from start
    b. Use remaining 500 ETH for Pendle
    c. Supply liquid ETH as collateral
    d. Borrow against it
    
    If LP tokens accepted:
    a. Supply LP tokens as collateral
    b. Enable as collateral
    c. Check borrowing power
    d. Proceed with borrowing

4.3 Execute borrowing
    Target 50% LTV for safety:
    
    a. Calculate maximum safe borrow
    b. Call borrow function with:
       - Asset: ETH/WETH
       - Amount: 500 ETH
       - Interest rate mode: Variable
       - Referral code if applicable
    c. Verify health factor >2.0
    d. Confirm ETH received

4.4 Monitor initial health
    Post-borrow checks:
    - Current LTV: Should be ~50%
    - Liquidation LTV: Typically 80-85%
    - Health factor: Target >2.0
    - Available to borrow: Remaining capacity
```

### Step 5: GMX GLP Deployment
```
5.1 Analyze GLP composition
    Current index weights:
    - ETH: 30-35% target
    - BTC: 30-35% target
    - USDC: 15-20% target
    - DAI: 5-10% target
    - Other stables: 5-10%
    
    Entry optimization:
    - If ETH underweight: Bonus on entry
    - If ETH overweight: Penalty on entry
    - Calculate net after fees

5.2 Mint GLP tokens
    Process with 500 borrowed ETH:
    
    a. Bridge ETH to Arbitrum if needed
    b. Approve ETH to GLP manager
    c. Call mintAndStakeGlp with:
       - Token: ETH
       - Amount: 500 ETH
       - Min USDG: Slippage protection
       - Min GLP: Expected output
    d. Receive staked GLP automatically
    e. Begin earning immediately

5.3 Configure GMX rewards
    Reward streams to enable:
    
    esGMX rewards:
    - Auto-staking enabled
    - Vesting period: 365 days
    - Can compound for higher APY
    
    ETH distributions:
    - Paid every second
    - Auto-claimed or manual
    - Can compound to GLP
    
    Multiplier points:
    - Accrue at 100% APR
    - Boost future rewards
    - Never expire

5.4 Track GLP position metrics
    Important values:
    - GLP tokens received
    - Entry price per GLP
    - Current index composition
    - Fee APR (historical 15-25%)
    - esGMX APR (10-15%)
    - Total position value in ETH
```

### Step 6: Delta Hedge Implementation
```
6.1 Calculate total exposure
    Position breakdown:
    - Pendle PT/LP: 1000 ETH long
    - Borrowed: -500 ETH short
    - GLP ETH portion: ~150 ETH long (30% of 500)
    - GLP BTC portion: ~150 ETH equivalent
    - Net long exposure: 800 ETH
    
    Hedge requirement:
    - Need to short 800 ETH in perps
    - Maintain perfect neutrality
    - Account for GLP rebalancing

6.2 Distribute perpetual shorts
    Multi-venue approach for optimization:
    
    40% on GMX V2 (320 ETH):
    - Same platform as GLP
    - Good funding rates
    - Capital efficiency
    
    30% on Synthetix (240 ETH):
    - Optimism deployment
    - SNX rewards possible
    - Deep liquidity
    
    30% on Vertex (240 ETH):
    - Cross-margin efficiency
    - Low fees
    - Good funding historically

6.3 Execute GMX V2 perpetual short
    On Arbitrum:
    
    a. Allocate USDC collateral (64k for 320 ETH)
    b. Approve USDC to GMX router
    c. Create short position:
       - Market: ETH/USD
       - Size: 320 ETH
       - Collateral: 64,000 USDC
       - Leverage: 10x
       - Acceptable price range: ±0.5%
    d. Verify position opened
    e. Check liquidation price buffer

6.4 Execute Synthetix short
    On Optimism:
    
    a. Bridge USDC to Optimism
    b. Deposit sUSD collateral
    c. Open perp position:
       - Market: ETH-PERP
       - Size: -240 ETH
       - Margin: 48,000 sUSD
    d. Monitor funding rate
    e. Set up alerts

6.5 Execute Vertex short
    Cross-margin setup:
    
    a. Deposit USDC to Vertex
    b. Configure cross-margin mode
    c. Open ETH perpetual short:
       - Size: 240 ETH
       - Using shared collateral pool
    d. Verify margin requirements
    e. Enable auto-liquidation protection

6.6 Verify complete hedge
    Final position check:
    - Total long: 800 ETH
    - Total short: 800 ETH
    - Net delta: 0 ETH ±5 ETH tolerance
    - All positions confirmed
```

## Phase 3: Continuous Monitoring (Every 30 Seconds)

### Step 7: Position Health Monitoring
```
7.1 Lending health tracking
    Critical metrics every 30 seconds:
    
    Monitor parameters:
    - Current collateral value
    - Outstanding debt value
    - Health factor calculation
    - Distance to liquidation
    - Available borrowing capacity
    
    Alert thresholds:
    - Health <1.5: Warning notification
    - Health <1.3: Prepare deleveraging
    - Health <1.15: Execute deleverage
    - Health <1.1: Emergency exit all

7.2 Pendle position monitoring
    PT-specific tracking:
    
    Market metrics:
    - PT price vs underlying
    - Days to maturity
    - Implied yield changes
    - LP pool balance ratio
    - Liquidity depth
    
    Risk indicators:
    - If PT discount >15%: Investigate
    - If liquidity drops 50%: Alert
    - If yield spikes >5%: Rebalance

7.3 GLP composition tracking
    Monitor index weights:
    
    Rebalancing detection:
    - ETH weight changes >5%
    - BTC correlation shifts
    - Stablecoin fluctuations
    - Fee tier changes
    
    Adjust hedge if:
    - ETH allocation changes >10%
    - Total value drift >5%
    - Rewards rate changes significantly

7.4 Perpetual position monitoring
    For each venue track:
    
    Position health:
    - Margin ratio maintenance
    - Funding payment schedule
    - Liquidation distance
    - Unrealized PnL
    
    Market conditions:
    - Funding rate direction
    - Open interest changes
    - Basis spread to spot
    - Volume and liquidity
```

### Step 8: Yield & P&L Tracking
```
8.1 Real-time yield calculation
    Revenue streams (daily on 1000 ETH base):
    
    Fixed income:
    + Pendle fixed yield: 0.238 ETH/day
    + Pendle LP fees: 0.008 ETH/day
    + PENDLE rewards: 0.115 ETH value/day
    
    Variable income:
    + GMX ETH distribution: 0.274 ETH/day
    + esGMX rewards: 0.164 ETH value/day
    + Multiplier points value: 0.137 ETH/day
    
    Funding income/cost:
    + Perpetual funding: ±0.05 ETH/day
    
    Total gross: 0.936-0.986 ETH/day

8.2 Cost tracking
    Daily expenses:
    - ETH borrow interest: -0.041 ETH/day
    - Gas for monitoring: -0.003 ETH/day
    - Rebalancing costs: -0.005 ETH/day
    - Slippage allowance: -0.002 ETH/day
    
    Total costs: -0.051 ETH/day
    
    Net profit: ~0.885 ETH/day
    Effective APY: 32.3% on 1000 ETH

8.3 Performance metrics
    Key indicators:
    - Running APY (7/30/90 day)
    - Sharpe ratio tracking
    - Maximum drawdown
    - Win rate (profitable days)
    - Correlation to ETH price
    - Gas efficiency ratio

8.4 Reward accumulation tracking
    Claimable balances:
    - PENDLE tokens earned
    - esGMX accumulated
    - ETH distributions pending
    - Multiplier points balance
    - Vesting schedules active
```

## Phase 4: Dynamic Rebalancing (Every 15 Minutes)

### Step 9: Rebalancing Triggers
```
9.1 Delta drift management
    Tolerance: ±20 ETH exposure
    
    Causes of drift:
    - GLP composition changes
    - Liquidations/additions
    - Price impact on collateral
    - Funding payments
    
    If delta >20 ETH:
    - Calculate adjustment size
    - Identify cheapest venue
    - Execute rehedge

9.2 Yield optimization triggers
    Monitor for opportunities:
    
    If new PT market >2% better:
    - Calculate switching costs
    - Factor in days to maturity
    - Execute if payback <60 days
    
    If GLP fees drop <10% APY:
    - Consider alternative venues
    - Evaluate GMX V2 pools
    - Review other reward tokens

9.3 Health factor management
    Maintain 1.5-2.0 range:
    
    If health <1.5:
    - Reduce borrow by 10%
    - Or add collateral
    - Adjust perpetual hedge
    
    If health >2.5:
    - Opportunity to increase leverage
    - Borrow additional ETH
    - Deploy to highest yield

9.4 Funding rate optimization
    Monitor all perpetual venues:
    
    If funding >0.02% for 8 hours:
    - Rotate to negative funding venue
    - Close highest cost position first
    - Reopen on better platform
    
    If extreme negative funding <-0.05%:
    - Opportunity to increase shorts
    - Earn additional funding yield
```

### Step 10: Rebalancing Execution
```
10.1 PT maturity management
    As maturity approaches:
    
    30 days before maturity:
    - Begin monitoring new markets
    - Calculate roll costs
    - Plan transition strategy
    
    7 days before:
    - Start unwinding position
    - Move to next maturity
    - Maintain continuous exposure

10.2 GLP rebalancing response
    When index weights change:
    
    If ETH weight increases:
    - Need more perpetual shorts
    - Calculate additional size
    - Execute across venues
    
    If ETH weight decreases:
    - Reduce perpetual shorts
    - Free up collateral
    - Maintain delta neutrality

10.3 Leverage adjustments
    Increasing leverage:
    
    a. Borrow additional 100 ETH
    b. Add to GLP position
    c. Mint more GLP tokens
    d. Increase perpetual hedge by 30 ETH
    e. Verify health factor >1.5
    
    Decreasing leverage:
    
    a. Sell portion of GLP
    b. Receive ETH back
    c. Repay borrowed amount
    d. Reduce perpetual positions
    e. Confirm improved health

10.4 Emergency deleveraging
    If critical event detected:
    
    Priority order:
    1. Close all perpetual shorts
    2. Exit GLP position to ETH
    3. Repay all borrowed ETH
    4. Unwind Pendle LP position
    5. Convert PT to underlying
    6. Exit to ETH or stables
    
    Target: Complete in <10 minutes
```

## Phase 5: Yield Optimization & Compounding

### Step 11: Reward Management
```
11.1 Claim scheduling
    Optimal frequency by token:
    
    Daily claims:
    - GMX ETH distributions
    - High value, auto-compound
    
    Weekly claims:
    - PENDLE rewards
    - esGMX rewards
    - Accumulated fees
    
    Monthly:
    - Vested esGMX
    - Multiplier point rewards
    - LP incentives

11.2 Compound strategies
    Reinvestment paths:
    
    ETH distributions:
    - Add to GLP for compound
    - Or increase Pendle position
    - Or reduce borrow costs
    
    PENDLE tokens:
    - Sell 50% for ETH
    - Lock 50% for vePENDLE
    - Boost future yields

    esGMX handling:
    - Stake for more rewards
    - Vest over 365 days
    - Compound multiplier points

11.3 Tax optimization
    Track for reporting:
    - Fixed income (Pendle yields)
    - Variable income (GLP fees)
    - Reward token income
    - Trading gains/losses
    - Interest expense deductions

11.4 Yield enhancement tactics
    Advanced optimizations:
    
    Bribe markets:
    - Vote incentives on Pendle
    - Direct vePENDLE for bribes
    - Earn additional 2-5% APY
    
    Liquidity provision:
    - Provide liquidity for rewards
    - Stake LP tokens earned
    - Layer additional yields
```

### Step 12: Advanced Strategies
```
12.1 Cross-protocol arbitrage
    Opportunities to monitor:
    
    PT pricing inefficiencies:
    - Pendle vs other fixed rate
    - Maturity spread trades
    - YT/PT basis trades
    
    GLP arbitrage:
    - Entry/exit fee opportunities
    - Composition imbalances
    - Cross-chain GLP differences

12.2 Structured products overlay
    Additional yield sources:
    
    Options strategies:
    - Sell covered calls on ETH
    - Use portion of position
    - Generate 5-10% additional
    
    Liquidity provision:
    - Use excess collateral
    - Provide to related pools
    - Earn trading fees

12.3 Flash loan optimizations
    Capital efficient entries:
    
    Position building:
    - Flash borrow entry capital
    - Build entire position
    - Repay from position itself
    - Zero upfront cost
    
    Rebalancing:
    - Flash loans for large moves
    - No slippage on rebalancing
    - Atomic transactions

12.4 MEV protection strategies
    Protect large transactions:
    
    Private mempools:
    - Use Flashbots Protect
    - Submit via private relay
    - Avoid frontrunning
    
    Time transactions:
    - Execute during low activity
    - Bundle related transactions
    - Use commit-reveal patterns
```

## Phase 6: Risk Management & Reporting

### Step 13: Risk Controls
```
13.1 Position limits
    Hard constraints:
    - Maximum leverage: 2x
    - Single protocol: <40% exposure
    - Maximum position: $100M
    - Minimum health: 1.3
    - Delta tolerance: ±25 ETH

13.2 Circuit breakers
    Automatic triggers:
    
    Protocol risks:
    - PT depeg >10%: REDUCE 50%
    - GLP discount >5%: EXIT GLP
    - Lending freeze: FULL EXIT
    
    Market risks:
    - ETH flash crash >20%: DELEVERAGE
    - Funding >0.1% daily: ROTATE
    - Gas >500 gwei: PAUSE

13.3 Redundancy systems
    Backup infrastructure:
    
    RPC endpoints:
    - Primary: Dedicated node
    - Backup: Infura/Alchemy
    - Emergency: Public RPC
    
    Execution venues:
    - Primary: FordeFi
    - Backup: Direct contracts
    - Emergency: Manual intervention

13.4 Insurance strategies
    Risk mitigation:
    - Protocol coverage (if available)
    - Smart contract insurance
    - Maintain 20% cash reserve
    - Hedging tail risks
```

### Step 14: Reporting & Analytics
```
14.1 Real-time dashboard
    Live metrics updated continuously:
    
    Position overview:
    - Total value locked
    - Current leverage ratio
    - Health factor graph
    - Delta exposure meter
    
    Yield metrics:
    - Current APY (live)
    - 24h/7d/30d returns
    - Reward accumulation
    - Fee generation
    
    Risk indicators:
    - Liquidation distance
    - Funding rate trends
    - Protocol TVL changes
    - Gas cost tracking

14.2 Daily report generation
    Comprehensive summary at 00:00 UTC:
    
    Position summary:
    - Starting values all positions
    - Ending values all positions
    - Net change in ETH terms
    - Net change in USD terms
    
    Performance breakdown:
    - Pendle yield earned
    - GLP fees collected
    - Rewards claimed
    - Funding paid/received
    - Borrow costs incurred
    - Gas fees spent
    - Net profit in ETH
    
    Risk report:
    - Minimum health factor
    - Maximum leverage hit
    - Delta drift instances
    - Rebalance count
    - Near miss events
    
    Actions log:
    - All transactions executed
    - Rebalancing performed
    - Claims processed
    - Alerts triggered

14.3 Alert configuration
    Multi-channel notifications:
    
    Critical (immediate):
    - Health factor <1.3
    - Position liquidated
    - Protocol exploit detected
    - Delta drift >50 ETH
    
    Warning (5 min delay):
    - Funding rate spike
    - PT liquidity drop
    - GLP imbalance
    - Gas spike detected
    
    Info (hourly digest):
    - Yields claimed
    - Rebalance complete
    - Daily P&L update
    - System health check
```

## Complete Automation Flow

The system operates on multiple concurrent loops with different intervals:

**Every 30 seconds - Critical monitoring:**
- Check all position health factors
- Monitor PT/underlying ratio
- Verify GLP composition
- Track perpetual margins
- Calculate current delta
- Detect liquidation risks

**Every 5 minutes - Market scanning:**
- Scan all Pendle markets
- Check new PT offerings
- Monitor GLP fee rates
- Track funding rates
- Identify arbitrage opportunities
- Calculate yield optimizations

**Every 15 minutes - Rebalancing logic:**
- Evaluate delta drift
- Rebalance if needed
- Adjust leverage ratios
- Optimize funding costs
- Rotate venues if beneficial
- Compound small rewards

**Every hour - Maintenance:**
- Claim accumulated ETH
- Process PENDLE rewards
- Update price feeds
- Clear pending operations
- Archive metrics data

**Every 24 hours - Operations:**
- Generate full report
- Claim all rewards
- Compound positions
- Review strategy parameters
- Plan next day execution
- Send performance summary

**Continuous monitoring:**
- Log all transactions
- Track gas prices
- Monitor protocol TVLs
- Watch for exploits
- Update dashboards
- Record metrics

**Error handling:**
- Catch all exceptions
- Classify severity
- Execute fallback logic
- Alert on failures
- Attempt recovery
- Log all events

## Implementation Roadmap

### Week 1: Infrastructure
- Deploy monitoring systems
- Integrate Pendle protocols
- Set up lending connections
- Implement health tracking

### Week 2: Core Positions
- Pendle PT/LP logic
- Borrowing automation
- GLP integration
- Basic hedging

### Week 3: Advanced Features
- Multi-venue perpetuals
- Rebalancing algorithms
- Reward compounding
- Emergency procedures

### Week 4: Production
- Complete testing suite
- Deploy with limits
- Monitor for stability
- Scale gradually

## Economic Analysis

### Setup Costs
- Initial gas: ~$1,000
- Position building: ~$500
- Perpetual setup: ~$300
- Total one-time: ~$1,800

### Ongoing Costs (Monthly)
- Monitoring gas: ~$200
- Rebalancing: ~$500
- Claims: ~$300
- Emergency reserves: ~$200
- Total monthly: ~$1,200

### Expected Returns
- Gross APY: ~35% in ETH terms
- Net APY: ~32% after costs
- ETH return: 320 ETH/year on 1000 ETH
- USD value: ~$640k at $2000 ETH
- Break-even: <1 week

## Critical Success Factors

1. **Maintain PT Liquidity** - Must be able to exit Pendle position quickly
2. **Monitor GLP Composition** - ETH weight changes affect delta
3. **Track All Funding** - Multiple perpetual venues need coordination
4. **Health Factor >1.5** - Never compromise safety for yield
5. **Compound Weekly** - Maximize compounding effect
6. **Have Exit Plan** - Full unwind possible in <15 minutes

## Emergency Procedures

### PT Liquidity Crisis
1. Stop new positions immediately
2. Use YT market for exit if needed
3. Accept slippage to exit
4. Repay all debts
5. Document losses

### GLP Depeg Event
1. Exit GLP immediately
2. Accept exit fees
3. Adjust perpetual hedge
4. Move to alternative venue
5. Wait for stability

### Smart Contract Risk
1. Monitor all protocols continuously
2. Exit affected protocol immediately
3. Use flashloans if needed
4. Accept any slippage
5. Post-mortem analysis

### Market Crash Scenario
1. Auto-deleverage at preset levels
2. Close perpetuals first
3. Exit GLP if needed
4. Maintain core PT position
5. Wait for recovery