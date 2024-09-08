package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	healthData := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	err := app.writeJSON(w, http.StatusOK, envelope{"health": healthData}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
