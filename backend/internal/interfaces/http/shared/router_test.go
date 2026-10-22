package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type acceptingVerifier struct{}

func (acceptingVerifier) ParseAccess(string) (int64, error) { return 1, nil }

func TestHealthzReturnsDataEnvelope(t *testing.T) {
	router := NewRouter(Dependencies{Tokens: acceptingVerifier{}})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["data"]["status"] != "ok" {
		t.Fatalf("body=%s", response.Body.String())
	}
}
