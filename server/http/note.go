package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	// "flag"
	"net/http"
	"strings"

	httplib "github.com/alekslesik/online-note-z/lib/http"
	"github.com/alekslesik/online-note-z/note"
	models "github.com/alekslesik/online-note-z/server/http/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	// "golang.org/x/text/cases"
)

// POST /notes/create
func CreateNote(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// take logger and context
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		// models.Note instance
		var noteRequest models.Note

		// decode request body to instance
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
			l.Info().Msgf("Retrieving user notes for %s was successful!", username)
			httplib.JSON(w, notes, http.StatusOK)
		}
	}
}

// DELETE /notes/{id}
func DeleteNote(s NoteService) http.HandlerFunc {
	// take logger and context
	return func(w http.ResponseWriter, r *http.Request) {
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		// parse URL path and take "id" and convert it to UUID
		reqUUID, err := uuid.Parse(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			l.Info().Msgf("Could not convert ID to UUID.")
			httplib.JSON(w, httplib.Msg{"error": "could not convert note id to uuid"}, http.StatusBadRequest)
			return
		}

		// delete note
		id, err := s.DeleteNote(ctx, reqUUID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			l.Info().Msg("User has no notes to delete from!")
		case errors.Is(err, note.ErrDBInternal):
			l.Info().Err(err).Msgf("Could not delete Note %v from the DB!", reqUUID)
			httplib.JSON(w, httplib.Msg{"error": "could not delete note from DB"}, http.StatusInternalServerError)
			return

		// return successful JSON response to user
		default:
			l.Info().Msgf("Deleting note %v was successful!", id)
			httplib.JSON(w, httplib.Msg{"success": "note deleted"}, http.StatusOK)
			return
		}
	}
}

// PUT /notes/{id}
func UpdateNote(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// take logger and context
		l, ctx, cancel := httplib.SetupHandler(w, r.Context())
		defer cancel()

		var isTextValid bool = true

		// parse URL path and take "id" and convert it to UUID
		reqUUID, err := uuid.Parse(strings.Split(r.URL.Path, "/")[2])
		if err != nil {
			l.Info().Msgf("Could not convert ID to UUID.")
			httplib.JSON(w, httplib.Msg{"error": "could not convert note id to uuid"}, http.StatusBadRequest)
			return
		}

		// create struct for decode
		updateRequest := struct {
			Title string `json:"title" validate:"required, min=4"`
			Text  string `json:"text"`
		}{}

		// decode request body
		err = json.NewDecoder(r.Body).Decode(&updateRequest)
		if err != nil {
			l.Error().Err(err).Msgf("error decoding the Note into httplib.JSON during registration. %v", err)
			httplib.JSON(w, httplib.Msg{"error": "internal error decoding Note struct"}, http.StatusInternalServerError)
			return
		}

		validate := validator.New()

		// struct validate
		err = validate.Struct(&updateRequest)
		if err != nil {
			l.Error().Err(err).Msgf("title must be more than 4 characters long!")
			httplib.JSON(w, httplib.Msg{"error": "title of a note must be more than 4 characters long!"}, http.StatusBadRequest)
			return
		}

		if updateRequest.Text == "" {
			isTextValid = false
		}

		// update struct in DB
		id, err := s.UpdateNote(ctx, reqUUID, updateRequest.Title, updateRequest.Text, isTextValid)
		switch {
		case errors.Is(err, note.ErrDBInternal):
			l.Info().Err(err).Msgf("Could not update Note %v", reqUUID)
			httplib.JSON(w, httplib.Msg{"error": "could not update note"}, http.StatusInternalServerError)
			return
		default:
			l.Info().Msgf("Updating note %v was successful!", id)
			httplib.JSON(w, httplib.Msg{"success": "note deleted"}, http.StatusOK)
			return
		}
	}
}
