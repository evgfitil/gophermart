package api

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evgfitil/gophermart.git/internal/mocks"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func TestHandleGetUserBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalanceStorage := mocks.NewMockBalanceStorage(ctrl)
	handler := HandleGetUserBalance(mockBalanceStorage)

	tokenAuth = jwtauth.New("HS256", []byte("jwtDefaultSecret"), nil)
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": "test_user", "exp": time.Now().Add(5 * time.Second).Unix()})

	r := http.NewServeMux()
	r.Handle("/api/user/balance", jwtauth.Verifier(tokenAuth)(jwtauth.Authenticator(handler)))

	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		statusCode int
		balance    *models.Balance
	}
	tests := []struct {
		name          string
		requestMethod string
		authHeader    string
		mockSetup     func()
		want          want
	}{
		{
			name:          "successful get balance",
			requestMethod: http.MethodGet,
			authHeader:    "Bearer " + tokenString,
			mockSetup: func() {
				mockBalanceStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockBalanceStorage.EXPECT().GetUserBalance(gomock.Any(), 1).Return(&models.Balance{Current: 500.5, Withdrawn: 42}, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				balance:    &models.Balance{Current: 500.5, Withdrawn: 42},
			},
		},
		{
			name:          "unauthorized user",
			requestMethod: http.MethodGet,
			authHeader:    "",
			mockSetup:     func() {},
			want: want{
				statusCode: http.StatusUnauthorized,
				balance:    nil,
			},
		},
		{
			name:          "internal server error",
			requestMethod: http.MethodGet,
			authHeader:    "Bearer " + tokenString,
			mockSetup: func() {
				mockBalanceStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockBalanceStorage.EXPECT().GetUserBalance(gomock.Any(), 1).Return(nil, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				balance:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, err := http.NewRequest(tt.requestMethod, ts.URL+"/api/user/balance", nil)
			require.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.balance != nil {
				var balance models.Balance
				err = json.NewDecoder(resp.Body).Decode(&balance)
				require.NoError(t, err)
				assert.Equal(t, *tt.want.balance, balance)
			}
		})
	}
}
