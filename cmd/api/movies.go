package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/noonacedia/cinematrique/internal/data"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "creating a movie...\n")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	data := &data.Movie{
		ID:        1,
		CreatedAt: time.Now(),
		Title:     "Big Sleep",
		Year:      1952,
		Runtime:   128,
		Genres:    []string{"noir", "drama"},
		Version:   1,
	}
	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "Server cannot process response", http.StatusInternalServerError)
	}
}
