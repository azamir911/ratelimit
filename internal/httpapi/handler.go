package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/azamir911/ratelimit"
)

const maxRequestBodyBytes = 1 << 20

type checkRequest struct {
	Key  string `json:"key"`
	Cost uint64 `json:"cost,omitempty"`
}

type checkResponse struct {
	Allowed         bool      `json:"allowed"`
	Limit           uint64    `json:"limit"`
	Count           uint64    `json:"count"`
	Remaining       uint64    `json:"remaining"`
	ResetAt         time.Time `json:"reset_at"`
	RetryAfterMilli int64     `json:"retry_after_ms"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// NewHandler creates the standalone service HTTP handler.
func NewHandler(limiter *ratelimit.Limiter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/check", checkHandler(limiter))
	return mux
}

func checkHandler(limiter *ratelimit.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var request checkRequest
		if err := decoder.Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request"})
			return
		}
		if err := ensureEOF(decoder); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request must contain one JSON object"})
			return
		}
		request.Key = strings.TrimSpace(request.Key)
		if request.Cost == 0 {
			request.Cost = 1
		}

		decision, err := limiter.AllowN(request.Key, request.Cost)
		if err != nil {
			switch {
			case errors.Is(err, ratelimit.ErrEmptyKey), errors.Is(err, ratelimit.ErrInvalidCost):
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			case errors.Is(err, ratelimit.ErrCapacity), errors.Is(err, ratelimit.ErrClosed):
				writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: err.Error()})
			default:
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
			}
			return
		}

		writeJSON(w, http.StatusOK, checkResponse{
			Allowed:         decision.Allowed,
			Limit:           decision.Limit,
			Count:           decision.Count,
			Remaining:       decision.Remaining,
			ResetAt:         decision.ResetAt,
			RetryAfterMilli: decision.RetryAfter.Milliseconds(),
		})
	}
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("unexpected trailing JSON value")
	}
	return err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
