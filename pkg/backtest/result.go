package backtest

import (
	"fmt"
	"math"

	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/primitives"
	"github.com/johnayoung/go-crypto-quant-toolkit/pkg/strategy"
)

// Result contains the outcomes of a backtest execution.
// It includes both raw data (portfolio value over time) and calculated
// performance metrics (returns, risk-adjusted returns, drawdown).
//
// All metrics use precise decimal arithmetic to avoid floating-point errors
// in financial calculations.
type Result struct {
	// InitialValue is the starting portfolio value
	InitialValue primitives.Amount

	// FinalValue is the ending portfolio value
	FinalValue primitives.Amount

	// ValueHistory tracks portfolio value at each rebalancing point
	ValueHistory []ValuePoint

	// Portfolio is the final portfolio state after backtest completion
	Portfolio *strategy.Portfolio

	// Calculated metrics (populated by calculateMetrics)
	TotalReturn       primitives.Decimal // Total return as decimal (e.g., 0.15 = 15%)
	AnnualizedReturn  primitives.Decimal // Annualized return
	Sharpe            primitives.Decimal // Sharpe ratio (assuming 0 risk-free rate)
	MaxDrawdown       primitives.Decimal // Maximum drawdown as decimal (e.g., 0.20 = 20%)
	MaxDrawdownAmount primitives.Amount  // Maximum drawdown in absolute terms
}

// ValuePoint represents the portfolio value at a specific point in time.
type ValuePoint struct {
	Time  primitives.Time
	Value primitives.Amount
}

// calculateMetrics computes derived performance metrics from the backtest results.
// This method is called automatically by Engine.Run() after backtest completion.
//
// Calculated metrics:
//   - TotalReturn: (FinalValue - InitialValue) / InitialValue
//   - AnnualizedReturn: Annualized total return based on time period
//   - Sharpe: Risk-adjusted return (return / volatility), assumes 0 risk-free rate
//   - MaxDrawdown: Largest peak-to-trough decline as percentage
//   - MaxDrawdownAmount: Largest peak-to-trough decline in absolute terms
func (r *Result) calculateMetrics() error {
	if r.InitialValue.IsZero() {
		return fmt.Errorf("initial value cannot be zero")
	}
	if len(r.ValueHistory) < 2 {
		return fmt.Errorf("insufficient value history (need at least 2 points)")
	}

	// Calculate total return
	initialDec := r.InitialValue.Decimal()
	finalDec := r.FinalValue.Decimal()
	returnDec, err := finalDec.Sub(initialDec).Div(initialDec)
	if err != nil {
		return fmt.Errorf("failed to calculate total return: %w", err)
	}
	r.TotalReturn = returnDec

	// Calculate annualized return
	if err := r.calculateAnnualizedReturn(); err != nil {
		return fmt.Errorf("failed to calculate annualized return: %w", err)
	}

	// Calculate Sharpe ratio (using period-to-period returns)
	if err := r.calculateSharpe(); err != nil {
		return fmt.Errorf("failed to calculate Sharpe ratio: %w", err)
	}

	// Calculate maximum drawdown
	if err := r.calculateMaxDrawdown(); err != nil {
		return fmt.Errorf("failed to calculate max drawdown: %w", err)
	}

	return nil
}

// calculateAnnualizedReturn computes the annualized return based on the time period.
// Formula: AnnualizedReturn = (1 + TotalReturn)^(365.25*24*60*60 / period_seconds) - 1
func (r *Result) calculateAnnualizedReturn() error {
	if len(r.ValueHistory) < 2 {
		return fmt.Errorf("insufficient history")
	}

	// Get time period in seconds
	startTime := r.ValueHistory[0].Time
	endTime := r.ValueHistory[len(r.ValueHistory)-1].Time
	periodSeconds := endTime.Sub(startTime).Seconds()

	if periodSeconds <= 0 {
		return fmt.Errorf("invalid time period: %f seconds", periodSeconds)
	}

	// Seconds in a year (accounting for leap years)
	const secondsPerYear = 365.25 * 24 * 60 * 60

	// Convert to float64 for exponentiation (necessary for annualization)
	// Note: This is acceptable here as we're calculating a percentage, not a dollar amount
	totalReturnFloat := r.TotalReturn.Float64()
	exponent := secondsPerYear / periodSeconds

	// Calculate: (1 + TotalReturn)^(secondsPerYear/periodSeconds) - 1
	annualizedFloat := math.Pow(1+totalReturnFloat, exponent) - 1

	r.AnnualizedReturn = primitives.NewDecimalFromFloat(annualizedFloat)
	return nil
}

// calculateSharpe computes the Sharpe ratio using point-to-point returns.
// Formula: Sharpe = Mean(returns) / StdDev(returns) * sqrt(periods_per_year)
// Assumes risk-free rate = 0
func (r *Result) calculateSharpe() error {
	if len(r.ValueHistory) < 2 {
		return fmt.Errorf("insufficient history for Sharpe calculation")
	}

	// Calculate point-to-point returns
	returns := make([]primitives.Decimal, 0, len(r.ValueHistory)-1)
	for i := 1; i < len(r.ValueHistory); i++ {
		prevValue := r.ValueHistory[i-1].Value.Decimal()
		currValue := r.ValueHistory[i].Value.Decimal()

		if prevValue.IsZero() {
			continue // Skip if previous value is zero
		}

		ret, err := currValue.Sub(prevValue).Div(prevValue)
		if err != nil {
			continue // Skip on division errors
		}
		returns = append(returns, ret)
	}

	if len(returns) < 2 {
		// Not enough returns to calculate volatility
		r.Sharpe = primitives.Zero()
		return nil
	}

	// Calculate mean return
	sum := primitives.Zero()
	for _, ret := range returns {
		sum = sum.Add(ret)
	}
	nReturns := primitives.NewDecimal(int64(len(returns)))
	mean, err := sum.Div(nReturns)
	if err != nil {
		return fmt.Errorf("failed to calculate mean: %w", err)
	}

	// Calculate standard deviation
	varianceSum := primitives.Zero()
	for _, ret := range returns {
		diff := ret.Sub(mean)
		varianceSum = varianceSum.Add(diff.Mul(diff))
	}
	variance, err := varianceSum.Div(nReturns)
	if err != nil {
		return fmt.Errorf("failed to calculate variance: %w", err)
	}

	// Convert to float for sqrt (standard deviation calculation)
	varianceFloat := variance.Float64()
	stdDev := primitives.NewDecimalFromFloat(math.Sqrt(varianceFloat))

	if stdDev.IsZero() {
		// Zero volatility means infinite Sharpe, but we'll set to zero
		r.Sharpe = primitives.Zero()
		return nil
	}

	// Calculate average time between snapshots (for annualization)
	totalSeconds := r.ValueHistory[len(r.ValueHistory)-1].Time.Sub(r.ValueHistory[0].Time).Seconds()
	avgSecondsPerPeriod := totalSeconds / float64(len(returns))

	// Periods per year
	const secondsPerYear = 365.25 * 24 * 60 * 60
	periodsPerYear := secondsPerYear / avgSecondsPerPeriod

	// Sharpe = (Mean / StdDev) * sqrt(periods_per_year)
	sharpeRaw, err := mean.Div(stdDev)
	if err != nil {
		return fmt.Errorf("failed to calculate Sharpe: %w", err)
	}
	annualizationFactor := primitives.NewDecimalFromFloat(math.Sqrt(periodsPerYear))
	sharpe := sharpeRaw.Mul(annualizationFactor)

	r.Sharpe = sharpe
	return nil
}

// calculateMaxDrawdown computes the maximum peak-to-trough decline.
// Drawdown = (Trough - Peak) / Peak
func (r *Result) calculateMaxDrawdown() error {
	if len(r.ValueHistory) < 2 {
		return fmt.Errorf("insufficient history")
	}

	maxDrawdown := primitives.Zero()
	maxDrawdownAmount := primitives.Zero()
	peak := r.ValueHistory[0].Value.Decimal()

	for i := 1; i < len(r.ValueHistory); i++ {
		currentValue := r.ValueHistory[i].Value.Decimal()

		// Update peak if we've reached a new high
		if currentValue.GreaterThan(peak) {
			peak = currentValue
		}

		// Calculate drawdown from peak
		if peak.IsPositive() {
			drawdownAmt := peak.Sub(currentValue)
			drawdown, err := drawdownAmt.Div(peak)
			if err != nil {
				continue // Skip on division errors
			}

			if drawdown.GreaterThan(maxDrawdown) {
				maxDrawdown = drawdown
				maxDrawdownAmount = drawdownAmt
			}
		}
	}

	r.MaxDrawdown = maxDrawdown
	amt, err := primitives.NewAmount(maxDrawdownAmount)
	if err != nil {
		// If amount is negative due to calculation, use zero
		r.MaxDrawdownAmount = primitives.ZeroAmount()
	} else {
		r.MaxDrawdownAmount = amt
	}
	return nil
}

// Summary returns a human-readable summary of the backtest results.
func (r *Result) Summary() string {
	totalRetPct := r.TotalReturn.Mul(primitives.NewDecimal(100))
	annRetPct := r.AnnualizedReturn.Mul(primitives.NewDecimal(100))
	maxDDPct := r.MaxDrawdown.Mul(primitives.NewDecimal(100))

	return fmt.Sprintf(
		"Backtest Results:\n"+
			"  Initial Value: %s\n"+
			"  Final Value: %s\n"+
			"  Total Return: %.2f%%\n"+
			"  Annualized Return: %.2f%%\n"+
			"  Sharpe Ratio: %.2f\n"+
			"  Max Drawdown: %.2f%% (%s)\n"+
			"  Data Points: %d",
		r.InitialValue.String(),
		r.FinalValue.String(),
		totalRetPct.Float64(),
		annRetPct.Float64(),
		r.Sharpe.Float64(),
		maxDDPct.Float64(),
		r.MaxDrawdownAmount.String(),
		len(r.ValueHistory),
	)
}
