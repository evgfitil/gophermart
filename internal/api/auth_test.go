package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/evgfitil/gophermart.git/internal/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) CreateUser(ctx context.Context, username string, passwordHash string) error {
	args := m.Called(ctx, username, passwordHash)
	return args.Error(0)
}

func (m *MockDB) GetUserByUsername(ctx context.Context, username string) (string, error) {
	args := m.Called(ctx, username)
	return args.String(0), args.Error(1)
}

func (m *MockDB) IsUserUnique(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockDB) Ping(ctx context.Context) error {
	return nil
}

func TestRegisterHandler(t *testing.T) {
	mockStorage := new(MockDB)

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
