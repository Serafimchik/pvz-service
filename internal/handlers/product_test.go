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

func TestAddProductHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("AddProduct", "электроника", "test-pvz-id").Return(&repository.Product{
		ID:          "test-product-id",
		DateTime:    parseTime("2023-10-01T12:00:00Z"),
		Type:        "электроника",
		ReceptionID: "test-reception-id",
	}, nil)

	handler := AddProductHandler(mockRepo)

	reqBody := map[string]string{
		"type":  "электроника",
		"pvzId": "test-pvz-id",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response repository.Product
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-product-id", response.ID)
	assert.Equal(t, "электроника", response.Type)
	assert.Equal(t, "test-reception-id", response.ReceptionID)

	mockRepo.AssertCalled(t, "AddProduct", "электроника", "test-pvz-id")
}

func TestAddProductHandler_InvalidType(t *testing.T) {
	mockRepo := new(MockRepository)

	handler := AddProductHandler(mockRepo)

	reqBody := map[string]string{
		"type":  "невалидный-тип",
		"pvzId": "test-pvz-id",
	}
	reqBodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(reqBodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	assert.Equal(t, "Invalid product type\n", rr.Body.String())

	mockRepo.AssertNotCalled(t, "AddProduct")
}

func TestDeleteLastProductHandler_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("DeleteLastProduct", "test-pvz-id").Return(nil)

	handler := DeleteLastProductHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodPost, "/pvz/test-pvz-id/delete_last_product", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("pvzId", "test-pvz-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	mockRepo.AssertCalled(t, "DeleteLastProduct", "test-pvz-id")
}

func TestDeleteLastProductHandler_MissingPVZID(t *testing.T) {
	mockRepo := new(MockRepository)

	handler := DeleteLastProductHandler(mockRepo)

	req, _ := http.NewRequest(http.MethodPost, "/pvz//delete_last_product", nil)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	assert.Equal(t, "Missing pvzId in path\n", rr.Body.String())

	mockRepo.AssertNotCalled(t, "DeleteLastProduct")
}
