package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type envelope map[string]any

func (app *application) readIDPathParam(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, errors.New("ID path parameter is invalid")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	jsonText, err := json.Marshal(data)
	if err != nil {
		return err
	}
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(append(jsonText, '\n'))
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	mb := 1_024 * 1_024
	r.Body = http.MaxBytesReader(w, r.Body, int64(mb))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at char %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", mb)
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	err = decoder.Decode(r.Body)
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
