package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

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

func GetPVZListHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 30 {
			limit = 10
		}

		startDateStr := r.URL.Query().Get("startDate")
		endDateStr := r.URL.Query().Get("endDate")
		var startDate, endDate *time.Time

		if startDateStr != "" {
			parsedStartDate, err := time.Parse(time.RFC3339, startDateStr)
			if err != nil {
				http.Error(w, "Invalid startDate format", http.StatusBadRequest)
				return
			}
			startDate = &parsedStartDate
		}

		if endDateStr != "" {
			parsedEndDate, err := time.Parse(time.RFC3339, endDateStr)
			if err != nil {
				http.Error(w, "Invalid endDate format", http.StatusBadRequest)
				return
			}
			endDate = &parsedEndDate
		}

		pvzList, err := repo.GetPVZList(startDate, endDate, page, limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(pvzList)
	}
}
