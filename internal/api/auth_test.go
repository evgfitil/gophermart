package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/evgfitil/gophermart.git/internal/models"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) CreateUser(ctx context.Context, username string, passwordHash string) error {
	args := m.Called(ctx, username, passwordHash)
	return args.Error(0)
}

func (m *MockUserStorage) GetUserByUsername(ctx context.Context, username string) (string, error) {
	args := m.Called(ctx, username)
	return args.String(0), args.Error(1)
}

func (m *MockUserStorage) GetUserID(ctx context.Context, username string) (int, error) {
	return 0, nil
}

func (m *MockUserStorage) IsUserUnique(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserStorage) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	return nil, nil
}

func TestAuth(t *testing.T) {
	mockStorage := new(MockUserStorage)

	// Setup mocks for successful authentication
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	mockStorage.On("GetUserByUsername", mock.Anything, "test_user").Return(string(hashedPassword), nil)

	// Setup mocks for user authentication errors
	mockStorage.On("GetUserByUsername", mock.Anything, "wrong_user").Return("", sql.ErrNoRows)
	mockStorage.On("GetUserByUsername", mock.Anything, "bad_user", mock.Anything).Return("", errors.New("internal error"))

	ts := httptest.NewServer(AuthHandler(mockStorage))
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

func TestRegisterHandler(t *testing.T) {
	mockStorage := new(MockUserStorage)

	// Setup mocks for successful user creation
	mockStorage.On("CreateUser", mock.Anything, "test_user", mock.Anything).Return(nil)
	mockStorage.On("IsUserUnique", mock.Anything, "test_user").Return(true, nil)
	mockStorage.On("GetUserByUsername", mock.Anything, "test_user").Return("test_user", nil)

	// Setup mocks for user creation errors
	mockStorage.On("CreateUser", mock.Anything, "bad_user", mock.Anything).Return(errors.New("internal error"))
	mockStorage.On("IsUserUnique", mock.Anything, "bad_user").Return(true, nil)

	// Setup mocks for existing user check
	mockStorage.On("IsUserUnique", mock.Anything, "exists_user").Return(false, nil)

	ts := httptest.NewServer(RegisterHandler(mockStorage))
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
