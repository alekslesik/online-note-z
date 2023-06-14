package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	// "time"

	db "github.com/adykaaa/online-notes/db/sqlc"
	mocksvc "github.com/adykaaa/online-notes/note/mock"
	// auth "github.com/adykaaa/online-notes/server/http/auth"
	"github.com/alekslesik/online-note-z/lib/password"
	"github.com/alekslesik/online-note-z/note"
	models "github.com/alekslesik/online-note-z/server/http/models"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type regUserMatcher db.RegisterUserParams

func (m *regUserMatcher) Matches(x interface{}) bool {
	reflectedValue := reflect.ValueOf(x).Elem()
	if m.Username != reflectedValue.FieldByName("Username").String() {
		return false
	}
	if m.Email != reflectedValue.FieldByName("Email").String() {
		return false
	}
	err := password.Validate(reflectedValue.FieldByName("Password").String(), m.Password)
	if err != nil {
		return false
	}

	return true
}

func (m *regUserMatcher) String() string {
	return fmt.Sprintf("Username: %s, Email: %s", m.Username, m.Email)
}

func TestRegisterUser(t *testing.T) {
	jsonValidator := validator.New()

	testCases := []struct {
		name          string
		body          *models.User
		validateJSON  func(t *testing.T, v *validator.Validate, u *models.User)
		mockSvcCall   func(mocksvc *mocksvc.MockNoteService, u *models.User)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "user registration OK",

			body: &models.User{
				Username: "user1",
				Password: "password1",
				Email:    "user1@user.com",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.NoError(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
				mocksvc.EXPECT().RegisterUser(gomock.Any(), &regUserMatcher{
					Username: u.Username,
					Password: u.Password,
					Email:    u.Email,
				}).Times(1).Return(u.Username, nil)
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusCreated, recorder.Code)
			},
		}, {
			name: "returns bad request - short username",

			body: &models.User{
				Username: "u",
				Password: "password1",
				Email:    "user1@user.com",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.Error(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "returns bad request - short password",

			body: &models.User{
				Username: "username1",
				Password: "pw1",
				Email:    "user1@user.com",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.Error(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "returns bad request - malformatted email",

			body: &models.User{
				Username: "username1",
				Password: "password1",
				Email:    "wrongemail@",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.Error(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "returns forbidden - duplicate username",

			body: &models.User{
				Username: "username1",
				Password: "password1",
				Email:    "user1@user.com",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.NoError(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
				mocksvc.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(1).Return("", note.ErrAlreadyExists)
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "returns internal error - DB failure",

			body: &models.User{
				Username: "username1",
				Password: "password1",
				Email:    "user1@user.com",
			},

			validateJSON: func(t *testing.T, v *validator.Validate, user *models.User) {
				err := v.Struct(user)
				require.NoError(t, err)
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService, u *models.User) {
				mocksvc.EXPECT().RegisterUser(gomock.Any(), gomock.Any()).Times(1).Return("", note.ErrDBInternal)
			},

			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for c := range testCases {
		tc := testCases[c]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mocksvc := mocksvc.NewMockNoteService(ctrl)

			tc.validateJSON(t, jsonValidator, tc.body)
			tc.mockSvcCall(mocksvc, tc.body)

			b, err := json.Marshal(tc.body)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))

			handler := RegisterUser(mocksvc)
			handler(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}

func TestGetAllNotesFromUser(t *testing.T) {

	const username = "testuser1"

	testCases := []struct {
		name          string
		addQuery      func(t *testing.T, r *http.Request)
		mockSvcCall   func(svcmock *mocksvc.MockNoteService)
		checkResponse func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name: "gettings notes from user OK",

			addQuery: func(t *testing.T, r *http.Request) {
				q := r.URL.Query()
				q.Add("username", username)
				r.URL.RawQuery = q.Encode()
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService) {
				mocksvc.EXPECT().GetAllNotesFromUser(gomock.Any(), username).Times(1).Return([]db.Note{}, nil)
			},

			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, rec.Code)
			},
		},
		{
			name: "returns bad request - missing url param",

			addQuery: func(t *testing.T, r *http.Request) {
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService) {
			},

			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, rec.Code)
			},
		},
		{
			name: "returns internal server error - db error",

			addQuery: func(t *testing.T, r *http.Request) {
				q := r.URL.Query()
				q.Add("username", username)
				r.URL.RawQuery = q.Encode()
			},

			mockSvcCall: func(mocksvc *mocksvc.MockNoteService) {
				mocksvc.EXPECT().GetAllNotesFromUser(gomock.Any(), username).Times(1).Return(nil, note.ErrDBInternal)
			},

			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, rec.Code)
			},
		},
	}

	for c := range testCases {
		tc := testCases[c]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mocksvc := mocksvc.NewMockNoteService(ctrl)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/notes", nil)

			tc.addQuery(t, req)
			tc.mockSvcCall(mocksvc)

			handler := GetAllNotesFromUser(mocksvc)
			handler(rec, req)
			tc.checkResponse(t, rec)
		})
	}
}
