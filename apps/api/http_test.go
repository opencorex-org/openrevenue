package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	app "github.com/opencorex-org/openrevenue/internal/administration/application"
)

type unhealthyDependencies struct{}

func (unhealthyDependencies) Check(context.Context) Readiness {
	return Readiness{Database: true, Migrations: false, Broker: false}
}

func authorizedRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer synthetic")
	request.Header.Set("X-Tenant-ID", "revenue")
	request.Header.Set("X-Jurisdiction-Code", "LK")
	request.Header.Set("X-Actor-ID", "officer")
	request.Header.Set("X-Correlation-ID", "api-registration-test")
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestOperationalAndAuthenticationBoundaries(t *testing.T) {
	router := Router(app.New(nil))
	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}
	for header, expected := range map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"Strict-Transport-Security": "max-age=63072000; includeSubDomains",
		"Permissions-Policy":        "camera=(), geolocation=(), microphone=(), payment=(), usb=()",
		"Cache-Control":             "no-store",
	} {
		if health.Header().Get(header) != expected {
			t.Fatalf("secure header %s = %q", header, health.Header().Get(header))
		}
	}

	protected := httptest.NewRecorder()
	router.ServeHTTP(protected, httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-events", nil))
	if protected.Code != http.StatusUnauthorized {
		t.Fatalf("protected status = %d", protected.Code)
	}

	missingScopeRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-events", nil)
	missingScopeRequest.Header.Set("Authorization", "Bearer synthetic")
	missingScope := httptest.NewRecorder()
	router.ServeHTTP(missingScope, missingScopeRequest)
	if missingScope.Code != http.StatusBadRequest {
		t.Fatalf("missing scope status = %d", missingScope.Code)
	}

	validRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit-events", nil)
	validRequest.Header.Set("Authorization", "Bearer synthetic")
	validRequest.Header.Set("X-Tenant-ID", "revenue")
	validRequest.Header.Set("X-Jurisdiction-Code", "LK")
	validRequest.Header.Set("X-Actor-ID", "auditor")
	validRequest.Header.Set("X-Correlation-ID", "api-test")
	valid := httptest.NewRecorder()
	router.ServeHTTP(valid, validRequest)
	if valid.Code != http.StatusOK {
		t.Fatalf("valid context status = %d body = %s", valid.Code, valid.Body)
	}

	metrics := httptest.NewRecorder()
	router.ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(metrics.Body.String(), "openrevenue_domain_context_rejections_total 1") {
		t.Fatalf("domain-context metric missing: %s", metrics.Body)
	}
}

func TestReadinessDistinguishesDependencyAndMigrationFailure(t *testing.T) {
	router := RouterWithReadiness(app.New(nil), unhealthyDependencies{})
	health := httptest.NewRecorder()
	router.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("liveness must remain healthy, got %d", health.Code)
	}
	ready := httptest.NewRecorder()
	router.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if ready.Code != http.StatusServiceUnavailable ||
		!strings.Contains(ready.Body.String(), `"migrations":false`) ||
		!strings.Contains(ready.Body.String(), `"broker":false`) {
		t.Fatalf("unexpected readiness response: %d %s", ready.Code, ready.Body)
	}
}

func TestTaxpayerRegistrationHTTPVerticalSlice(t *testing.T) {
	router := Router(app.New(nil))

	createRequest := authorizedRequest(t, http.MethodPost, "/api/v1/taxpayers", `{"name":"Fictional Cooperative","identifier":"demo-401"}`)
	createRequest.Header.Set("Idempotency-Key", "create-401")
	created := httptest.NewRecorder()
	router.ServeHTTP(created, createRequest)
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d body = %s", created.Code, created.Body)
	}
	var taxpayer struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &taxpayer); err != nil || taxpayer.ID == "" {
		t.Fatalf("taxpayer response = %s, %v", created.Body, err)
	}

	submitted := httptest.NewRecorder()
	router.ServeHTTP(submitted, authorizedRequest(
		t, http.MethodPost,
		fmt.Sprintf("/api/v1/taxpayers/%s/tax-registrations", taxpayer.ID),
		`{"taxType":"SAMPLE_INCOME"}`,
	))
	if submitted.Code != http.StatusCreated || !strings.Contains(submitted.Body.String(), `"status":"SUBMITTED"`) {
		t.Fatalf("submit status = %d body = %s", submitted.Code, submitted.Body)
	}
	var registration struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(submitted.Body.Bytes(), &registration); err != nil || registration.ID == "" {
		t.Fatalf("registration response = %s, %v", submitted.Body, err)
	}

	approved := httptest.NewRecorder()
	router.ServeHTTP(approved, authorizedRequest(
		t, http.MethodPost,
		fmt.Sprintf("/api/v1/tax-registrations/%s/approve", registration.ID), "",
	))
	if approved.Code != http.StatusOK || !strings.Contains(approved.Body.String(), `"status":"APPROVED"`) {
		t.Fatalf("approve status = %d body = %s", approved.Code, approved.Body)
	}

	retrieved := httptest.NewRecorder()
	router.ServeHTTP(retrieved, authorizedRequest(
		t, http.MethodGet, fmt.Sprintf("/api/v1/tax-registrations/%s", registration.ID), "",
	))
	if retrieved.Code != http.StatusOK || !strings.Contains(retrieved.Body.String(), `"status":"APPROVED"`) {
		t.Fatalf("retrieve status = %d body = %s", retrieved.Code, retrieved.Body)
	}

	repeatedApproval := httptest.NewRecorder()
	router.ServeHTTP(repeatedApproval, authorizedRequest(
		t, http.MethodPost,
		fmt.Sprintf("/api/v1/tax-registrations/%s/approve", registration.ID), "",
	))
	if repeatedApproval.Code != http.StatusConflict {
		t.Fatalf("second approval status = %d body = %s", repeatedApproval.Code, repeatedApproval.Body)
	}
	if repeatedApproval.Header().Get("Content-Type") != "application/problem+json" ||
		strings.Contains(repeatedApproval.Body.String(), "only a submitted registration") {
		t.Fatalf("unsafe problem response = %s", repeatedApproval.Body)
	}
	drafted := httptest.NewRecorder()
	router.ServeHTTP(drafted, authorizedRequest(
		t, http.MethodPost, "/api/v1/returns",
		fmt.Sprintf(
			`{"taxpayerId":%q,"registrationId":%q,"periodId":"FY-DEMO-2026","formVersion":"sample-income-v1","ruleVersion":"fictional-flat-rate-v1","lines":[{"code":"GROSS","amountMinor":10005}]}`,
			taxpayer.ID, registration.ID,
		),
	))
	if drafted.Code != http.StatusCreated || !strings.Contains(drafted.Body.String(), `"formVersion":"sample-income-v1"`) {
		t.Fatalf("draft status = %d body = %s", drafted.Code, drafted.Body)
	}
	var taxReturn struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(drafted.Body.Bytes(), &taxReturn); err != nil || taxReturn.ID == "" {
		t.Fatalf("return response = %s, %v", drafted.Body, err)
	}

	for _, action := range []string{"validate", "calculate"} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, authorizedRequest(
			t, http.MethodPost, fmt.Sprintf("/api/v1/returns/%s/%s", taxReturn.ID, action), "",
		))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d body = %s", action, response.Code, response.Body)
		}
	}
	submission := httptest.NewRecorder()
	router.ServeHTTP(submission, authorizedRequest(
		t, http.MethodPost, fmt.Sprintf("/api/v1/returns/%s/submit", taxReturn.ID), "",
	))
	if submission.Code != http.StatusCreated ||
		!strings.Contains(submission.Body.String(), `"minor":1001`) {
		t.Fatalf("submission status = %d body = %s", submission.Code, submission.Body)
	}
	var assessment struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(submission.Body.Bytes(), &assessment); err != nil || assessment.ID == "" {
		t.Fatalf("assessment response = %s, %v", submission.Body, err)
	}
	paymentRequest := authorizedRequest(
		t, http.MethodPost, "/api/v1/payments",
		fmt.Sprintf(
			`{"taxpayerId":%q,"amountMinor":1001,"currency":"XCR"}`,
			taxpayer.ID,
		),
	)
	paymentRequest.Header.Set("Idempotency-Key", "payment-api-1")
	paymentResponse := httptest.NewRecorder()
	router.ServeHTTP(paymentResponse, paymentRequest)
	if paymentResponse.Code != http.StatusCreated ||
		!strings.Contains(paymentResponse.Body.String(), `"status":"UNAPPLIED"`) {
		t.Fatalf("payment status = %d body = %s", paymentResponse.Code, paymentResponse.Body)
	}
	var payment struct {
		ID      string `json:"id"`
		Version uint64 `json:"version"`
	}
	if err := json.Unmarshal(paymentResponse.Body.Bytes(), &payment); err != nil || payment.ID == "" {
		t.Fatalf("payment response = %s, %v", paymentResponse.Body, err)
	}
	allocationResponse := httptest.NewRecorder()
	router.ServeHTTP(allocationResponse, authorizedRequest(
		t, http.MethodPost, fmt.Sprintf("/api/v1/payments/%s/allocations", payment.ID),
		fmt.Sprintf(
			`{"assessmentId":%q,"amountMinor":1001,"currency":"XCR","expectedVersion":%d}`,
			assessment.ID, payment.Version,
		),
	))
	if allocationResponse.Code != http.StatusOK ||
		!strings.Contains(allocationResponse.Body.String(), `"status":"ALLOCATED"`) {
		t.Fatalf("allocation status = %d body = %s", allocationResponse.Code, allocationResponse.Body)
	}
	ledgerResponse := httptest.NewRecorder()
	router.ServeHTTP(ledgerResponse, authorizedRequest(
		t, http.MethodGet, fmt.Sprintf("/api/v1/taxpayers/%s/ledger", taxpayer.ID), "",
	))
	if ledgerResponse.Code != http.StatusOK ||
		!strings.Contains(ledgerResponse.Body.String(), `"netDueMinor":0`) ||
		strings.Count(ledgerResponse.Body.String(), `"postingId":`) != 6 {
		t.Fatalf("ledger status = %d body = %s", ledgerResponse.Code, ledgerResponse.Body)
	}
	amendment := httptest.NewRecorder()
	router.ServeHTTP(amendment, authorizedRequest(
		t, http.MethodPost, fmt.Sprintf("/api/v1/returns/%s/amend", taxReturn.ID), "",
	))
	if amendment.Code != http.StatusCreated ||
		!strings.Contains(amendment.Body.String(), `"revision":2`) {
		t.Fatalf("amend status = %d body = %s", amendment.Code, amendment.Body)
	}
	history := httptest.NewRecorder()
	router.ServeHTTP(history, authorizedRequest(
		t, http.MethodGet, fmt.Sprintf("/api/v1/returns/%s/history", taxReturn.ID), "",
	))
	if history.Code != http.StatusOK ||
		strings.Count(history.Body.String(), `"revision":`) != 2 {
		t.Fatalf("history status = %d body = %s", history.Code, history.Body)
	}
}
