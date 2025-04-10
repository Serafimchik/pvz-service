package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/repository"
)

func CreatePVZHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pvz repository.PVZ
		if err := json.NewDecoder(r.Body).Decode(&pvz); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		validCities := map[string]bool{"Москва": true, "Санкт-Петербург": true, "Казань": true}
		if !validCities[pvz.City] {
			http.Error(w, "Invalid city", http.StatusBadRequest)
			return
		}

		createdPVZ, err := repo.CreatePVZ(pvz.City)
		if err != nil {
			http.Error(w, "Failed to create PVZ", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdPVZ)
	}
}
