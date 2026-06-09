package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := app.readJSON(w, r, &requestPayload); err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	// Authenticate the user
	user, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
		return
	}
	
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid email or password"), http.StatusUnauthorized)
		return
	}

	payload := jsonResponse{
		Error: false,
		Message: fmt.Sprintf("Authentication successful for user %s", user.Email),
		Data: user,
	}

	app.writeJSON(w, http.StatusOK, payload)
}