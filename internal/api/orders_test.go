package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evgfitil/gophermart.git/internal/apperrors"
	"github.com/evgfitil/gophermart.git/internal/mocks"
	"github.com/evgfitil/gophermart.git/internal/models"
)

func TestHandleGetUserOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStorage := mocks.NewMockOrderStorage(ctrl)
	mockUserStorage := mocks.NewMockUserStorage(ctrl)
	handler := HandleGetUserOrders(mockOrderStorage, mockUserStorage)

	tokenAuth = jwtauth.New("HS256", []byte("jwtDefaultSecret"), nil)
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": "test_user", "exp": time.Now().Add(5 * time.Second).Unix()})

	r := http.NewServeMux()
	r.Handle("/api/user/orders", jwtauth.Verifier(tokenAuth)(jwtauth.Authenticator(handler)))

	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		statusCode int
		orders     []models.Order
	}
	tests := []struct {
		name          string
		requestMethod string
		authHeader    string
		mockSetup     func()
		want          want
	}{
		{
			name:          "successful get orders",
			requestMethod: http.MethodGet,
			authHeader:    "Bearer " + tokenString,
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().GetOrders(gomock.Any(), 1).Return([]models.Order{
					{OrderNumber: "1234567890", Status: "NEW", UploadedAt: time.Now()},
				}, nil)
			},
			want: want{
				statusCode: http.StatusOK,
				orders: []models.Order{
					{OrderNumber: "1234567890", Status: "NEW", UploadedAt: time.Now()},
				},
			},
		},
		{
			name:          "unauthorized user",
			requestMethod: http.MethodGet,
			authHeader:    "",
			mockSetup:     func() {},
			want: want{
				statusCode: http.StatusUnauthorized,
				orders:     nil,
			},
		},
		{
			name:          "internal server error",
			requestMethod: http.MethodGet,
			authHeader:    "Bearer " + tokenString,
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().GetOrders(gomock.Any(), 1).Return(nil, errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				orders:     nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, err := http.NewRequest(tt.requestMethod, ts.URL+"/api/user/orders", nil)
			require.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.orders != nil {
				var orders []models.Order
				err = json.NewDecoder(resp.Body).Decode(&orders)
				require.NoError(t, err)

				for i := range orders {
					assert.Equal(t, tt.want.orders[i].OrderNumber, orders[i].OrderNumber)
					assert.Equal(t, tt.want.orders[i].Status, orders[i].Status)
				}
			}
		})
	}
}

func TestHandleUploadOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrderStorage := mocks.NewMockOrderStorage(ctrl)
	mockUserStorage := mocks.NewMockUserStorage(ctrl)
	handler := HandleUploadOrder(mockOrderStorage, mockUserStorage)

	tokenAuth = jwtauth.New("HS256", []byte("jwtDefaultSecret"), nil)
	_, tokenString, _ := tokenAuth.Encode(map[string]interface{}{"user_id": "test_user", "exp": time.Now().Add(5 * time.Second).Unix()})

	r := http.NewServeMux()
	r.Handle("/api/user/orders", jwtauth.Verifier(tokenAuth)(jwtauth.Authenticator(handler)))

	ts := httptest.NewServer(r)
	defer ts.Close()

	type want struct {
		statusCode int
		response   string
	}
	tests := []struct {
		name          string
		requestMethod string
		authHeader    string
		requestBody   string
		mockSetup     func()
		want          want
	}{
		{
			name:          "successful upload order",
			requestMethod: http.MethodPost,
			authHeader:    "Bearer " + tokenString,
			requestBody:   "12345678903",
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().ProcessOrder(gomock.Any(), gomock.Any()).Return(nil)
			},
			want: want{
				statusCode: http.StatusAccepted,
				response:   "upload in successfully",
			},
		},
		{
			name:          "unauthorized user",
			requestMethod: http.MethodPost,
			authHeader:    "",
			requestBody:   "12345678903",
			mockSetup:     func() {},
			want: want{
				statusCode: http.StatusUnauthorized,
				response:   "",
			},
		},
		{
			name:          "invalid order number",
			requestMethod: http.MethodPost,
			authHeader:    "Bearer " + tokenString,
			requestBody:   "invalid",
			mockSetup:     func() {},
			want: want{
				statusCode: http.StatusUnprocessableEntity,
				response:   "",
			},
		},
		{
			name:          "order already exists",
			requestMethod: http.MethodPost,
			authHeader:    "Bearer " + tokenString,
			requestBody:   "12345678903",
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().ProcessOrder(gomock.Any(), gomock.Any()).Return(apperrors.ErrOrderAlreadyExists)
			},
			want: want{
				statusCode: http.StatusOK,
				response:   "",
			},
		},
		{
			name:          "order number taken",
			requestMethod: http.MethodPost,
			authHeader:    "Bearer " + tokenString,
			requestBody:   "12345678903",
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().ProcessOrder(gomock.Any(), gomock.Any()).Return(apperrors.ErrOrderNumberTaken)
			},
			want: want{
				statusCode: http.StatusConflict,
				response:   "",
			},
		},
		{
			name:          "internal server error",
			requestMethod: http.MethodPost,
			authHeader:    "Bearer " + tokenString,
			requestBody:   "12345678903",
			mockSetup: func() {
				mockUserStorage.EXPECT().GetUserID(gomock.Any(), "test_user").Return(1, nil)
				mockOrderStorage.EXPECT().ProcessOrder(gomock.Any(), gomock.Any()).Return(errors.New("internal error"))
			},
			want: want{
				statusCode: http.StatusInternalServerError,
				response:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req, err := http.NewRequest(tt.requestMethod, ts.URL+"/api/user/orders", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			if tt.want.response != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want.response, string(body))
			}
		})
	}
}
