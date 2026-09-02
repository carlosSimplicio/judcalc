package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const maxJSONBodyBytes int64 = 64 * 1024

var errRequestBodyTooLarge = errors.New("request body too large")

type FixedCosts struct {
	repository domain.FixedCostsRepository
	logger     *slog.Logger
}

func NewFixedCosts(repository domain.FixedCostsRepository, logger *slog.Logger) *FixedCosts {
	if logger == nil {
		logger = slog.Default()
	}
	return &FixedCosts{repository: repository, logger: logger}
}

type monthlyCostRequest struct {
	MonthlyAmountCents *int64 `json:"monthly_amount_cents"`
}

type annualCostRequest struct {
	AnnualAmountCents *int64 `json:"annual_amount_cents"`
}

type fixedCostsRequestBody struct {
	OABAnnualFee                   *annualCostRequest  `json:"oab_annual_fee"`
	DigitalCertificate             *monthlyCostRequest `json:"digital_certificate"`
	Accountant                     *monthlyCostRequest `json:"accountant"`
	LegalSoftware                  *monthlyCostRequest `json:"legal_software"`
	Internet                       *monthlyCostRequest `json:"internet"`
	Phone                          *monthlyCostRequest `json:"phone"`
	RecurringTransport             *monthlyCostRequest `json:"recurring_transport"`
	CoworkingOrOfficeRent          *monthlyCostRequest `json:"coworking_or_office_rent"`
	ProfessionalDomainWebsiteEmail *monthlyCostRequest `json:"professional_domain_website_email"`
	Marketing                      *monthlyCostRequest `json:"marketing"`
	OfficeSupplies                 *monthlyCostRequest `json:"office_supplies"`
	EquipmentAndDepreciation       *monthlyCostRequest `json:"equipment_and_depreciation"`
	OtherCosts                     *monthlyCostRequest `json:"other_costs"`
}

type patchFixedCostsRequest struct {
	Costs *fixedCostsRequestBody `json:"costs"`
}

type monthlyCostResponse struct {
	MonthlyAmountCents int64 `json:"monthly_amount_cents"`
}

type annualCostResponse struct {
	AnnualAmountCents  int64 `json:"annual_amount_cents"`
	MonthlyAmountCents int64 `json:"monthly_amount_cents"`
}

type fixedCostsResponseBody struct {
	OABAnnualFee                   annualCostResponse  `json:"oab_annual_fee"`
	DigitalCertificate             monthlyCostResponse `json:"digital_certificate"`
	Accountant                     monthlyCostResponse `json:"accountant"`
	LegalSoftware                  monthlyCostResponse `json:"legal_software"`
	Internet                       monthlyCostResponse `json:"internet"`
	Phone                          monthlyCostResponse `json:"phone"`
	RecurringTransport             monthlyCostResponse `json:"recurring_transport"`
	CoworkingOrOfficeRent          monthlyCostResponse `json:"coworking_or_office_rent"`
	ProfessionalDomainWebsiteEmail monthlyCostResponse `json:"professional_domain_website_email"`
	Marketing                      monthlyCostResponse `json:"marketing"`
	OfficeSupplies                 monthlyCostResponse `json:"office_supplies"`
	EquipmentAndDepreciation       monthlyCostResponse `json:"equipment_and_depreciation"`
	OtherCosts                     monthlyCostResponse `json:"other_costs"`
}

type fixedCostsResponse struct {
	UserID int64                  `json:"user_id"`
	Costs  fixedCostsResponseBody `json:"costs"`
}

func (handler *FixedCosts) Get(ctx *gin.Context) {
	user, ok := AuthenticatedUser(ctx)
	if !ok {
		writeError(ctx, http.StatusUnauthorized, "unauthorized", "Autenticação obrigatória.")
		return
	}

	costs, _, err := handler.repository.GetFixedCosts(ctx.Request.Context(), user.ID)
	if err != nil {
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao buscar custos fixos", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": newFixedCostsResponse(costs)})
}

func (handler *FixedCosts) Patch(ctx *gin.Context) {
	user, ok := AuthenticatedUser(ctx)
	if !ok {
		writeError(ctx, http.StatusUnauthorized, "unauthorized", "Autenticação obrigatória.")
		return
	}
	var request patchFixedCostsRequest
	if err := decodeStrictJSON(ctx, &request); err != nil {
		writeJSONDecodeError(ctx, err)
		return
	}

	patch, err := request.toDomainPatch(user.ID)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_body", err.Error())
		return
	}
	costs, err := handler.repository.UpsertFixedCosts(ctx.Request.Context(), patch)
	if err != nil {
		handler.logger.ErrorContext(ctx.Request.Context(), "falha ao salvar custos fixos", "error", err)
		writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": newFixedCostsResponse(costs)})
}

func decodeStrictJSON(ctx *gin.Context, destination any) error {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return errRequestBodyTooLarge
		}
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return errRequestBodyTooLarge
		}
		return errors.New("o corpo deve conter um único objeto JSON")
	}
	return nil
}

func (request patchFixedCostsRequest) toDomainPatch(userID int64) (domain.FixedCostsPatch, error) {
	patch := domain.FixedCostsPatch{UserID: userID}
	if request.Costs == nil {
		return patch, errors.New("Informe ao menos um custo para atualizar.")
	}

	patch.OABAnnualFeeCents = annualValue(request.Costs.OABAnnualFee)
	patch.DigitalCertificateCents = monthlyValue(request.Costs.DigitalCertificate)
	patch.AccountantCents = monthlyValue(request.Costs.Accountant)
	patch.LegalSoftwareCents = monthlyValue(request.Costs.LegalSoftware)
	patch.InternetCents = monthlyValue(request.Costs.Internet)
	patch.PhoneCents = monthlyValue(request.Costs.Phone)
	patch.RecurringTransportCents = monthlyValue(request.Costs.RecurringTransport)
	patch.CoworkingOrOfficeRentCents = monthlyValue(request.Costs.CoworkingOrOfficeRent)
	patch.ProfessionalDomainWebsiteEmailCents = monthlyValue(request.Costs.ProfessionalDomainWebsiteEmail)
	patch.MarketingCents = monthlyValue(request.Costs.Marketing)
	patch.OfficeSuppliesCents = monthlyValue(request.Costs.OfficeSupplies)
	patch.EquipmentAndDepreciationCents = monthlyValue(request.Costs.EquipmentAndDepreciation)
	patch.OtherCostsCents = monthlyValue(request.Costs.OtherCosts)

	values := []*int64{
		patch.OABAnnualFeeCents, patch.DigitalCertificateCents, patch.AccountantCents,
		patch.LegalSoftwareCents, patch.InternetCents, patch.PhoneCents,
		patch.RecurringTransportCents, patch.CoworkingOrOfficeRentCents,
		patch.ProfessionalDomainWebsiteEmailCents, patch.MarketingCents,
		patch.OfficeSuppliesCents, patch.EquipmentAndDepreciationCents,
		patch.OtherCostsCents,
	}
	hasValue := false
	for _, value := range values {
		if value == nil {
			continue
		}
		hasValue = true
		if *value < 0 {
			return patch, errors.New("Os valores dos custos não podem ser negativos.")
		}
	}
	if !hasValue {
		return patch, errors.New("Informe ao menos um custo para atualizar.")
	}
	return patch, nil
}

func annualValue(cost *annualCostRequest) *int64 {
	if cost == nil {
		return nil
	}
	return cost.AnnualAmountCents
}

func monthlyValue(cost *monthlyCostRequest) *int64 {
	if cost == nil {
		return nil
	}
	return cost.MonthlyAmountCents
}

func newFixedCostsResponse(costs domain.FixedCosts) fixedCostsResponse {
	return fixedCostsResponse{
		UserID: costs.UserID,
		Costs: fixedCostsResponseBody{
			OABAnnualFee: annualCostResponse{
				AnnualAmountCents:  costs.OABAnnualFeeCents,
				MonthlyAmountCents: domain.MonthlyAverageCents(costs.OABAnnualFeeCents),
			},
			DigitalCertificate:             monthlyCostResponse{costs.DigitalCertificateCents},
			Accountant:                     monthlyCostResponse{costs.AccountantCents},
			LegalSoftware:                  monthlyCostResponse{costs.LegalSoftwareCents},
			Internet:                       monthlyCostResponse{costs.InternetCents},
			Phone:                          monthlyCostResponse{costs.PhoneCents},
			RecurringTransport:             monthlyCostResponse{costs.RecurringTransportCents},
			CoworkingOrOfficeRent:          monthlyCostResponse{costs.CoworkingOrOfficeRentCents},
			ProfessionalDomainWebsiteEmail: monthlyCostResponse{costs.ProfessionalDomainWebsiteEmailCents},
			Marketing:                      monthlyCostResponse{costs.MarketingCents},
			OfficeSupplies:                 monthlyCostResponse{costs.OfficeSuppliesCents},
			EquipmentAndDepreciation:       monthlyCostResponse{costs.EquipmentAndDepreciationCents},
			OtherCosts:                     monthlyCostResponse{costs.OtherCostsCents},
		},
	}
}
