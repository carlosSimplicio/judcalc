package domain

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

func (costs FixedCosts) MonthlyTotalCents() int64 {
	return MonthlyAverageCents(costs.OABAnnualFeeCents) +
		costs.DigitalCertificateCents + costs.AccountantCents +
		costs.LegalSoftwareCents + costs.InternetCents + costs.PhoneCents +
		costs.RecurringTransportCents + costs.CoworkingOrOfficeRentCents +
		costs.ProfessionalDomainWebsiteEmailCents + costs.MarketingCents +
		costs.OfficeSuppliesCents + costs.EquipmentAndDepreciationCents +
		costs.OtherCostsCents
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
