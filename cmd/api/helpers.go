package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func (app *application) readIDPathParam(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, errors.New("ID path parameter is invalid")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	jsonText, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// set any custom headers we want to set
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(jsonText, '\n'))
	return nil
}
