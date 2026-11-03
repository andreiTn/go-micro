package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/andreiTn/go-micro/authentication-service/data"
	"github.com/andreiTn/go-micro/pkg/httpx"
)

// UserStore loads users for authentication.
type UserStore interface {
	GetByEmail(email string) (*data.User, error)
}

func (app *Config) authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := httpx.ReadJSON(w, r, &requestPayload); err != nil {
		_ = httpx.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Users.GetByEmail(requestPayload.Email)
	if err != nil {
		_ = httpx.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		_ = httpx.ErrorJSON(w, errors.New("invalid credentials"), http.StatusUnauthorized)
		return
	}

	if err := app.logRequest(r.Context(), "authentication", fmt.Sprintf("%s logged in", user.Email)); err != nil {
		_ = httpx.ErrorJSON(w, err)
		return
	}

	_ = httpx.WriteJSON(w, http.StatusAccepted, httpx.Response{
		Error:   false,
		Message: fmt.Sprintf("Logged in as %s", user.Email),
		Data:    user,
	})
}

func (app *Config) logRequest(ctx context.Context, name, msg string) error {
	_, err := app.LoggerHTTP.PostJSONExpect(ctx, app.LoggerURL, map[string]string{
		"name": name,
		"data": msg,
	}, http.StatusAccepted)
	return err
}
