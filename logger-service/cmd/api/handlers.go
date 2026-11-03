package main

import (
	"net/http"

	"github.com/andreiTn/go-micro/logger-service/internal/data"
	"github.com/andreiTn/go-micro/pkg/httpx"
)

type JSONPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) WriteLog(w http.ResponseWriter, r *http.Request) {
	var requestPayload JSONPayload

	if err := httpx.ReadJSON(w, r, &requestPayload); err != nil {
		_ = httpx.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	event := data.LogEntry{
		Name: requestPayload.Name,
		Data: requestPayload.Data,
	}

	if err := app.Models.LogEntry.Insert(r.Context(), event); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "Logged",
	})
}
