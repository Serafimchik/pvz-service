package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz-service/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) GenerateJWT(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *MockAuthenticator) ParseToken(tokenString string) (map[string]interface{}, error) {
	args := m.Called(tokenString)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestDummyLoginHandler_Success(t *testing.T) {
	mockAuth := new(MockAuthenticator)
	mockAuth.On("GenerateJWT", "moderator").Return("test-token", nil)

	reqBody := map[string]string{"role": "moderator"}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := DummyLoginHandler(mockAuth)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", response["token"])

	mockAuth.AssertCalled(t, "GenerateJWT", "moderator")
}

func TestLoginHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	mockAuth := new(MockAuthenticator)

	mockRepo.On("LoginUser", "test@example.com", "password123").Return(&repository.User{
		ID:    "123e4567-e89b-12d3-a456-426614174000",
		Email: "test@example.com",
		Role:  "moderator",
	}, nil)

	mockAuth.On("GenerateJWT", "moderator").Return("test-token", nil)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := LoginHandler(mockRepo, mockAuth)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-token", response["token"])

	mockRepo.AssertCalled(t, "LoginUser", "test@example.com", "password123")
	mockAuth.AssertCalled(t, "GenerateJWT", "moderator")
}

func TestRegisterHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	mockRepo.On("RegisterUser", "test@example.com", "password123", "employee").Return(&repository.User{
		ID:       "test-user-id",
		Email:    "test@example.com",
		Password: "hashed-password",
		Role:     "employee",
	}, nil)

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "employee",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler := RegisterHandler(mockRepo)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response repository.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-user-id", response.ID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "employee", response.Role)

	mockRepo.AssertCalled(t, "RegisterUser", "test@example.com", "password123", "employee")
}
