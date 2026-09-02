# Cálculo de honorários por serviço Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Entregar um endpoint autenticado que apresente separadamente a referência da OAB, o custo mínimo sustentável e a estimativa técnica de um serviço.

**Architecture:** Um calculador puro no domínio consolida os custos mensais, valida horas e níveis, aplica fatores fixos e arredonda somente os valores expostos. O handler orquestra a busca do serviço e dos custos do usuário autenticado, converte a ausência de custos em resposta parcial com aviso e não persiste o cálculo.

**Tech Stack:** Go 1.24, Gin, `database/sql`, SQLite/FTS5, testes com `testing` e `httptest` da biblioteca padrão.

**Spec:** `docs/superpowers/specs/2026-09-01-calculo-honorarios-servico-design.md`

## Global Constraints

- A rota é `POST /api/v1/services/fee-calculation` e exige token Bearer.
- `service_id` é um inteiro positivo obrigatório no corpo JSON.
- O usuário vem exclusivamente da sessão autenticada.
- Horas aceitam números inteiros ou fracionários, sempre maiores que zero.
- Complexidade usa `low=1.00`, `medium=1.25`, `high=1.50`.
- Risco usa `low=1.00`, `medium=1.10`, `high=1.20`.
- A referência da OAB permanece separada e percentuais não são convertidos em dinheiro.
- Custos ausentes ou zerados produzem `200 OK`, resultados econômicos `null` e aviso `fixed_costs_unavailable`.
- Valores calculados usam precisão completa e são arredondados independentemente para centavos inteiros.
- Não persistir resultados, alterar o esquema SQLite nem modificar a seção de banco de dados do `AGENTS.md`.
- Não incluir remuneração desejada, tributos, custos variáveis, margem de segurança ou faixa sugerida.

---

### Task 1: Implementar o cálculo puro no domínio

**Files:**
- Create: `backend/internal/domain/fee_calculation.go`
- Create: `backend/internal/domain/fee_calculation_test.go`
- Modify: `backend/internal/domain/fixed_costs.go`
- Modify: `backend/internal/httpapi/handlers/fixed_costs.go:210-240`

**Interfaces:**
- Produces: `type FeeLevel string` e constantes `FeeLevelLow`, `FeeLevelMedium`, `FeeLevelHigh`
- Produces: `FeeCalculationInput`, `FeeCalculation`, `ErrInvalidFeeCalculation`
- Produces: `CalculateFee(FixedCosts, FeeCalculationInput) (FeeCalculation, error)`
- Produces: `MonthlyAverageCents(annualAmountCents int64) int64`
- Produces: `(FixedCosts).MonthlyTotalCents() int64`

- [ ] **Step 1: Escrever testes falhos para consolidação mensal, fatores, precisão e ausência de custos**

Criar `backend/internal/domain/fee_calculation_test.go`:

```go
package domain

import (
	"errors"
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
```

- [ ] **Step 2: Executar os testes e confirmar a falha esperada**

Run: `cd backend && go test ./internal/domain -run 'TestMonthlyTotal|TestCalculateFee'`

Expected: FAIL de compilação porque `CalculateFee`, os tipos e os métodos ainda não existem.

- [ ] **Step 3: Implementar tipos, consolidação e fórmulas no domínio**

Adicionar a `backend/internal/domain/fixed_costs.go`:

```go
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
```

Criar `backend/internal/domain/fee_calculation.go`:

```go
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
	MonthlyFixedCostsCents       int64
	OperationalHourCostCents     *int64
	MinimumSustainableCostCents  *int64
	TechnicalEstimateCents       *int64
	ComplexityFactor             float64
	RiskFactor                   float64
}

var complexityFactors = map[FeeLevel]float64{
	FeeLevelLow: 1, FeeLevelMedium: 1.25, FeeLevelHigh: 1.5,
}

var riskFactors = map[FeeLevel]float64{
	FeeLevelLow: 1, FeeLevelMedium: 1.1, FeeLevelHigh: 1.2,
}

func CalculateFee(costs FixedCosts, input FeeCalculationInput) (FeeCalculation, error) {
	complexityFactor, validComplexity := complexityFactors[input.Complexity]
	riskFactor, validRisk := riskFactors[input.Risk]
	if !positiveFinite(input.EstimatedHours) || !positiveFinite(input.BillableHoursPerMonth) ||
		!validComplexity || !validRisk {
		return FeeCalculation{}, ErrInvalidFeeCalculation
	}

	result := FeeCalculation{
		MonthlyFixedCostsCents: costs.MonthlyTotalCents(),
		ComplexityFactor: complexityFactor,
		RiskFactor: riskFactor,
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
```

Em `backend/internal/httpapi/handlers/fixed_costs.go`, trocar `monthlyAverage(costs.OABAnnualFeeCents)` por `domain.MonthlyAverageCents(costs.OABAnnualFeeCents)` e remover a função privada `monthlyAverage`.

- [ ] **Step 4: Executar testes focados e toda a suíte do backend**

Run: `cd backend && go test ./internal/domain ./internal/httpapi/handlers ./...`

Expected: PASS em todos os pacotes.

- [ ] **Step 5: Formatar e confirmar que não houve mudança de comportamento nos custos fixos**

Run: `cd backend && gofmt -w internal/domain/fixed_costs.go internal/domain/fee_calculation.go internal/domain/fee_calculation_test.go internal/httpapi/handlers/fixed_costs.go && go test ./internal/httpapi -run 'TestGetFixedCosts|TestPatchFixedCosts'`

Expected: PASS nos testes existentes de custos fixos.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/fixed_costs.go backend/internal/domain/fee_calculation.go backend/internal/domain/fee_calculation_test.go backend/internal/httpapi/handlers/fixed_costs.go
git commit -m "feat: add service fee calculator"
```

---

### Task 2: Buscar serviço por identificador

**Files:**
- Modify: `backend/internal/domain/repositories.go:9-11`
- Modify: `backend/internal/storage/sqlite/service_repository.go`
- Modify: `backend/internal/storage/sqlite/service_repository_test.go`
- Modify: `backend/internal/httpapi/router_test.go:33-36,64-66`

**Interfaces:**
- Consumes: `domain.Service`
- Produces: `ServiceRepository.GetService(context.Context, int64) (Service, bool, error)`
- Produces no stub HTTP: campos `service domain.Service`, `exists bool`, `getErr error`, `requestedID int64`

- [ ] **Step 1: Escrever teste falho para serviço existente, campos nulos e ausência**

Adicionar a `backend/internal/storage/sqlite/service_repository_test.go`:

```go
func TestServiceRepositoryGetsServiceByIDAndReportsMissing(t *testing.T) {
	_, repository, closeDatabase := newTestRepositories(t)
	defer closeDatabase()
	ctx := context.Background()

	service, exists, err := repository.GetService(ctx, 1)
	if err != nil || !exists {
		t.Fatalf("service = %#v, exists = %v, err = %v", service, exists, err)
	}
	if service.ID != 1 || service.AreaID != 3 || service.Name != "Ação previdenciária comum" ||
		service.AmountCents == nil || *service.AmountCents != 10000 ||
		service.PercentageMin == nil || *service.PercentageMin != 10 ||
		service.PercentageMax == nil || *service.PercentageMax != 20 {
		t.Fatalf("unexpected service: %#v", service)
	}

	nullable, exists, err := repository.GetService(ctx, 2)
	if err != nil || !exists || nullable.AmountCents != nil ||
		nullable.PercentageMin != nil || nullable.PercentageMax != nil {
		t.Fatalf("nullable service = %#v, exists = %v, err = %v", nullable, exists, err)
	}

	missing, exists, err := repository.GetService(ctx, 999)
	if err != nil || exists || missing != (domain.Service{}) {
		t.Fatalf("missing = %#v, exists = %v, err = %v", missing, exists, err)
	}
}
```

- [ ] **Step 2: Executar o teste e confirmar a falha esperada**

Run: `cd backend && go test ./internal/storage/sqlite -run TestServiceRepositoryGetsServiceByIDAndReportsMissing`

Expected: FAIL de compilação porque `GetService` ainda não existe.

- [ ] **Step 3: Adicionar a interface e implementar a consulta reutilizando um scanner**

Alterar `domain.ServiceRepository`:

```go
type ServiceRepository interface {
	ListServices(context.Context, ListOptions) (ListResult[Service], error)
	GetService(context.Context, int64) (Service, bool, error)
}
```

Em `backend/internal/storage/sqlite/service_repository.go`, adicionar:

```go
type serviceRowScanner interface {
	Scan(...any) error
}

func (repository *ServiceRepository) GetService(ctx context.Context, serviceID int64) (domain.Service, bool, error) {
	row := repository.database.QueryRowContext(ctx, `
		SELECT id, area_id, name, amount_cents, percentage_min, percentage_max
		FROM services WHERE id = ?`, serviceID)
	service, err := scanService(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Service{}, false, nil
	}
	if err != nil {
		return domain.Service{}, false, fmt.Errorf("buscar serviço: %w", err)
	}
	return service, true, nil
}

func scanService(row serviceRowScanner) (domain.Service, error) {
	var service domain.Service
	var amount sql.NullInt64
	var minimum, maximum sql.NullFloat64
	err := row.Scan(&service.ID, &service.AreaID, &service.Name, &amount, &minimum, &maximum)
	service.AmountCents = nullableInt64(amount)
	service.PercentageMin = nullableFloat64(minimum)
	service.PercentageMax = nullableFloat64(maximum)
	return service, err
}
```

Adicionar `"errors"` aos imports. No laço de `ListServices`, substituir o bloco local de `Scan` por:

```go
service, err := scanService(rows)
if err != nil {
	return result, fmt.Errorf("ler serviço: %w", err)
}
result.Items = append(result.Items, service)
```

- [ ] **Step 4: Atualizar o stub HTTP para satisfazer a interface ampliada**

Em `backend/internal/httpapi/router_test.go`, definir o stub como:

```go
type serviceRepositoryStub struct {
	result      domain.ListResult[domain.Service]
	err         error
	service     domain.Service
	exists      bool
	getErr      error
	requestedID int64
}

func (stub *serviceRepositoryStub) GetService(_ context.Context, serviceID int64) (domain.Service, bool, error) {
	stub.requestedID = serviceID
	return stub.service, stub.exists, stub.getErr
}
```

- [ ] **Step 5: Formatar e executar testes do repositório e da API**

Run: `cd backend && gofmt -w internal/domain/repositories.go internal/storage/sqlite/service_repository.go internal/storage/sqlite/service_repository_test.go internal/httpapi/router_test.go && go test ./internal/storage/sqlite ./internal/httpapi`

Expected: PASS nos dois pacotes, incluindo listagem, FTS e autenticação existentes.

- [ ] **Step 6: Commit**

```bash
git add backend/internal/domain/repositories.go backend/internal/storage/sqlite/service_repository.go backend/internal/storage/sqlite/service_repository_test.go backend/internal/httpapi/router_test.go
git commit -m "feat: look up services by id"
```

---

### Task 3: Expor o endpoint autenticado de cálculo

**Files:**
- Create: `backend/internal/httpapi/handlers/fee_calculation.go`
- Create: `backend/internal/httpapi/fee_calculation_test.go`
- Modify: `backend/internal/httpapi/router.go:25-37`
- Modify: `backend/internal/httpapi/router_test.go:38-43,68-72`

**Interfaces:**
- Consumes: `domain.ServiceRepository.GetService`, `domain.FixedCostsRepository.GetFixedCosts`, `domain.CalculateFee`
- Produces: `handlers.NewFeeCalculation(domain.ServiceRepository, domain.FixedCostsRepository, *slog.Logger) *FeeCalculation`
- Produces: `(*FeeCalculation).Calculate(*gin.Context)`
- Produces: resposta `data` com `service`, `oab_reference`, `inputs`, `calculation` e `warnings`

- [ ] **Step 1: Escrever o teste falho do caminho completo com valores e fatores**

Criar `backend/internal/httpapi/fee_calculation_test.go` com:

```go
package httpapi

import (
	"errors"
	"net/http"
	"testing"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

func TestFeeCalculationReturnsOABReferenceAndEconomicValues(t *testing.T) {
	amount := int64(150000)
	minimum, maximum := 10.0, 20.0
	services := &serviceRepositoryStub{service: domain.Service{
		ID: 42, AreaID: 3, Name: "Serviço de exemplo", AmountCents: &amount,
		PercentageMin: &minimum, PercentageMax: &maximum,
	}, exists: true}
	costs := &fixedCostsRepositoryStub{result: domain.FixedCosts{
		UserID: 123, OABAnnualFeeCents: 120000, InternetCents: 470000,
	}, exists: true}
	response := performJSONRequest(t, &areaRepositoryStub{}, services, costs, http.MethodPost,
		"/api/v1/services/fee-calculation",
		`{"service_id":42,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"medium","risk":"high"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if services.requestedID != 42 || costs.requestedUserID != 123 {
		t.Fatalf("service id = %d, user id = %d", services.requestedID, costs.requestedUserID)
	}
	var raw map[string]any
	decodeResponse(t, response, &raw)
	data := raw["data"].(map[string]any)
	service := data["service"].(map[string]any)
	calculation := data["calculation"].(map[string]any)
	if service["id"] != float64(42) || service["area_id"] != float64(3) || service["name"] != "Serviço de exemplo" {
		t.Fatalf("service = %#v", service)
	}
	if calculation["monthly_fixed_costs_cents"] != float64(480000) ||
		calculation["operational_hour_cost_cents"] != float64(6000) ||
		calculation["minimum_sustainable_cost_cents"] != float64(60000) ||
		calculation["technical_estimate_cents"] != float64(90000) {
		t.Fatalf("unexpected calculation: %#v", calculation)
	}
	oab := data["oab_reference"].(map[string]any)
	inputs := data["inputs"].(map[string]any)
	if oab["amount_cents"] != float64(150000) || oab["percentage_min"] != 10.0 || oab["percentage_max"] != 20.0 ||
		inputs["estimated_hours"] != 10.0 || inputs["billable_hours_per_month"] != 80.0 ||
		inputs["complexity"] != "medium" || inputs["complexity_factor"] != 1.25 ||
		inputs["risk"] != "high" || inputs["risk_factor"] != 1.2 {
		t.Fatalf("oab = %#v, inputs = %#v", oab, inputs)
	}
	if warnings := data["warnings"].([]any); len(warnings) != 0 {
		t.Fatalf("warnings = %#v", warnings)
	}
}
```

Em `fixedCostsRepositoryStub`, registrar o usuário solicitado:

```go
type fixedCostsRepositoryStub struct {
	result          domain.FixedCosts
	exists          bool
	err             error
	patch           domain.FixedCostsPatch
	requestedUserID int64
}

func (stub *fixedCostsRepositoryStub) GetFixedCosts(_ context.Context, userID int64) (domain.FixedCosts, bool, error) {
	stub.requestedUserID = userID
	if stub.result.UserID == 0 {
		stub.result.UserID = userID
	}
	return stub.result, stub.exists, stub.err
}
```

- [ ] **Step 2: Acrescentar testes falhos para resposta parcial, percentuais, validações e erros**

Adicionar ao mesmo arquivo:

```go
func TestFeeCalculationReturnsOABPercentageAndWarningWithoutFixedCosts(t *testing.T) {
	minimum, maximum := 10.0, 20.0
	services := &serviceRepositoryStub{service: domain.Service{
		ID: 7, AreaID: 2, Name: "Serviço percentual",
		PercentageMin: &minimum, PercentageMax: &maximum,
	}, exists: true}
	response := performJSONRequest(t, &areaRepositoryStub{}, services, &fixedCostsRepositoryStub{}, http.MethodPost,
		"/api/v1/services/fee-calculation",
		`{"service_id":7,"estimated_hours":5.5,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body map[string]any
	decodeResponse(t, response, &body)
	data := body["data"].(map[string]any)
	oab := data["oab_reference"].(map[string]any)
	calculation := data["calculation"].(map[string]any)
	warnings := data["warnings"].([]any)
	if oab["amount_cents"] != nil || oab["percentage_min"] != 10.0 || oab["percentage_max"] != 20.0 {
		t.Fatalf("oab = %#v", oab)
	}
	if calculation["monthly_fixed_costs_cents"] != float64(0) ||
		calculation["operational_hour_cost_cents"] != nil ||
		calculation["minimum_sustainable_cost_cents"] != nil ||
		calculation["technical_estimate_cents"] != nil {
		t.Fatalf("calculation = %#v", calculation)
	}
	if len(warnings) != 1 || warnings[0].(map[string]any)["code"] != "fixed_costs_unavailable" {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func TestFeeCalculationRejectsInvalidBodies(t *testing.T) {
	tests := []struct{ body, code string }{
		{`{}`, "invalid_service_id"},
		{`{"service_id":0,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`, "invalid_service_id"},
		{`{"service_id":-1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`, "invalid_service_id"},
		{`{"service_id":"1","estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`, "invalid_body"},
		{`{`, "invalid_body"},
		{`{"service_id":1,"estimated_hours":0,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`, "invalid_body"},
		{`{"service_id":1,"estimated_hours":10,"billable_hours_per_month":0,"complexity":"low","risk":"low"}`, "invalid_body"},
		{`{"service_id":1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"invalid","risk":"low"}`, "invalid_body"},
		{`{"service_id":1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"invalid"}`, "invalid_body"},
		{`{"service_id":1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low","extra":true}`, "invalid_body"},
	}
	for _, test := range tests {
		response := performJSONRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{exists: true}, &fixedCostsRepositoryStub{}, http.MethodPost, "/api/v1/services/fee-calculation", test.body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("request %s: status = %d, body = %s", test.body, response.Code, response.Body.String())
		}
		var body responseError
		decodeResponse(t, response, &body)
		if body.Error.Code != test.code {
			t.Fatalf("request %s: code = %q, want %q", test.body, body.Error.Code, test.code)
		}
	}
}

func TestFeeCalculationMapsNotFoundAndRepositoryErrors(t *testing.T) {
	validBody := `{"service_id":1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`
	tests := []struct {
		name string
		services *serviceRepositoryStub
		costs *fixedCostsRepositoryStub
		status int
		code string
	}{
		{"missing", &serviceRepositoryStub{}, &fixedCostsRepositoryStub{}, http.StatusNotFound, "service_not_found"},
		{"service error", &serviceRepositoryStub{getErr: errors.New("SELECT secret")}, &fixedCostsRepositoryStub{}, http.StatusInternalServerError, "internal_error"},
		{"cost error", &serviceRepositoryStub{service: domain.Service{ID: 1}, exists: true}, &fixedCostsRepositoryStub{err: errors.New("SELECT secret")}, http.StatusInternalServerError, "internal_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performJSONRequest(t, &areaRepositoryStub{}, test.services, test.costs, http.MethodPost, "/api/v1/services/fee-calculation", validBody)
			if response.Code != test.status || contains(response.Body.String(), "secret") {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var body responseError
			decodeResponse(t, response, &body)
			if body.Error.Code != test.code {
				t.Fatalf("code = %q, want %q", body.Error.Code, test.code)
			}
		})
	}
}
```

- [ ] **Step 3: Executar os testes e confirmar que a rota ainda não existe**

Run: `cd backend && go test ./internal/httpapi -run FeeCalculation`

Expected: FAIL; as requisições recebem `404` porque o handler e a rota ainda não existem.

- [ ] **Step 4: Implementar o handler e seus DTOs**

Criar `backend/internal/httpapi/handlers/fee_calculation.go` com estas estruturas e fluxo:

```go
package handlers

import (
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
	if logger == nil { logger = slog.Default() }
	return &FeeCalculation{services: services, fixedCosts: fixedCosts, logger: logger}
}

type feeCalculationRequest struct {
	ServiceID             int64   `json:"service_id"`
	EstimatedHours        float64 `json:"estimated_hours"`
	BillableHoursPerMonth float64 `json:"billable_hours_per_month"`
	Complexity            string  `json:"complexity"`
	Risk                  string  `json:"risk"`
}

type feeServiceResponse struct {
	ID int64 `json:"id"`
	AreaID int64 `json:"area_id"`
	Name string `json:"name"`
}

type oabReferenceResponse struct {
	AmountCents *int64 `json:"amount_cents"`
	PercentageMin *float64 `json:"percentage_min"`
	PercentageMax *float64 `json:"percentage_max"`
}

type feeInputsResponse struct {
	EstimatedHours float64 `json:"estimated_hours"`
	BillableHoursPerMonth float64 `json:"billable_hours_per_month"`
	Complexity string `json:"complexity"`
	ComplexityFactor float64 `json:"complexity_factor"`
	Risk string `json:"risk"`
	RiskFactor float64 `json:"risk_factor"`
}

type feeValuesResponse struct {
	MonthlyFixedCostsCents int64 `json:"monthly_fixed_costs_cents"`
	OperationalHourCostCents *int64 `json:"operational_hour_cost_cents"`
	MinimumSustainableCostCents *int64 `json:"minimum_sustainable_cost_cents"`
	TechnicalEstimateCents *int64 `json:"technical_estimate_cents"`
}

type feeWarningResponse struct { Code string `json:"code"`; Message string `json:"message"` }

type feeCalculationResponse struct {
	Service feeServiceResponse `json:"service"`
	OABReference oabReferenceResponse `json:"oab_reference"`
	Inputs feeInputsResponse `json:"inputs"`
	Calculation feeValuesResponse `json:"calculation"`
	Warnings []feeWarningResponse `json:"warnings"`
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
		EstimatedHours: request.EstimatedHours,
		BillableHoursPerMonth: request.BillableHoursPerMonth,
		Complexity: domain.FeeLevel(request.Complexity), Risk: domain.FeeLevel(request.Risk),
	}
	service, exists, err := handler.services.GetService(ctx.Request.Context(), request.ServiceID)
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
```

Completar o mesmo arquivo com:

```go
func (handler *FeeCalculation) internalError(ctx *gin.Context, message string, err error) {
	handler.logger.ErrorContext(ctx.Request.Context(), message, "error", err)
	writeError(ctx, http.StatusInternalServerError, "internal_error", "Não foi possível processar a solicitação.")
}

func newFeeCalculationResponse(service domain.Service, request feeCalculationRequest, calculation domain.FeeCalculation) feeCalculationResponse {
	warnings := make([]feeWarningResponse, 0)
	if calculation.OperationalHourCostCents == nil {
		warnings = append(warnings, feeWarningResponse{Code: "fixed_costs_unavailable", Message: fixedCostsUnavailableMessage})
	}
	return feeCalculationResponse{
		Service: feeServiceResponse{ID: service.ID, AreaID: service.AreaID, Name: service.Name},
		OABReference: oabReferenceResponse{AmountCents: service.AmountCents, PercentageMin: service.PercentageMin, PercentageMax: service.PercentageMax},
		Inputs: feeInputsResponse{
			EstimatedHours: request.EstimatedHours, BillableHoursPerMonth: request.BillableHoursPerMonth,
			Complexity: request.Complexity, ComplexityFactor: calculation.ComplexityFactor,
			Risk: request.Risk, RiskFactor: calculation.RiskFactor,
		},
		Calculation: feeValuesResponse{
			MonthlyFixedCostsCents: calculation.MonthlyFixedCostsCents,
			OperationalHourCostCents: calculation.OperationalHourCostCents,
			MinimumSustainableCostCents: calculation.MinimumSustainableCostCents,
			TechnicalEstimateCents: calculation.TechnicalEstimateCents,
		},
		Warnings: warnings,
	}
}
```

- [ ] **Step 5: Registrar a rota no grupo protegido**

Em `backend/internal/httpapi/router.go`, criar o handler junto aos demais e registrar a rota:

```go
feeCalculationHandler := handlers.NewFeeCalculation(services, fixedCosts, logger)
```

```go
protected.POST("/services/fee-calculation", feeCalculationHandler.Calculate)
```

- [ ] **Step 6: Formatar e executar os testes HTTP focados**

Run: `cd backend && gofmt -w internal/httpapi/handlers/fee_calculation.go internal/httpapi/fee_calculation_test.go internal/httpapi/router.go internal/httpapi/router_test.go && go test ./internal/httpapi -run FeeCalculation`

Expected: PASS nos testes de cálculo, resposta parcial, validação, `404` e falhas internas.

- [ ] **Step 7: Executar toda a suíte do backend**

Run: `cd backend && go test ./...`

Expected: PASS em todos os pacotes.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/httpapi/handlers/fee_calculation.go backend/internal/httpapi/fee_calculation_test.go backend/internal/httpapi/router.go backend/internal/httpapi/router_test.go
git commit -m "feat: expose service fee calculation endpoint"
```

---

### Task 4: Documentar o endpoint e concluir a verificação

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: contrato HTTP entregue na Task 3
- Produces: documentação pública da rota, corpo, valores calculados e resposta parcial

- [ ] **Step 1: Documentar a rota e um exemplo completo**

Adicionar `POST /api/v1/services/fee-calculation` à lista de endpoints do `README.md`. Após a seção de custos fixos, acrescentar:

````markdown
### Cálculo de honorários

O cálculo usa os custos fixos do usuário autenticado e mantém a referência da
OAB separada dos valores econômicos. Envie horas positivas e os níveis `low`,
`medium` ou `high`:

```json
{
  "service_id": 42,
  "estimated_hours": 10,
  "billable_hours_per_month": 80,
  "complexity": "medium",
  "risk": "high"
}
```

```text
POST /api/v1/services/fee-calculation
```

Complexidade usa fatores `1.00`, `1.25` e `1.50`; risco usa `1.00`, `1.10` e
`1.20`. A resposta apresenta a referência da OAB, os fatores usados, o custo
operacional por hora, o custo mínimo sustentável e a estimativa técnica.

Quando não há custos fixos utilizáveis, a referência da OAB continua presente,
os resultados econômicos são `null` e `warnings` contém
`fixed_costs_unavailable`.
````

- [ ] **Step 2: Verificar formatação e diferenças da documentação**

Run: `git diff --check -- README.md && git diff -- README.md`

Expected: nenhuma advertência de whitespace; diff limitado à lista de endpoints e à nova seção.

- [ ] **Step 3: Executar a verificação completa do backend**

Run: `cd backend && go test ./... && go vet ./...`

Expected: PASS em todos os testes e nenhuma advertência do `go vet`.

- [ ] **Step 4: Confirmar que banco e AGENTS.md permaneceram fora da implementação**

Run: `git diff --name-only HEAD~3..HEAD`

Expected: nenhuma linha para `python/database/schema.sql`, `python/scripts/init_database.py`, `python/data/oab-sp.json` ou `AGENTS.md`.

- [ ] **Step 5: Commit**

```bash
git add README.md
git commit -m "docs: document service fee calculation"
```
