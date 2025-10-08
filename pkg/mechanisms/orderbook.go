package mechanisms

import (
	"context"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
)

// OrderBook represents order book based trading mechanisms.
// This interface models CEX-style limit order books where trades occur
// through matching buy and sell orders at specific price levels.
//
// Contract:
//   - BestBid and BestAsk must return the top of book (best available prices)
//   - PlaceOrder must validate order parameters and return a unique OrderID
//   - CancelOrder must accept OrderIDs from PlaceOrder
//   - Orders must be validated (positive sizes, valid prices)
//
// Error Conditions:
//   - Invalid order parameters (negative size, invalid price)
//   - Order would cross the book inappropriately
//   - Insufficient liquidity at price level
//   - Invalid order ID for cancellation
//
// Thread Safety: Implementations are not required to be thread-safe.
type OrderBook interface {
	MarketMechanism

	// BestBid returns the highest bid price and available size.
	// If the order book has no bids, returns zero price and amount with no error.
	//
	// Returns:
	//   - Price: Best bid price
	//   - Amount: Total size available at best bid
	//   - error: Returns error only if query fails (not if book is empty)
	BestBid(ctx context.Context) (primitives.Price, primitives.Amount, error)

	// BestAsk returns the lowest ask price and available size.
	// If the order book has no asks, returns zero price and amount with no error.
	//
	// Returns:
	//   - Price: Best ask price
	//   - Amount: Total size available at best ask
	//   - error: Returns error only if query fails (not if book is empty)
	BestAsk(ctx context.Context) (primitives.Price, primitives.Amount, error)

	// PlaceOrder simulates placing an order in the order book.
	// Returns a unique identifier that can be used to cancel the order.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - order: Order details (side, type, price, size, etc.)
	//
	// Returns:
	//   - OrderID: Unique identifier for the placed order
	//   - error: Returns error if order parameters are invalid or placement fails
	PlaceOrder(ctx context.Context, order Order) (OrderID, error)

	// CancelOrder simulates canceling an existing order.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - id: Order identifier from PlaceOrder
	//
	// Returns:
	//   - error: Returns error if order doesn't exist or cancellation fails
	CancelOrder(ctx context.Context, id OrderID) error

	// Depth returns the order book depth (bids and asks) up to a certain level.
	// Level 1 = best bid/ask, Level 2 = top N levels, Level 3 = full book.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeouts
	//   - levels: Number of price levels to return (0 = full book)
	//
	// Returns:
	//   - OrderBookDepth: Bids and asks at each price level
	//   - error: Returns error if query fails
	Depth(ctx context.Context, levels int) (OrderBookDepth, error)
}

// OrderID uniquely identifies an order in the order book.
// The format is implementation-specific (could be UUID, sequential number, hash, etc.).
type OrderID string

// OrderSide represents whether an order is buying or selling.
type OrderSide string

const (
	// OrderSideBuy represents a buy order (bid)
	OrderSideBuy OrderSide = "buy"

	// OrderSideSell represents a sell order (ask)
	OrderSideSell OrderSide = "sell"
)

// OrderType represents the type of order execution.
type OrderType string

const (
	// OrderTypeLimit is a limit order (executes at specified price or better)
	OrderTypeLimit OrderType = "limit"

	// OrderTypeMarket is a market order (executes immediately at best available price)
	OrderTypeMarket OrderType = "market"

	// OrderTypeStopLoss is a stop loss order (becomes market order when price reached)
	OrderTypeStopLoss OrderType = "stop_loss"

	// OrderTypeStopLimit is a stop limit order (becomes limit order when price reached)
	OrderTypeStopLimit OrderType = "stop_limit"
)

// TimeInForce represents how long an order remains active.
type TimeInForce string

const (
	// TimeInForceGTC (Good Till Cancel) remains active until filled or canceled
	TimeInForceGTC TimeInForce = "GTC"

	// TimeInForceIOC (Immediate Or Cancel) fills immediately or cancels
	TimeInForceIOC TimeInForce = "IOC"

	// TimeInForceFOK (Fill Or Kill) fills completely immediately or cancels
	TimeInForceFOK TimeInForce = "FOK"

	// TimeInForceGTD (Good Till Date) remains active until specified time
	TimeInForceGTD TimeInForce = "GTD"
)

// Order represents an order to place in the order book.
type Order struct {
	// Side is the order side (buy or sell)
	Side OrderSide

	// Type is the order type (limit, market, stop, etc.)
	Type OrderType

	// Price is the limit price (for limit orders)
	// Ignored for market orders
	Price primitives.Price

	// Size is the order quantity
	Size primitives.Amount

	// TimeInForce specifies how long the order remains active
	TimeInForce TimeInForce

	// StopPrice is the trigger price for stop orders
	StopPrice primitives.Price

	// ExpiryTime is the expiration time for GTD orders
	ExpiryTime primitives.Time

	// Additional order parameters
	Metadata map[string]interface{}
}

// PriceLevel represents a single price level in the order book.
type PriceLevel struct {
	// Price is the price at this level
	Price primitives.Price

	// Size is the total quantity available at this price
	Size primitives.Amount

	// OrderCount is the number of orders at this price level
	OrderCount int
}

// OrderBookDepth represents the state of the order book.
type OrderBookDepth struct {
	// Bids are the buy orders (sorted by price descending)
	Bids []PriceLevel

	// Asks are the sell orders (sorted by price ascending)
	Asks []PriceLevel

	// Timestamp is the time this depth snapshot was taken
	Timestamp primitives.Time
}
