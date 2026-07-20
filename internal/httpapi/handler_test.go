package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/azamir911/ratelimit"
)

func TestCheckEndpoint(t *testing.T) {
	t.Parallel()
	limiter, err := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()
	handler := NewHandler(limiter)

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodPost, "/v1/check", strings.NewReader(`{"key":"tenant:42"}`)))
	if first.Code != http.StatusOK || !strings.Contains(first.Body.String(), `"allowed":true`) {
		t.Fatalf("unexpected first response: code=%d body=%s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, httptest.NewRequest(http.MethodPost, "/v1/check", strings.NewReader(`{"key":"tenant:42"}`)))
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), `"allowed":false`) {
		t.Fatalf("unexpected second response: code=%d body=%s", second.Code, second.Body.String())
	}
}

func TestCheckEndpointRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	limiter, err := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()
	handler := NewHandler(limiter)

	tests := []string{
		`{}`,
		`{"key":"a","unknown":true}`,
		`{"key":"a"} {"key":"b"}`,
	}
	for _, body := range tests {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/v1/check", strings.NewReader(body)))
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %q, got %d", body, recorder.Code)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()
	limiter, err := ratelimit.New(ratelimit.Config{Limit: 1, Window: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	defer limiter.Close()

	recorder := httptest.NewRecorder()
	NewHandler(limiter).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}
