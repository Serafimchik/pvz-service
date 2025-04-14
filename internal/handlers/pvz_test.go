package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pvz-service/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestCreatePVZHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("CreatePVZ", "Москва").Return(&repository.PVZ{
		ID:               "test-pvz-id",
		RegistrationDate: time.Now(),
		City:             "Москва",
	}, nil)

	handler := CreatePVZHandler(mockRepo)

	reqBody := map[string]string{
		"city": "Москва",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response repository.PVZ
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-pvz-id", response.ID)
	assert.Equal(t, "Москва", response.City)

	mockRepo.AssertCalled(t, "CreatePVZ", "Москва")
}

func TestCreatePVZHandler_InvalidCity(t *testing.T) {
	mockRepo := new(MockRepository)

	handler := CreatePVZHandler(mockRepo)

	reqBody := map[string]string{
		"city": "НевалидныйГород",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	assert.Equal(t, "Invalid city\n", rr.Body.String())

	mockRepo.AssertNotCalled(t, "CreatePVZ")
}

func TestGetPVZListHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)

	startDate, _ := time.Parse(time.RFC3339, "2023-10-01T00:00:00Z")
	endDate, _ := time.Parse(time.RFC3339, "2023-10-31T23:59:59Z")
	mockRepo.On("GetPVZList", &startDate, &endDate, 1, 10).Return([]repository.PVZWithReceptions{
		{
			PVZ: repository.PVZ{
				ID:               "test-pvz-id",
				RegistrationDate: time.Now(),
				City:             "Москва",
			},
			Receptions: []repository.ReceptionWithProducts{},
		},
	}, nil)

	handler := GetPVZListHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodGet, "/pvz?startDate=2023-10-01T00:00:00Z&endDate=2023-10-31T23:59:59Z&page=1&limit=10", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []repository.PVZWithReceptions
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-pvz-id", response[0].PVZ.ID)
	assert.Equal(t, "Москва", response[0].PVZ.City)

	mockRepo.AssertCalled(t, "GetPVZList", &startDate, &endDate, 1, 10)
}

func TestGetPVZListHandler_InvalidStartDate(t *testing.T) {
	mockRepo := new(MockRepository)

	handler := GetPVZListHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodGet, "/pvz?startDate=invalid-date&endDate=2023-10-31T23:59:59Z&page=1&limit=10", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	assert.Equal(t, "Invalid startDate format\n", rr.Body.String())

	mockRepo.AssertNotCalled(t, "GetPVZList")
}
