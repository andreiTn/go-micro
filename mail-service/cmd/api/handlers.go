package main

import (
	"net/http"

	"github.com/andreiTn/go-micro/pkg/httpx"
)

func (app *Config) sendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	if err := httpx.ReadJSON(w, r, &requestPayload); err != nil {
		_ = httpx.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	if err := app.Mailer.SendSMTPMessage(msg); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: "Sent to " + requestPayload.To,
	})
}
