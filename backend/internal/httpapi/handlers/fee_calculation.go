package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

const fixedCostsUnavailableMessage = "Cadastre custos fixos para calcular o custo sustentável e a estimativa técnica."

type FeeCalculation struct {
	services   domain.ServiceRepository
	fixedCosts domain.FixedCostsRepository
	logger     *slog.Logger
}

func NewFeeCalculation(services domain.ServiceRepository, fixedCosts domain.FixedCostsRepository, logger *slog.Logger) *FeeCalculation {
	if logger == nil {
		logger = slog.Default()
	}
	return &FeeCalculation{services: services, fixedCosts: fixedCosts, logger: logger}
}

type feeCalculationRequest struct {
	ServiceID             requiredServiceID `json:"service_id"`
	EstimatedHours        float64           `json:"estimated_hours"`
	BillableHoursPerMonth float64           `json:"billable_hours_per_month"`
	Complexity            string            `json:"complexity"`
	Risk                  string            `json:"risk"`
}

type requiredServiceID int64

func (id *requiredServiceID) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("service_id must be an integer")
	}
	var parsed int64
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	*id = requiredServiceID(parsed)
	return nil
}

type feeServiceResponse struct {
	ID     int64  `json:"id"`
	AreaID int64  `json:"area_id"`
	Name   string `json:"name"`
}

type oabReferenceResponse struct {
	AmountCents   *int64   `json:"amount_cents"`
	PercentageMin *float64 `json:"percentage_min"`
	PercentageMax *float64 `json:"percentage_max"`
}

type feeInputsResponse struct {
	EstimatedHours        float64 `json:"estimated_hours"`
	BillableHoursPerMonth float64 `json:"billable_hours_per_month"`
	Complexity            string  `json:"complexity"`
	ComplexityFactor      float64 `json:"complexity_factor"`
	Risk                  string  `json:"risk"`
	RiskFactor            float64 `json:"risk_factor"`
}

type feeValuesResponse struct {
	MonthlyFixedCostsCents      int64  `json:"monthly_fixed_costs_cents"`
	OperationalHourCostCents    *int64 `json:"operational_hour_cost_cents"`
	MinimumSustainableCostCents *int64 `json:"minimum_sustainable_cost_cents"`
	TechnicalEstimateCents      *int64 `json:"technical_estimate_cents"`
}

type feeWarningResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type feeCalculationResponse struct {
	Service      feeServiceResponse   `json:"service"`
	OABReference oabReferenceResponse `json:"oab_reference"`
	Inputs       feeInputsResponse    `json:"inputs"`
	Calculation  feeValuesResponse    `json:"calculation"`
	Warnings     []feeWarningResponse `json:"warnings"`
}

func (handler *FeeCalculation) Calculate(ctx *gin.Context) {
	user, ok := AuthenticatedUser(ctx)
	if !ok {
		writeError(ctx, http.StatusUnauthorized, "unauthorized", "Autenticação obrigatória.")
		return
	}

	var request feeCalculationRequest
	if err := decodeStrictJSON(ctx, &request); err != nil {
		writeJSONDecodeError(ctx, err)
		return
	}
	if request.ServiceID <= 0 {
		writeError(ctx, http.StatusBadRequest, "invalid_service_id", "O id do serviço é inválido.")
		return
	}

	input := domain.FeeCalculationInput{
		EstimatedHours:        request.EstimatedHours,
		BillableHoursPerMonth: request.BillableHoursPerMonth,
		Complexity:            domain.FeeLevel(request.Complexity),
		Risk:                  domain.FeeLevel(request.Risk),
	}
	if err := domain.ValidateFeeCalculationInput(input); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid_body", "Os dados para cálculo são inválidos.")
		return
	}

	service, exists, err := handler.services.GetService(ctx.Request.Context(), int64(request.ServiceID))
	if err != nil {
		handler.internalError(ctx, "falha ao buscar serviço para cálculo", err)
		return
	}
	if !exists {
		writeError(ctx, http.StatusNotFound, "service_not_found", "Serviço não encontrado.")
		return
	}

	costs, _, err := handler.fixedCosts.GetFixedCosts(ctx.Request.Context(), user.ID)
	if err != nil {
		handler.internalError(ctx, "falha ao buscar custos para cálculo", err)
		return
	}
	calculation, err := domain.CalculateFee(costs, input)
	if errors.Is(err, domain.ErrInvalidFeeCalculation) {
		writeError(ctx, http.StatusBadRequest, "invalid_body", "Os dados para cálculo são inválidos.")
		return
	}
	if err != nil {
		handler.internalError(ctx, "falha ao calcular honorários", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": newFeeCalculationResponse(service, request, calculation)})
}

func (handler *FeeCalculation) internalError(ctx *gin.Context, message string, err error) {
	handler.logger.ErrorContext(ctx.Request.Context(), message, "error", err)
	writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
}

func newFeeCalculationResponse(service domain.Service, request feeCalculationRequest, calculation domain.FeeCalculation) feeCalculationResponse {
	warnings := make([]feeWarningResponse, 0)
	if calculation.OperationalHourCostCents == nil {
		warnings = append(warnings, feeWarningResponse{
			Code:    "fixed_costs_unavailable",
			Message: fixedCostsUnavailableMessage,
		})
	}
	return feeCalculationResponse{
		Service: feeServiceResponse{ID: service.ID, AreaID: service.AreaID, Name: service.Name},
		OABReference: oabReferenceResponse{
			AmountCents: service.AmountCents, PercentageMin: service.PercentageMin, PercentageMax: service.PercentageMax,
		},
		Inputs: feeInputsResponse{
			EstimatedHours: request.EstimatedHours, BillableHoursPerMonth: request.BillableHoursPerMonth,
			Complexity: request.Complexity, ComplexityFactor: calculation.ComplexityFactor,
			Risk: request.Risk, RiskFactor: calculation.RiskFactor,
		},
		Calculation: feeValuesResponse{
			MonthlyFixedCostsCents:      calculation.MonthlyFixedCostsCents,
			OperationalHourCostCents:    calculation.OperationalHourCostCents,
			MinimumSustainableCostCents: calculation.MinimumSustainableCostCents,
			TechnicalEstimateCents:      calculation.TechnicalEstimateCents,
		},
		Warnings: warnings,
	}
}
