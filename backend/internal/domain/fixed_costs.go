package domain

import (
	"errors"
	"math"
)

var ErrFixedCostsOverflow = errors.New("fixed costs total overflow")

type FixedCosts struct {
	UserID                              int64
	OABAnnualFeeCents                   int64
	DigitalCertificateCents             int64
	AccountantCents                     int64
	LegalSoftwareCents                  int64
	InternetCents                       int64
	PhoneCents                          int64
	RecurringTransportCents             int64
	CoworkingOrOfficeRentCents          int64
	ProfessionalDomainWebsiteEmailCents int64
	MarketingCents                      int64
	OfficeSuppliesCents                 int64
	EquipmentAndDepreciationCents       int64
	OtherCostsCents                     int64
}

func MonthlyAverageCents(annualAmountCents int64) int64 {
	monthly := annualAmountCents / 12
	if annualAmountCents%12 >= 6 {
		monthly++
	}
	return monthly
}

func (costs FixedCosts) MonthlyTotalCents() (int64, error) {
	amounts := [...]int64{
		MonthlyAverageCents(costs.OABAnnualFeeCents), costs.DigitalCertificateCents,
		costs.AccountantCents, costs.LegalSoftwareCents, costs.InternetCents,
		costs.PhoneCents, costs.RecurringTransportCents,
		costs.CoworkingOrOfficeRentCents, costs.ProfessionalDomainWebsiteEmailCents,
		costs.MarketingCents, costs.OfficeSuppliesCents,
		costs.EquipmentAndDepreciationCents, costs.OtherCostsCents,
	}
	total := int64(0)
	for _, amount := range amounts {
		if amount > 0 && total > math.MaxInt64-amount ||
			amount < 0 && total < math.MinInt64-amount {
			return 0, ErrFixedCostsOverflow
		}
		total += amount
	}
	return total, nil
}

type FixedCostsPatch struct {
	UserID                              int64
	OABAnnualFeeCents                   *int64
	DigitalCertificateCents             *int64
	AccountantCents                     *int64
	LegalSoftwareCents                  *int64
	InternetCents                       *int64
	PhoneCents                          *int64
	RecurringTransportCents             *int64
	CoworkingOrOfficeRentCents          *int64
	ProfessionalDomainWebsiteEmailCents *int64
	MarketingCents                      *int64
	OfficeSuppliesCents                 *int64
	EquipmentAndDepreciationCents       *int64
	OtherCostsCents                     *int64
}
