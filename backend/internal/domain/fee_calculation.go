package domain

import (
	"errors"
	"math"
)

var ErrInvalidFeeCalculation = errors.New("invalid fee calculation")

type FeeLevel string

const (
	FeeLevelLow    FeeLevel = "low"
	FeeLevelMedium FeeLevel = "medium"
	FeeLevelHigh   FeeLevel = "high"
)

type FeeCalculationInput struct {
	EstimatedHours        float64
	BillableHoursPerMonth float64
	Complexity            FeeLevel
	Risk                  FeeLevel
}

type FeeCalculation struct {
	MonthlyFixedCostsCents      int64
	OperationalHourCostCents    *int64
	MinimumSustainableCostCents *int64
	TechnicalEstimateCents      *int64
	ComplexityFactor            float64
	RiskFactor                  float64
}

var complexityFactors = map[FeeLevel]float64{
	FeeLevelLow: 1, FeeLevelMedium: 1.25, FeeLevelHigh: 1.5,
}

var riskFactors = map[FeeLevel]float64{
	FeeLevelLow: 1, FeeLevelMedium: 1.1, FeeLevelHigh: 1.2,
}

func ValidateFeeCalculationInput(input FeeCalculationInput) error {
	if !positiveFinite(input.EstimatedHours) || !positiveFinite(input.BillableHoursPerMonth) {
		return ErrInvalidFeeCalculation
	}
	if _, ok := complexityFactors[input.Complexity]; !ok {
		return ErrInvalidFeeCalculation
	}
	if _, ok := riskFactors[input.Risk]; !ok {
		return ErrInvalidFeeCalculation
	}
	return nil
}

func CalculateFee(costs FixedCosts, input FeeCalculationInput) (FeeCalculation, error) {
	if err := ValidateFeeCalculationInput(input); err != nil {
		return FeeCalculation{}, err
	}
	complexityFactor := complexityFactors[input.Complexity]
	riskFactor := riskFactors[input.Risk]
	result := FeeCalculation{
		MonthlyFixedCostsCents: costs.MonthlyTotalCents(),
		ComplexityFactor:       complexityFactor,
		RiskFactor:             riskFactor,
	}
	if result.MonthlyFixedCostsCents == 0 {
		return result, nil
	}
	hourCost := float64(result.MonthlyFixedCostsCents) / input.BillableHoursPerMonth
	minimum := hourCost * input.EstimatedHours
	technical := minimum * complexityFactor * riskFactor
	result.OperationalHourCostCents = roundedCents(hourCost)
	result.MinimumSustainableCostCents = roundedCents(minimum)
	result.TechnicalEstimateCents = roundedCents(technical)
	return result, nil
}

func positiveFinite(value float64) bool {
	return value > 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
}

func roundedCents(value float64) *int64 {
	rounded := int64(math.Round(value))
	return &rounded
}
