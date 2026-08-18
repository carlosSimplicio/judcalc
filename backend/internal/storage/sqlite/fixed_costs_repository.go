package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type FixedCostsRepository struct {
	database *sql.DB
}

func NewFixedCostsRepository(database *sql.DB) *FixedCostsRepository {
	return &FixedCostsRepository{database: database}
}

const fixedCostsColumns = `user_id, oab_annual_fee_cents,
	digital_certificate_cents, accountant_cents, legal_software_cents,
	internet_cents, phone_cents, recurring_transport_cents,
	coworking_or_office_rent_cents, professional_domain_website_email_cents,
	marketing_cents, office_supplies_cents, equipment_and_depreciation_cents,
	other_costs_cents`

func (repository *FixedCostsRepository) GetFixedCosts(ctx context.Context, userID string) (domain.FixedCosts, bool, error) {
	row := repository.database.QueryRowContext(ctx,
		"SELECT "+fixedCostsColumns+" FROM user_fixed_costs WHERE user_id = ?",
		userID,
	)
	costs, err := scanFixedCosts(row)
	if err == sql.ErrNoRows {
		return domain.FixedCosts{UserID: userID}, false, nil
	}
	if err != nil {
		return domain.FixedCosts{}, false, fmt.Errorf("buscar custos fixos: %w", err)
	}
	return costs, true, nil
}

func (repository *FixedCostsRepository) UpsertFixedCosts(ctx context.Context, patch domain.FixedCostsPatch) (domain.FixedCosts, error) {
	const query = `
		INSERT INTO user_fixed_costs (` + fixedCostsColumns + `)
		VALUES (
			@user_id, COALESCE(@oab, 0), COALESCE(@certificate, 0),
			COALESCE(@accountant, 0), COALESCE(@software, 0),
			COALESCE(@internet, 0), COALESCE(@phone, 0),
			COALESCE(@transport, 0), COALESCE(@rent, 0),
			COALESCE(@domain, 0), COALESCE(@marketing, 0),
			COALESCE(@supplies, 0), COALESCE(@equipment, 0),
			COALESCE(@other, 0)
		)
		ON CONFLICT(user_id) DO UPDATE SET
			oab_annual_fee_cents = COALESCE(@oab, oab_annual_fee_cents),
			digital_certificate_cents = COALESCE(@certificate, digital_certificate_cents),
			accountant_cents = COALESCE(@accountant, accountant_cents),
			legal_software_cents = COALESCE(@software, legal_software_cents),
			internet_cents = COALESCE(@internet, internet_cents),
			phone_cents = COALESCE(@phone, phone_cents),
			recurring_transport_cents = COALESCE(@transport, recurring_transport_cents),
			coworking_or_office_rent_cents = COALESCE(@rent, coworking_or_office_rent_cents),
			professional_domain_website_email_cents = COALESCE(@domain, professional_domain_website_email_cents),
			marketing_cents = COALESCE(@marketing, marketing_cents),
			office_supplies_cents = COALESCE(@supplies, office_supplies_cents),
			equipment_and_depreciation_cents = COALESCE(@equipment, equipment_and_depreciation_cents),
			other_costs_cents = COALESCE(@other, other_costs_cents)
		RETURNING ` + fixedCostsColumns

	row := repository.database.QueryRowContext(ctx, query, fixedCostsArguments(patch)...)
	costs, err := scanFixedCosts(row)
	if err != nil {
		return domain.FixedCosts{}, fmt.Errorf("salvar custos fixos: %w", err)
	}
	return costs, nil
}

func fixedCostsArguments(patch domain.FixedCostsPatch) []any {
	return []any{
		sql.Named("user_id", patch.UserID),
		sql.Named("oab", patch.OABAnnualFeeCents),
		sql.Named("certificate", patch.DigitalCertificateCents),
		sql.Named("accountant", patch.AccountantCents),
		sql.Named("software", patch.LegalSoftwareCents),
		sql.Named("internet", patch.InternetCents),
		sql.Named("phone", patch.PhoneCents),
		sql.Named("transport", patch.RecurringTransportCents),
		sql.Named("rent", patch.CoworkingOrOfficeRentCents),
		sql.Named("domain", patch.ProfessionalDomainWebsiteEmailCents),
		sql.Named("marketing", patch.MarketingCents),
		sql.Named("supplies", patch.OfficeSuppliesCents),
		sql.Named("equipment", patch.EquipmentAndDepreciationCents),
		sql.Named("other", patch.OtherCostsCents),
	}
}

type rowScanner interface {
	Scan(...any) error
}

func scanFixedCosts(row rowScanner) (domain.FixedCosts, error) {
	var costs domain.FixedCosts
	err := row.Scan(
		&costs.UserID,
		&costs.OABAnnualFeeCents,
		&costs.DigitalCertificateCents,
		&costs.AccountantCents,
		&costs.LegalSoftwareCents,
		&costs.InternetCents,
		&costs.PhoneCents,
		&costs.RecurringTransportCents,
		&costs.CoworkingOrOfficeRentCents,
		&costs.ProfessionalDomainWebsiteEmailCents,
		&costs.MarketingCents,
		&costs.OfficeSuppliesCents,
		&costs.EquipmentAndDepreciationCents,
		&costs.OtherCostsCents,
	)
	return costs, err
}
