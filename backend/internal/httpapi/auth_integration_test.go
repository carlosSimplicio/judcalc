package httpapi

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/carlosSimplicio/judcalc/backend/internal/auth"
	sqlitestorage "github.com/carlosSimplicio/judcalc/backend/internal/storage/sqlite"
)

func TestAuthenticationIntegrationPersistsSecureSessionsAndIsolatesCosts(t *testing.T) {
	path := filepath.Join(t.TempDir(), "integration.db")
	schema, err := os.ReadFile("../../../python/database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := seed.Exec(string(schema)); err != nil {
		seed.Close()
		t.Fatal(err)
	}
	seed.Close()

	database, err := sqlitestorage.OpenDatabase(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	authentication := auth.NewService(sqlitestorage.NewAuthRepository(database))
	router := NewRouter(
		sqlitestorage.NewAreaRepository(database),
		sqlitestorage.NewServiceRepository(database),
		sqlitestorage.NewFixedCostsRepository(database),
		authentication,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)

	first := signUpIntegrationUser(t, router, "first@example.com", "First")
	second := signUpIntegrationUser(t, router, "second@example.com", "Second")
	var passwordHash string
	if err := database.QueryRow("SELECT password_hash FROM users WHERE id = ?", first.Data.User.ID).Scan(&passwordHash); err != nil {
		t.Fatal(err)
	}
	if passwordHash == "password-123" || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password-123")) != nil {
		t.Fatal("database does not contain the expected bcrypt password hash")
	}
	rows, err := database.Query("SELECT token_hash FROM auth_tokens")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var storedHash string
		if err := rows.Scan(&storedHash); err != nil {
			t.Fatal(err)
		}
		if len(storedHash) != 64 || storedHash == first.Data.AccessToken || storedHash == second.Data.AccessToken {
			t.Fatalf("unsafe stored token: %q", storedHash)
		}
	}

	patch := integrationRequest(t, router, http.MethodPatch, "/api/v1/fixed-costs", `{"costs":{"internet":{"monthly_amount_cents":15000}}}`, first.Data.AccessToken)
	if patch.Code != http.StatusOK {
		t.Fatalf("patch status = %d, body = %s", patch.Code, patch.Body.String())
	}
	firstCosts := integrationRequest(t, router, http.MethodGet, "/api/v1/fixed-costs", "", first.Data.AccessToken)
	secondCosts := integrationRequest(t, router, http.MethodGet, "/api/v1/fixed-costs", "", second.Data.AccessToken)
	if integrationInternetCost(t, firstCosts) != 15_000 || integrationInternetCost(t, secondCosts) != 0 {
		t.Fatalf("fixed costs were not isolated: first=%s second=%s", firstCosts.Body.String(), secondCosts.Body.String())
	}

	signIn := integrationRequest(t, router, http.MethodPost, "/api/v1/auth/sign-in", `{"email":"FIRST@example.com","password":"password-123"}`, "")
	if signIn.Code != http.StatusOK {
		t.Fatalf("sign-in status = %d, body = %s", signIn.Code, signIn.Body.String())
	}
	var signedIn integrationSessionResponse
	decodeResponse(t, signIn, &signedIn)
	areas := integrationRequest(t, router, http.MethodGet, "/api/v1/areas", "", signedIn.Data.AccessToken)
	if areas.Code != http.StatusOK {
		t.Fatalf("new session was not accepted: status=%d body=%s", areas.Code, areas.Body.String())
	}
}

type integrationSessionResponse struct {
	Data struct {
		User struct {
			ID int64 `json:"id"`
		} `json:"user"`
		AccessToken string `json:"access_token"`
	} `json:"data"`
}

func signUpIntegrationUser(t *testing.T, router http.Handler, email, name string) integrationSessionResponse {
	t.Helper()
	body := `{"email":"` + email + `","name":"` + name + `","password":"password-123"}`
	response := integrationRequest(t, router, http.MethodPost, "/api/v1/auth/sign-up", body, "")
	if response.Code != http.StatusCreated {
		t.Fatalf("sign-up status = %d, body = %s", response.Code, response.Body.String())
	}
	var session integrationSessionResponse
	decodeResponse(t, response, &session)
	if session.Data.User.ID <= 0 || session.Data.AccessToken == "" {
		t.Fatalf("invalid sign-up response: %#v", session)
	}
	return session
}

func integrationRequest(t *testing.T, router http.Handler, method, target, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func integrationInternetCost(t *testing.T, response *httptest.ResponseRecorder) int64 {
	t.Helper()
	if response.Code != http.StatusOK {
		t.Fatalf("costs status = %d, body = %s", response.Code, response.Body.String())
	}
	var body struct {
		Data struct {
			Costs struct {
				Internet struct {
					MonthlyAmountCents int64 `json:"monthly_amount_cents"`
				} `json:"internet"`
			} `json:"costs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body.Data.Costs.Internet.MonthlyAmountCents
}
