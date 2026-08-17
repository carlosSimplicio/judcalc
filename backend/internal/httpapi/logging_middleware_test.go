package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLoggerUsesInjectedLogger(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	router := NewRouter(&areaRepositoryStub{}, &serviceRepositoryStub{}, logger)

	request := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v; log = %s", err, output.String())
	}
	if entry["msg"] != "requisição HTTP" || entry["method"] != http.MethodGet || entry["path"] != "/unknown" {
		t.Fatalf("unexpected log entry: %#v", entry)
	}
	if entry["status"] != float64(http.StatusNotFound) {
		t.Fatalf("status = %#v", entry["status"])
	}
	if _, exists := entry["duration"]; !exists {
		t.Fatalf("duration missing: %#v", entry)
	}
	if _, exists := entry["client_ip"]; !exists {
		t.Fatalf("client_ip missing: %#v", entry)
	}
}
