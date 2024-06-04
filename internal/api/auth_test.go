package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evgfitil/gophermart.git/internal/mocks"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func TestHandleUserLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockUserStorage(ctrl)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mockStorage.EXPECT().GetUserByUsername(gomock.Any(), "test_user").Return(string(hashedPassword), nil).AnyTimes()
	mockStorage.EXPECT().GetUserByUsername(gomock.Any(), "wrong_user").Return("", sql.ErrNoRows).AnyTimes()
	mockStorage.EXPECT().GetUserByUsername(gomock.Any(), "bad_user").Return("", errors.New("internal error")).AnyTimes()

	ts := httptest.NewServer(HandleUserLogin(mockStorage))
	defer ts.Close()

	type want struct {
		statusCode int
		authHeader bool
	}

	tests := []struct {
		name          string
		requestMethod string
		requestBody   models.User
		want          want
	}{
		{
			name:          "successful authentication",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "test_user", Password: "password"},
			want:          want{http.StatusOK, true},
		},
		{
			name:          "wrong password",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "test_user", Password: "wrong_password"},
			want:          want{http.StatusUnauthorized, false},
		},
		{
			name:          "wrong username",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "wrong_user", Password: "password"},
			want:          want{http.StatusUnauthorized, false},
		},
		{
			name:          "bad request",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "test_user"},
			want:          want{http.StatusBadRequest, false},
		},
		{
			name:          "internal server error",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "bad_user", Password: "test_pass"},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, tt.requestMethod, ts.URL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
			_, ok := resp.Header["Authorization"]
			if tt.want.authHeader && !ok {
				t.Errorf("expected header 'Authorization' header to be set")
			}
		})
	}
}

func TestHandleUserRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockUserStorage(ctrl)
	mockStorage.EXPECT().CreateUser(gomock.Any(), "test_user", gomock.Any()).Return(nil).AnyTimes()
	mockStorage.EXPECT().IsUserUnique(gomock.Any(), "test_user").Return(true, nil).AnyTimes()
	mockStorage.EXPECT().CreateUser(gomock.Any(), "bad_user", gomock.Any()).Return(errors.New("internal error")).AnyTimes()
	mockStorage.EXPECT().IsUserUnique(gomock.Any(), "bad_user").Return(true, nil).AnyTimes()
	mockStorage.EXPECT().IsUserUnique(gomock.Any(), "exists_user").Return(false, nil).AnyTimes()

	ts := httptest.NewServer(HandleUserRegistration(mockStorage))
	defer ts.Close()

	type want struct {
		statusCode int
	}

	tests := []struct {
		name          string
		requestMethod string
		requestBody   models.User
		want          want
	}{
		{
			name:          "new user",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "test_user", Password: "test_pass"},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "exists user",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "exists_user", Password: "test_pass"},
			want: want{
				statusCode: http.StatusConflict,
			},
		},
		{
			name:          "bad request",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "wo_password"},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "internal server error",
			requestMethod: http.MethodPost,
			requestBody:   models.User{Username: "bad_user", Password: "test_pass"},
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, tt.requestMethod, ts.URL, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}
