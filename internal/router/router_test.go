package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/volchkovski/gophermart-loyalty/internal/mocks"
	"github.com/volchkovski/gophermart-loyalty/internal/models"
	"github.com/volchkovski/gophermart-loyalty/internal/services/auth"
)

type testCase struct {
	name           string
	method         string
	path           string
	body           interface{}
	authToken      string
	expectedStatus int
	setupMocks     func(*mocks.MockProcessor)
}

func TestHTTPRouter_AuthEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProcessor := mocks.NewMockProcessor(ctrl)

	router := NewHTTPRouter(mockProcessor)

	tests := []testCase{
		{
			name:           "Register Success",
			method:         http.MethodPost,
			path:           "/api/user/register",
			body:           map[string]string{"login": "testuser", "password": "password"},
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				m.EXPECT().Register(gomock.Any(), "testuser", "password").Return("mock-token", nil)
			},
		},
		{
			name:           "Register Invalid JSON",
			method:         http.MethodPost,
			path:           "/api/user/register",
			body:           "invalid-json",
			expectedStatus: http.StatusBadRequest,
			setupMocks:     func(m *mocks.MockProcessor) {},
		},
		{
			name:           "Login Success",
			method:         http.MethodPost,
			path:           "/api/user/login",
			body:           map[string]string{"login": "testuser", "password": "password"},
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				m.EXPECT().Login(gomock.Any(), "testuser", "password").Return("mock-token", nil)
			},
		},
		{
			name:           "Login Invalid Credentials",
			method:         http.MethodPost,
			path:           "/api/user/login",
			body:           map[string]string{"login": "testuser", "password": "wrongpass"},
			expectedStatus: http.StatusUnauthorized,
			setupMocks: func(m *mocks.MockProcessor) {
				m.EXPECT().Login(gomock.Any(), "testuser", "wrongpass").Return("", auth.ErrNoUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockProcessor)

			var reqBody *bytes.Buffer
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = bytes.NewBufferString(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					reqBody = bytes.NewBuffer(bodyBytes)
				}
			} else {
				reqBody = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK && (tt.path == "/api/user/register" || tt.path == "/api/user/login") {
				authHeader := recorder.Header().Get("Authorization")
				assert.Contains(t, authHeader, "Bearer mock-token")
			}
		})
	}
}

func TestHTTPRouter_OrderEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProcessor := mocks.NewMockProcessor(ctrl)
	router := NewHTTPRouter(mockProcessor)

	tests := []testCase{
		{
			name:           "Get Orders Unauthorized",
			method:         http.MethodGet,
			path:           "/api/user/orders",
			expectedStatus: http.StatusUnauthorized,
			setupMocks:     func(m *mocks.MockProcessor) {},
		},
		{
			name:           "Get Orders Success",
			method:         http.MethodGet,
			path:           "/api/user/orders",
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
				orders := []*models.Order{
					{Number: "123", Status: "NEW"},
				}
				m.EXPECT().Orders(gomock.Any(), int64(1)).Return(orders, nil)
			},
		},
		{
			name:           "Post Order Success",
			method:         http.MethodPost,
			path:           "/api/user/orders",
			body:           "4561261212345467",
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusAccepted,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
				m.EXPECT().SaveOrder(gomock.Any(), int64(1), "4561261212345467").Return(nil)
			},
		},
		{
			name:           "Post Order Invalid Luhn",
			method:         http.MethodPost,
			path:           "/api/user/orders",
			body:           "123456789",
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusUnprocessableEntity,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockProcessor)

			var reqBody *bytes.Buffer
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = bytes.NewBufferString(str)
				} else {
					bodyBytes, _ := json.Marshal(tt.body)
					reqBody = bytes.NewBuffer(bodyBytes)
				}
			} else {
				reqBody = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			if tt.body != nil {
				req.Header.Set("Content-Type", "text/plain")
			}
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)
		})
	}
}

func TestHTTPRouter_BalanceEndpoints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProcessor := mocks.NewMockProcessor(ctrl)
	router := NewHTTPRouter(mockProcessor)

	tests := []testCase{
		{
			name:           "Get Balance Success",
			method:         http.MethodGet,
			path:           "/api/user/balance",
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
				balance := &models.Balance{Current: 1000, Withdrawn: 200}
				m.EXPECT().Balance(gomock.Any(), int64(1)).Return(balance, nil)
			},
		},
		{
			name:           "Withdraw Success",
			method:         http.MethodPost,
			path:           "/api/user/balance/withdraw",
			body:           map[string]interface{}{"order": "4561261212345467", "sum": 100.5},
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
				m.EXPECT().Withdraw(gomock.Any(), int64(1), int64(10050), "4561261212345467").Return(nil)
			},
		},
		{
			name:           "Get Withdrawals Success",
			method:         http.MethodGet,
			path:           "/api/user/withdrawals",
			authToken:      "Bearer valid-token",
			expectedStatus: http.StatusOK,
			setupMocks: func(m *mocks.MockProcessor) {
				claims := &models.CustomClaims{UserID: 1}
				m.EXPECT().VerifyToken("valid-token").Return(claims, nil)
				withdrawals := []*models.Withdrawal{
					{Order: "123", Sum: 1000},
				}
				m.EXPECT().Withdrawals(gomock.Any(), int64(1)).Return(withdrawals, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks(mockProcessor)

			var reqBody *bytes.Buffer
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				reqBody = bytes.NewBuffer(bodyBytes)
			} else {
				reqBody = bytes.NewBuffer(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if tt.authToken != "" {
				req.Header.Set("Authorization", tt.authToken)
			}

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedStatus == http.StatusOK && (tt.path == "/api/user/balance" || tt.path == "/api/user/withdrawals") {
				contentType := recorder.Header().Get("Content-Type")
				assert.Contains(t, contentType, "application/json")
			}
		})
	}
}

func TestNewHTTPRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProcessor := mocks.NewMockProcessor(ctrl)
	router := NewHTTPRouter(mockProcessor)

	require.NotNil(t, router)
	require.NotNil(t, router.Router)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}
