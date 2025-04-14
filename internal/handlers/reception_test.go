package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateReceptionHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	mockRepo.On("CreateReception", "test-pvz-id").Return(&repository.Reception{
		ID:       "test-reception-id",
		DateTime: parseTime("2023-10-01T12:00:00Z"),
		PVZID:    "test-pvz-id",
		Status:   "in_progress",
	}, nil)

	handler := CreateReceptionHandler(mockRepo)

	reqBody := map[string]string{
		"pvzId": "test-pvz-id",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response repository.Reception
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-reception-id", response.ID)
	assert.Equal(t, "test-pvz-id", response.PVZID)
	assert.Equal(t, "in_progress", response.Status)

	mockRepo.AssertCalled(t, "CreateReception", "test-pvz-id")
}

func TestCloseReceptionHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	mockRepo.On("CloseReception", "test-pvz-id").Return(&repository.Reception{
		ID:       "test-reception-id",
		DateTime: parseTime("2023-10-01T12:00:00Z"),
		PVZID:    "test-pvz-id",
		Status:   "close",
	}, nil)

	handler := CloseReceptionHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodPost, "/pvz/test-pvz-id/close_last_reception", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pvzId", "test-pvz-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response repository.Reception
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-reception-id", response.ID)
	assert.Equal(t, "test-pvz-id", response.PVZID)
	assert.Equal(t, "close", response.Status)

	mockRepo.AssertCalled(t, "CloseReception", "test-pvz-id")
}

func TestCloseReceptionHandler_MissingPVZID(t *testing.T) {
	mockRepo := new(MockRepository)

	handler := CloseReceptionHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodPost, "/pvz//close_last_reception", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	assert.Equal(t, "Missing pvzId in path\n", rr.Body.String())

	mockRepo.AssertNotCalled(t, "CloseReception")
}
