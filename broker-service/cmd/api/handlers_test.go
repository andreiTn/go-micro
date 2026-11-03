package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/andreiTn/go-micro/pkg/clients"
)

func TestAuthenticateProxiesAuthService(t *testing.T) {
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   false,
			"message": "Logged in",
			"data":    map[string]string{"email": "user@example.com"},
		})
	}))
	defer authSrv.Close()

	app := Config{
		AuthURL: authSrv.URL,
		HTTP:    clients.NewHTTP(5 * time.Second),
	}

	body, _ := json.Marshal(RequestPayload{
		Action: "auth",
		Auth: AuthPayload{
			Email:    "user@example.com",
			Password: "secret",
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/handle", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	app.HandleSubmission(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
