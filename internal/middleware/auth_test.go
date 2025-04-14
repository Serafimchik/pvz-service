package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockAuth := new(MockAuthenticator)

	mockClaims := map[string]interface{}{
		"role": "moderator",
	}
	mockAuth.On("ParseToken", "valid-token").Return(mockClaims, nil)

	authMiddleware := AuthMiddleware(mockAuth)

	var claimsInContext map[string]interface{}
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claimsInContext = r.Context().Value(claimsKey).(map[string]interface{})
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	rr := httptest.NewRecorder()

	authMiddleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Equal(t, mockClaims, claimsInContext)

	mockAuth.AssertCalled(t, "ParseToken", "valid-token")
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockAuth := new(MockAuthenticator)

	authMiddleware := AuthMiddleware(mockAuth)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rr := httptest.NewRecorder()

	authMiddleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	assert.Equal(t, "Missing Authorization header\n", rr.Body.String())

	mockAuth.AssertNotCalled(t, "ParseToken")
}
func TestGenerateJWT(t *testing.T) {
	auth := NewJWTAuthenticator()

	role := "moderator"
	token, err := auth.GenerateJWT(role)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedClaims, err := auth.ParseToken(token)
	assert.NoError(t, err)

	assert.Equal(t, role, parsedClaims["role"])

	expirationTime := int64(parsedClaims["exp"].(float64))
	assert.Greater(t, expirationTime, time.Now().Unix())
	assert.LessOrEqual(t, expirationTime, time.Now().Add(time.Hour).Unix())
}

func TestParseToken_ValidToken(t *testing.T) {
	auth := NewJWTAuthenticator()

	tokenString, err := auth.GenerateJWT("employee")
	assert.NoError(t, err)

	parsedClaims, err := auth.ParseToken(tokenString)
	assert.NoError(t, err)

	assert.Equal(t, "employee", parsedClaims["role"])
}

func TestParseToken_InvalidToken(t *testing.T) {
	auth := NewJWTAuthenticator()

	invalidToken := "invalid.token.string"
	parsedClaims, err := auth.ParseToken(invalidToken)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestParseToken_TamperedToken(t *testing.T) {
	auth := NewJWTAuthenticator()

	validToken, err := auth.GenerateJWT("moderator")
	assert.NoError(t, err)

	tamperedToken := validToken + "tampered"

	parsedClaims, err := auth.ParseToken(tamperedToken)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestParseToken_ExpiredToken(t *testing.T) {
	auth := NewJWTAuthenticator()

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"role": "moderator",
		"exp":  time.Now().Add(-1 * time.Hour).Unix(),
	})
	signedToken, err := expiredToken.SignedString([]byte("secret"))
	assert.NoError(t, err)

	parsedClaims, err := auth.ParseToken(signedToken)

	assert.Error(t, err)
	assert.Nil(t, parsedClaims)
	assert.Contains(t, err.Error(), "invalid token")
}
func TestRoleMiddleware_AllowedRole(t *testing.T) {
	roleMiddleware := RoleMiddleware("moderator")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.WithValue(context.Background(), claimsKey, map[string]interface{}{
		"role": "moderator",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	roleMiddleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRoleMiddleware_ForbiddenRole(t *testing.T) {
	roleMiddleware := RoleMiddleware("employee")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ctx := context.WithValue(context.Background(), claimsKey, map[string]interface{}{
		"role": "moderator",
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	rr := httptest.NewRecorder()

	roleMiddleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)

	assert.Equal(t, "Forbidden\n", rr.Body.String())
}

func TestLoggingMiddleware(t *testing.T) {
	var logOutput bytes.Buffer
	log.SetOutput(&logOutput)
	defer log.SetOutput(nil)

	loggingMiddleware := LoggingMiddleware

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test-path", nil)

	rr := httptest.NewRecorder()

	loggingMiddleware(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, logOutput.String(), "Request: GET /test-path")
}
