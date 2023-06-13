package server

import (
	"net/http"
	"time"

	auth "github.com/adykaaa/online-notes/server/http/auth"
)

// POST /register/
func RegisterUser(s NoteService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// POST /login/
func LoginUser(s NoteService, token auth.TokenManager, tokenDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// POST /logout/
func LogoutUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
