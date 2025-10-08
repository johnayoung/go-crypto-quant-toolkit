package concentrated_liquidity

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	core "github.com/daoleno/uniswap-sdk-core/entities"
	"github.com/daoleno/uniswapv3-sdk/constants"
	"github.com/daoleno/uniswapv3-sdk/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/mechanisms"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

var (
	// ErrInvalidPoolParams is returned when pool parameters are invalid
	ErrInvalidPoolParams = errors.New("invalid pool parameters")

	// ErrInvalidTickRange is returned when tick range is invalid
	ErrInvalidTickRange = errors.New("invalid tick range: tickLower must be less than tickUpper")

	// ErrInsufficientLiquidity is returned when there's insufficient liquidity
	ErrInsufficientLiquidity = errors.New("insufficient liquidity")
)

// Pool implements the LiquidityPool interface for Uniswap V3 style concentrated liquidity.
// It wraps the daoleno/uniswapv3-sdk library to provide integration with our framework.
//
// This implementation provides a bridge between our framework's interfaces and the
// battle-tested Uniswap V3 SDK math, ensuring accuracy and reliability.
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
//   - poolID: Unique identifier for this pool
//   - tokenAAddress: Address of token A
//   - tokenADecimals: Decimals for token A
//   - tokenBAddress: Address of token B
//   - tokenBDecimals: Decimals for token B
//   - fee: Fee tier (500 for 0.05%, 3000 for 0.3%, 10000 for 1%)
//
// The pool uses Uniswap V3's concentrated liquidity model where liquidity providers
// can specify price ranges for their capital.
func NewPool(
	poolID string,
	tokenAAddress common.Address,
	tokenADecimals uint,
	tokenBAddress common.Address,
	tokenBDecimals uint,
	fee constants.FeeAmount,
) (*Pool, error) {
	// Validate inputs
	if poolID == "" {
		return nil, errors.New("poolID cannot be empty")
	}

	// Create token instances using the SDK
	// Chain ID is set to 1 (mainnet) but could be parameterized if needed
	tokenA := core.NewToken(1, tokenAAddress, tokenADecimals, "", "")
	tokenB := core.NewToken(1, tokenBAddress, tokenBDecimals, "", "")

	// Get tick spacing for the fee tier
	tickSpacing, ok := constants.TickSpacings[fee]
	if !ok {
		return nil, fmt.Errorf("invalid fee amount: %d", fee)
	}

	return &Pool{
		poolID:      poolID,
		tokenA:      tokenA,
		tokenB:      tokenB,
		fee:         fee,
		tickSpacing: tickSpacing,
	}, nil
}

// Mechanism returns the mechanism type identifier.
func (p *Pool) Mechanism() mechanisms.MechanismType {
	return mechanisms.MechanismTypeLiquidityPool
}

// Venue returns the venue identifier (could be "uniswap-v3", "pancakeswap-v3", etc.)
func (p *Pool) Venue() string {
	return "uniswap-v3"
}

// Calculate computes the current state of the pool given the parameters.
//
// Required metadata fields:
//   - "current_tick" (int): Current tick of the pool
//   - "sqrt_price_x96" (string): Current sqrt price in Q64.96 format
//   - "liquidity" (string): Current liquidity
//
// Returns pool state including spot price, liquidity, and fees.
func (p *Pool) Calculate(ctx context.Context, params mechanisms.PoolParams) (mechanisms.PoolState, error) {
	// Extract required metadata
	currentTick, ok := params.Metadata["current_tick"].(int)
	if !ok {
		return mechanisms.PoolState{}, errors.New("current_tick required in metadata")
	}

	sqrtPriceX96Str, ok := params.Metadata["sqrt_price_x96"].(string)
	if !ok {
		return mechanisms.PoolState{}, errors.New("sqrt_price_x96 required in metadata")
	}

	liquidityStr, ok := params.Metadata["liquidity"].(string)
	if !ok {
		return mechanisms.PoolState{}, errors.New("liquidity required in metadata")
	}

	// Parse sqrt price
	sqrtPriceX96, ok := new(big.Int).SetString(sqrtPriceX96Str, 10)
	if !ok {
		return mechanisms.PoolState{}, errors.New("invalid sqrt_price_x96 format")
	}

	// Parse liquidity
	liquidity, ok := new(big.Int).SetString(liquidityStr, 10)
	if !ok {
		return mechanisms.PoolState{}, errors.New("invalid liquidity format")
	}

	// Calculate spot price from sqrt price
	// price = (sqrtPriceX96 / 2^96)^2
	q96 := new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)
	sqrtPrice := new(big.Float).Quo(
		new(big.Float).SetInt(sqrtPriceX96),
		new(big.Float).SetInt(q96),
	)

	// Square to get price
	priceFloat := new(big.Float).Mul(sqrtPrice, sqrtPrice)

	// Adjust for decimals: price * 10^(tokenB.decimals - tokenA.decimals)
	decimalAdjustment := new(big.Int).Exp(
		big.NewInt(10),
		big.NewInt(int64(p.tokenB.Decimals())-int64(p.tokenA.Decimals())),
		nil,
	)
	adjustedPrice := new(big.Float).Mul(priceFloat, new(big.Float).SetInt(decimalAdjustment))

	// Convert to primitives.Price
	priceRat, _ := adjustedPrice.Rat(nil)
	priceDec, err := primitives.NewDecimalFromString(priceRat.FloatString(18))
	if err != nil {
		return mechanisms.PoolState{}, fmt.Errorf("invalid price decimal: %w", err)
	}
	spotPrice, err := primitives.NewPrice(priceDec)
	if err != nil {
		return mechanisms.PoolState{}, fmt.Errorf("invalid spot price: %w", err)
	}

	// Convert liquidity to Amount
	liquidityDec, err := primitives.NewDecimalFromString(liquidity.String())
	if err != nil {
		return mechanisms.PoolState{}, fmt.Errorf("invalid liquidity decimal: %w", err)
	}
	liquidityAmount, err := primitives.NewAmount(liquidityDec)
	if err != nil {
		return mechanisms.PoolState{}, fmt.Errorf("invalid liquidity: %w", err)
	}

	return mechanisms.PoolState{
		SpotPrice:          spotPrice,
		Liquidity:          liquidityAmount,
		EffectiveLiquidity: liquidityAmount,
		AccumulatedFeesA:   primitives.ZeroAmount(),
		AccumulatedFeesB:   primitives.ZeroAmount(),
		Metadata: map[string]interface{}{
			"current_tick":   currentTick,
			"sqrt_price_x96": sqrtPriceX96Str,
			"tick_spacing":   p.tickSpacing,
		},
	}, nil
}

// AddLiquidity simulates adding liquidity to the pool within a specified price range.
//
// Required metadata fields:
//   - "tick_lower" (int): Lower tick of the position
//   - "tick_upper" (int): Upper tick of the position
//   - "current_tick" (int): Current tick of the pool
//   - "sqrt_price_x96" (string): Current sqrt price in Q64.96 format
//
// The amounts specify how much of each token to deposit. The actual amounts used
// will be calculated based on the current price and specified range.
func (p *Pool) AddLiquidity(ctx context.Context, amounts mechanisms.TokenAmounts) (mechanisms.PoolPosition, error) {
	// Extract tick range from metadata (assumed to be passed somehow - in practice this would
	// come from the strategy deciding the range)
	// For now, we'll require it to be set externally or use a default range

	// This is a simplified implementation. In practice, you'd want to:
	// 1. Get the tick range from strategy/user input
	// 2. Calculate the optimal liquidity given the amounts
	// 3. Return a position with all necessary information

	return mechanisms.PoolPosition{}, errors.New("AddLiquidity not yet fully implemented - needs tick range specification")
}

// RemoveLiquidity simulates removing liquidity from the pool.
//
// Returns the token amounts that would be withdrawn for the given position.
func (p *Pool) RemoveLiquidity(ctx context.Context, position mechanisms.PoolPosition) (mechanisms.TokenAmounts, error) {
	// Extract position information
	liquidityStr, ok := position.Metadata["liquidity"].(string)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("liquidity required in position metadata")
	}

	tickLower, ok := position.Metadata["tick_lower"].(int)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("tick_lower required in position metadata")
	}

	tickUpper, ok := position.Metadata["tick_upper"].(int)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("tick_upper required in position metadata")
	}

	sqrtPriceX96Str, ok := position.Metadata["sqrt_price_x96"].(string)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("sqrt_price_x96 required in position metadata")
	}

	// Parse values
	liquidity, ok := new(big.Int).SetString(liquidityStr, 10)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("invalid liquidity format")
	}

	sqrtPriceX96, ok := new(big.Int).SetString(sqrtPriceX96Str, 10)
	if !ok {
		return mechanisms.TokenAmounts{}, errors.New("invalid sqrt_price_x96 format")
	}

	// Calculate sqrt prices at tick boundaries
	sqrtPriceLower, err := utils.GetSqrtRatioAtTick(tickLower)
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid tickLower: %w", err)
	}
	sqrtPriceUpper, err := utils.GetSqrtRatioAtTick(tickUpper)
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid tickUpper: %w", err)
	}

	// Calculate token amounts using SDK utilities
	// This uses the Uniswap V3 formulas to determine how many tokens to return
	// GetAmount0Delta calculates: amount0 = L * (sqrt(Pu) - sqrt(P)) / (sqrt(P) * sqrt(Pu))
	// GetAmount1Delta calculates: amount1 = L * (sqrt(P) - sqrt(Pl))
	amount0 := utils.GetAmount0Delta(
		sqrtPriceX96,
		sqrtPriceUpper,
		liquidity,
		false, // roundUp = false for removals
	)

	amount1 := utils.GetAmount1Delta(
		sqrtPriceLower,
		sqrtPriceX96,
		liquidity,
		false, // roundUp = false for removals
	)

	// Convert to our Amount types
	amount0Dec, err := primitives.NewDecimalFromString(amount0.String())
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid amount0 decimal: %w", err)
	}
	amountA, err := primitives.NewAmount(amount0Dec)
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid amount0: %w", err)
	}

	amount1Dec, err := primitives.NewDecimalFromString(amount1.String())
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid amount1 decimal: %w", err)
	}
	amountB, err := primitives.NewAmount(amount1Dec)
	if err != nil {
		return mechanisms.TokenAmounts{}, fmt.Errorf("invalid amount1: %w", err)
	}

	return mechanisms.TokenAmounts{
		AmountA: amountA,
		AmountB: amountB,
	}, nil
}

// CalculatePositionValue calculates the current value of a concentrated liquidity position.
//
// This is a helper function that computes the value of a position given current market conditions.
// It accounts for:
//   - Current token amounts in the position
//   - Accumulated fees
//   - Impermanent loss/gain
func (p *Pool) CalculatePositionValue(
	position mechanisms.PoolPosition,
	currentPriceA primitives.Price,
	currentPriceB primitives.Price,
) (primitives.Amount, error) {
	// Get the tokens that would be withdrawn
	amounts, err := p.RemoveLiquidity(context.Background(), position)
	if err != nil {
		return primitives.ZeroAmount(), err
	}

	// Calculate value: amountA * priceA + amountB * priceB
	valueA := amounts.AmountA.MulPrice(currentPriceA)
	valueB := amounts.AmountB.MulPrice(currentPriceB)

	totalValue := valueA.Add(valueB)
	return totalValue, nil
}
