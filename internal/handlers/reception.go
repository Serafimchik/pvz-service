package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/repository"

	"github.com/go-chi/chi/v5"
)

func CreateReceptionHandler(repo repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct {
			PVZID string `json:"pvzId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		reception, err := repo.CreateReception(requestBody.PVZID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(reception)
	}
}

func CloseReceptionHandler(repo repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pvzID := chi.URLParam(r, "pvzId")
		if pvzID == "" {
			http.Error(w, "Missing pvzId in path", http.StatusBadRequest)
			return
		}

		reception, err := repo.CloseReception(pvzID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reception)
	}
}
