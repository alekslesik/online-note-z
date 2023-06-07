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
	"golang.org/x/text/cases"
)

// POST /notes/create
func CreateNote(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// take logger and context
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		// models.Note instanse
		var noteRequest models.Note

		// decode request body to instanse
		err := json.NewDecoder(r.Body).Decode(&noteRequest)
		if err != nil {
			l.Error().Err(err).Msgf("error decoding the Note into httplib.JSON during registration. %v", err)
			httplib.JSON(w, httplib.Msg{"error": "internal error decoding Note struct"}, http.StatusInternalServerError)
			return
		}

		validate := validator.New()

		// validate struct
		err = validate.Struct(&noteRequest)
		if err != nil {
			l.Error().Err(err).Msgf("error during Note struct validation %v", err)
			httplib.JSON(w, httplib.Msg{"error": "wrongly formatted or missing Note parameter"}, http.StatusBadRequest)
			return
		}

		// create node in DB
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

		// return successful JSON response to user
		default:
			l.Info().Msgf("Note with ID %v has been created for user: %s", retID, noteRequest.User)
			httplib.JSON(w, httplib.Msg{"success": "note creation successful!"}, http.StatusCreated)
		}
	}
}

// GET /notes/
func GetAllNotesFromUser(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// take logger and context
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		// get username from request URL
		username := r.URL.Query().Get("username")

		if username == "" {
			l.Error().Msgf("error fetching username, the request parameter is empty. %s", username)
			httplib.JSON(w, httplib.Msg{"error": "user not in request params"}, http.StatusBadRequest)
			return
		}

		// get all notes from user
		notes, err := s.GetAllNotesFromUser(ctx, username)
		switch {
		case errors.Is(err, note.ErrNotFound):
			l.Info().Msgf("Requested user has no Notes!. %s", username)
		case errors.Is(err, note.ErrDBInternal):
			l.Info().Err(err).Msgf("Could not retrieve Notes for user. %v", err)
			httplib.JSON(w, httplib.Msg{"error": "could not retrieve notes for user"}, http.StatusInternalServerError)
			return

		// return successful JSON response to user
		default:
			l.Info().Msgf("Retriving user notes for %s was successful!", username)
			httplib.JSON(w, notes, http.StatusOK)
		}
	}
}
