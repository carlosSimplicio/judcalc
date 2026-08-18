package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/carlosSimplicio/judcalc/backend/internal/domain"
)

type areaRepositoryStub struct {
	result  domain.ListResult[domain.Area]
	err     error
	options domain.ListOptions
}

func (stub *areaRepositoryStub) ListAreas(_ context.Context, options domain.ListOptions) (domain.ListResult[domain.Area], error) {
	stub.options = options
	return stub.result, stub.err
}

type serviceRepositoryStub struct {
	result domain.ListResult[domain.Service]
	err    error
}

type fixedCostsRepositoryStub struct {
	result domain.FixedCosts
	exists bool
	err    error
	patch  domain.FixedCostsPatch
}

type responseMetadata struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

type responseError struct {
	Error struct {
		Code string `json:"code"`
	} `json:"error"`
}

func (stub *serviceRepositoryStub) ListServices(context.Context, domain.ListOptions) (domain.ListResult[domain.Service], error) {
	return stub.result, stub.err
}

func (stub *fixedCostsRepositoryStub) GetFixedCosts(_ context.Context, userID string) (domain.FixedCosts, bool, error) {
	if stub.result.UserID == "" {
		stub.result.UserID = userID
	}
	return stub.result, stub.exists, stub.err
}

func (stub *fixedCostsRepositoryStub) UpsertFixedCosts(_ context.Context, patch domain.FixedCostsPatch) (domain.FixedCosts, error) {
	stub.patch = patch
	return stub.result, stub.err
}

func TestListAreasUsesDefaultsAndReturnsMetadata(t *testing.T) {
	areas := &areaRepositoryStub{result: domain.ListResult[domain.Area]{
		Items: []domain.Area{{ID: 1, Name: "Família"}},
		Total: 21,
	}}
	response := performRequest(t, areas, &serviceRepositoryStub{}, "/api/v1/areas?q=%20familia%20")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	expectedOptions := domain.ListOptions{Page: 1, PageSize: 20, Query: "familia"}
	if !reflect.DeepEqual(areas.options, expectedOptions) {
		t.Fatalf("options = %#v, want %#v", areas.options, expectedOptions)
	}

	var body struct {
		Data []domain.Area    `json:"data"`
		Meta responseMetadata `json:"meta"`
	}
	decodeResponse(t, response, &body)
	if len(body.Data) != 1 || body.Data[0].Name != "Família" {
		t.Fatalf("data = %#v", body.Data)
	}
	wantMeta := responseMetadata{Page: 1, PageSize: 20, Total: 21, TotalPages: 2}
	if body.Meta != wantMeta {
		t.Fatalf("meta = %#v, want %#v", body.Meta, wantMeta)
	}
}

func TestListServicesPreservesNullableFields(t *testing.T) {
	amount := int64(12345)
	services := &serviceRepositoryStub{result: domain.ListResult[domain.Service]{
		Items: []domain.Service{
			{ID: 1, AreaID: 2, Name: "Sem valor"},
			{ID: 2, AreaID: 2, Name: "Com valor", AmountCents: &amount},
		},
		Total: 2,
	}}
	response := performRequest(t, &areaRepositoryStub{}, services, "/api/v1/services?page=2&page_size=1")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body map[string]any
	decodeResponse(t, response, &body)
	data := body["data"].([]any)
	first := data[0].(map[string]any)
	if first["amount_cents"] != nil || first["percentage_min"] != nil || first["percentage_max"] != nil {
		t.Fatalf("nullable fields were not null: %#v", first)
	}
	if _, exists := first["area_id"]; !exists {
		t.Fatalf("area_id missing: %#v", first)
	}
}

func TestListRejectsInvalidPagination(t *testing.T) {
	tests := []string{
		"?page=0",
		"?page=text",
		"?page_size=0",
		"?page_size=101",
		"?page=9223372036854775807&page_size=100",
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			response := performRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{}, "/api/v1/areas"+query)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var body responseError
			decodeResponse(t, response, &body)
			if body.Error.Code != "invalid_query" {
				t.Fatalf("error code = %q", body.Error.Code)
			}
		})
	}
}

func TestListReturnsEmptyArrayAndZeroPages(t *testing.T) {
	areas := &areaRepositoryStub{result: domain.ListResult[domain.Area]{Items: []domain.Area{}, Total: 0}}
	response := performRequest(t, areas, &serviceRepositoryStub{}, "/api/v1/areas?page=20")

	var body map[string]any
	decodeResponse(t, response, &body)
	if data, ok := body["data"].([]any); !ok || len(data) != 0 {
		t.Fatalf("data should be an empty array: %#v", body["data"])
	}
	meta := body["meta"].(map[string]any)
	if meta["total_pages"] != float64(0) {
		t.Fatalf("total_pages = %#v", meta["total_pages"])
	}
}

func TestInternalErrorsDoNotLeakDetails(t *testing.T) {
	areas := &areaRepositoryStub{err: errors.New("SELECT secret FROM private_table")}
	response := performRequest(t, areas, &serviceRepositoryStub{}, "/api/v1/areas")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body == "" || contains(body, "secret") || contains(body, "SELECT") {
		t.Fatalf("unsafe response body: %s", body)
	}
}

func TestGetFixedCostsReturnsZerosWhenUserHasNoRecord(t *testing.T) {
	response := performRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{}, "/api/v1/fixed-costs/user-123")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	var body map[string]any
	decodeResponse(t, response, &body)
	data := body["data"].(map[string]any)
	if data["user_id"] != "user-123" {
		t.Fatalf("user_id = %#v", data["user_id"])
	}
	costs := data["costs"].(map[string]any)
	if len(costs) != 13 {
		t.Fatalf("cost categories = %d, want 13", len(costs))
	}
	oab := costs["oab_annual_fee"].(map[string]any)
	if oab["annual_amount_cents"] != float64(0) || oab["monthly_amount_cents"] != float64(0) {
		t.Fatalf("unexpected OAB cost: %#v", oab)
	}
}

func TestPatchFixedCostsUsesPartialValuesAndReturnsMonthlyOABAverage(t *testing.T) {
	annual := int64(120006)
	internet := int64(15000)
	repository := &fixedCostsRepositoryStub{result: domain.FixedCosts{
		UserID: "user-123", OABAnnualFeeCents: annual, InternetCents: internet,
	}}
	body := `{"user_id":" user-123 ","costs":{"oab_annual_fee":{"annual_amount_cents":120006},"internet":{"monthly_amount_cents":15000}}}`
	response := performJSONRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{}, repository, http.MethodPatch, "/api/v1/fixed-costs", body)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repository.patch.UserID != "user-123" || repository.patch.OABAnnualFeeCents == nil || *repository.patch.OABAnnualFeeCents != annual {
		t.Fatalf("unexpected patch: %#v", repository.patch)
	}
	if repository.patch.PhoneCents != nil {
		t.Fatalf("omitted phone should remain nil: %#v", repository.patch.PhoneCents)
	}
	var responseBody map[string]any
	decodeResponse(t, response, &responseBody)
	data := responseBody["data"].(map[string]any)
	costs := data["costs"].(map[string]any)
	oab := costs["oab_annual_fee"].(map[string]any)
	if oab["monthly_amount_cents"] != float64(10001) {
		t.Fatalf("monthly OAB average = %#v", oab["monthly_amount_cents"])
	}
}

func TestPatchFixedCostsRejectsInvalidBodies(t *testing.T) {
	tests := []string{
		`{`,
		`{"user_id":"","costs":{"internet":{"monthly_amount_cents":1}}}`,
		`{"user_id":"user-1","costs":{}}`,
		`{"user_id":"user-1","costs":{"internet":{}}}`,
		`{"user_id":"user-1","costs":{"internet":{"monthly_amount_cents":-1}}}`,
		`{"user_id":"user-1","costs":{"unknown":{"monthly_amount_cents":1}}}`,
		`{"user_id":"user-1","costs":{"internet":{"unknown":1}}}`,
	}
	for _, body := range tests {
		t.Run(body, func(t *testing.T) {
			response := performJSONRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{}, &fixedCostsRepositoryStub{}, http.MethodPatch, "/api/v1/fixed-costs", body)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
			var responseBody responseError
			decodeResponse(t, response, &responseBody)
			if responseBody.Error.Code != "invalid_body" {
				t.Fatalf("error code = %q", responseBody.Error.Code)
			}
		})
	}
}

func TestFixedCostsInternalErrorsDoNotLeakDetails(t *testing.T) {
	repository := &fixedCostsRepositoryStub{err: errors.New("UPDATE private_costs SET secret = 1")}
	response := performJSONRequest(t, &areaRepositoryStub{}, &serviceRepositoryStub{}, repository, http.MethodGet, "/api/v1/fixed-costs/user-1", "")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.Code)
	}
	if body := response.Body.String(); body == "" || contains(body, "secret") || contains(body, "UPDATE") {
		t.Fatalf("unsafe response body: %s", body)
	}
}

func performRequest(t *testing.T, areas domain.AreaRepository, services domain.ServiceRepository, target string) *httptest.ResponseRecorder {
	t.Helper()
	return performJSONRequest(t, areas, services, &fixedCostsRepositoryStub{}, http.MethodGet, target, "")
}

func performJSONRequest(t *testing.T, areas domain.AreaRepository, services domain.ServiceRepository, fixedCosts domain.FixedCostsRepository, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(areas, services, fixedCosts, logger)
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.Unmarshal(response.Body.Bytes(), destination); err != nil {
		t.Fatalf("decode response: %v; body = %s", err, response.Body.String())
	}
}

func contains(value, fragment string) bool {
	for index := 0; index+len(fragment) <= len(value); index++ {
		if value[index:index+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
