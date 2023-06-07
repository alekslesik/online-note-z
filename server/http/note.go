package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	httplib "github.com/alekslesik/online-note-z/lib/http"
	"github.com/alekslesik/online-note-z/note"
	models "github.com/alekslesik/online-note-z/server/http/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// POST /create
func CreateNote(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		var noteRequest models.Note

		err := json.NewDecoder(r.Body).Decode(&noteRequest)
		if err != nil {
			l.Error().Err(err).Msgf("error decoding the Note into httplib.JSON during registration. %v", err)
			httplib.JSON(w, httplib.Msg{"error": "internal error decoding Note struct"}, http.StatusInternalServerError)
			return
		}

		validate := validator.New()

		err = validate.Struct(&noteRequest)
		if err != nil {
			l.Error().Err(err).Msgf("error during Note struct validation %v", err)
			httplib.JSON(w, httplib.Msg{"error": "wrongly formatted or missing Note parameter"}, http.StatusBadRequest)
			return
		}

		retID, err := s.CreateNote(ctx, noteRequest.Title, noteRequest.User, noteRequest.Text)

		switch {
		case errors.Is(err, note.ErrAlreadyExists):
			l.Error().Err(err).Msgf("Note creation failed, a note with that title already exists")
			httplib.JSON(w, httplib.Msg{"error": "a Note with that title already exists! Titles must be unique."}, http.StatusForbidden)
			return
		case errors.Is(err, note.ErrDBInternal):
			l.Error().Err(err).Msgf("Error during Note creation! %v", err)
			httplib.JSON(w, httplib.Msg{"error": "internal error during note creation"}, http.StatusInternalServerError)
			return
		default:
			l.Info().Msgf("Note with ID %v has been created for user: %s", retID, noteRequest.User)
			httplib.JSON(w, httplib.Msg{"success": "note creation successful!"}, http.StatusCreated)
		}
	}
}
