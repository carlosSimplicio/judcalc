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

func TestFeeCalculationRejectsInvalidBodiesBeforeRepositoryCalls(t *testing.T) {
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
		t.Run(test.body, func(t *testing.T) {
			services := &serviceRepositoryStub{exists: true}
			costs := &fixedCostsRepositoryStub{}
			response := performJSONRequest(t, &areaRepositoryStub{}, services, costs, http.MethodPost, "/api/v1/services/fee-calculation", test.body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var body responseError
			decodeResponse(t, response, &body)
			if body.Error.Code != test.code {
				t.Fatalf("code = %q, want %q", body.Error.Code, test.code)
			}
			if services.requestedID != 0 || costs.requestedUserID != 0 {
				t.Fatalf("repositories called: service id = %d, user id = %d", services.requestedID, costs.requestedUserID)
			}
		})
	}
}

func TestFeeCalculationMapsNotFoundAndRepositoryErrors(t *testing.T) {
	validBody := `{"service_id":1,"estimated_hours":10,"billable_hours_per_month":80,"complexity":"low","risk":"low"}`
	tests := []struct {
		name     string
		services *serviceRepositoryStub
		costs    *fixedCostsRepositoryStub
		status   int
		code     string
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
