package httpx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadWriteJSONRoundTrip(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	body, _ := json.Marshal(payload{Name: "demo"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	var got payload
	if err := ReadJSON(rr, req, &got); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if got.Name != "demo" {
		t.Fatalf("got %q", got.Name)
	}

	w := httptest.NewRecorder()
	if err := WriteJSON(w, http.StatusOK, Response{Message: "ok", Data: got}); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
}

func TestErrorJSON(t *testing.T) {
	w := httptest.NewRecorder()
	_ = ErrorJSON(w, errSample{}, http.StatusUnauthorized)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", w.Code)
	}
	var resp Response
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Error || resp.Message != "boom" {
		t.Fatalf("unexpected %#v", resp)
	}
}

type errSample struct{}

func (errSample) Error() string { return "boom" }
