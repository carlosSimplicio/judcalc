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
