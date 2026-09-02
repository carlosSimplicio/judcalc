package domain

import (
	"errors"
	"math"
	"testing"
)

func TestMonthlyTotalCentsIncludesMonthlyOABAverageAndEveryCategory(t *testing.T) {
	costs := FixedCosts{
		OABAnnualFeeCents: 120006, DigitalCertificateCents: 1,
		AccountantCents: 2, LegalSoftwareCents: 3, InternetCents: 4,
		PhoneCents: 5, RecurringTransportCents: 6,
		CoworkingOrOfficeRentCents: 7, ProfessionalDomainWebsiteEmailCents: 8,
		MarketingCents: 9, OfficeSuppliesCents: 10,
		EquipmentAndDepreciationCents: 11, OtherCostsCents: 12,
	}
	if got, want := costs.MonthlyTotalCents(), int64(10079); got != want {
		t.Fatalf("monthly total = %d, want %d", got, want)
	}
}

func TestCalculateFeeUsesFullPrecisionAndFactors(t *testing.T) {
	result, err := CalculateFee(FixedCosts{InternetCents: 10001}, FeeCalculationInput{
		EstimatedHours: 3.5, BillableHoursPerMonth: 6,
		Complexity: FeeLevelMedium, Risk: FeeLevelHigh,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.ComplexityFactor != 1.25 || result.RiskFactor != 1.20 {
		t.Fatalf("factors = %v and %v", result.ComplexityFactor, result.RiskFactor)
	}
	if got, want := *result.OperationalHourCostCents, int64(1667); got != want {
		t.Fatalf("hour cost = %d, want %d", got, want)
	}
	if got, want := *result.MinimumSustainableCostCents, int64(5834); got != want {
		t.Fatalf("minimum = %d, want %d", got, want)
	}
	if got, want := *result.TechnicalEstimateCents, int64(8751); got != want {
		t.Fatalf("technical = %d, want %d", got, want)
	}
}

func TestCalculateFeeReturnsNilEconomicValuesWhenMonthlyCostsAreZero(t *testing.T) {
	result, err := CalculateFee(FixedCosts{}, FeeCalculationInput{
		EstimatedHours: 10, BillableHoursPerMonth: 80,
		Complexity: FeeLevelLow, Risk: FeeLevelLow,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.MonthlyFixedCostsCents != 0 || result.OperationalHourCostCents != nil ||
		result.MinimumSustainableCostCents != nil || result.TechnicalEstimateCents != nil {
		t.Fatalf("unexpected zero-cost result: %#v", result)
	}
}

func TestCalculateFeeRejectsInvalidInputsAndLevels(t *testing.T) {
	tests := []FeeCalculationInput{
		{EstimatedHours: 0, BillableHoursPerMonth: 80, Complexity: FeeLevelLow, Risk: FeeLevelLow},
		{EstimatedHours: 10, BillableHoursPerMonth: -1, Complexity: FeeLevelLow, Risk: FeeLevelLow},
		{EstimatedHours: 10, BillableHoursPerMonth: 80, Complexity: "unknown", Risk: FeeLevelLow},
		{EstimatedHours: 10, BillableHoursPerMonth: 80, Complexity: FeeLevelLow, Risk: "unknown"},
	}
	for _, input := range tests {
		if _, err := CalculateFee(FixedCosts{InternetCents: 100}, input); !errors.Is(err, ErrInvalidFeeCalculation) {
			t.Fatalf("input %#v returned %v", input, err)
		}
	}
}

func TestCalculateFeeAppliesEveryConfiguredLevel(t *testing.T) {
	wantComplexity := map[FeeLevel]float64{FeeLevelLow: 1, FeeLevelMedium: 1.25, FeeLevelHigh: 1.5}
	wantRisk := map[FeeLevel]float64{FeeLevelLow: 1, FeeLevelMedium: 1.1, FeeLevelHigh: 1.2}
	for complexity, complexityFactor := range wantComplexity {
		for risk, riskFactor := range wantRisk {
			result, err := CalculateFee(FixedCosts{InternetCents: 10000}, FeeCalculationInput{
				EstimatedHours: 1, BillableHoursPerMonth: 1,
				Complexity: complexity, Risk: risk,
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.ComplexityFactor != complexityFactor || result.RiskFactor != riskFactor {
				t.Fatalf("levels %q/%q returned %#v", complexity, risk, result)
			}
		}
	}
}

func TestValidateFeeCalculationInputRejectsNonFiniteHours(t *testing.T) {
	inputs := []FeeCalculationInput{
		{EstimatedHours: math.NaN(), BillableHoursPerMonth: 1, Complexity: FeeLevelLow, Risk: FeeLevelLow},
		{EstimatedHours: 1, BillableHoursPerMonth: math.Inf(1), Complexity: FeeLevelLow, Risk: FeeLevelLow},
	}
	for _, input := range inputs {
		if err := ValidateFeeCalculationInput(input); !errors.Is(err, ErrInvalidFeeCalculation) {
			t.Fatalf("input %#v returned %v", input, err)
		}
	}
}
