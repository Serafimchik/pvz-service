package handlers

import (
	"encoding/json"
	"net/http"

	"pvz-service/internal/repository"
)

func AddProductHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestBody struct {
			Type  string `json:"type"`
			PVZID string `json:"pvzId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		validTypes := map[string]bool{"электроника": true, "одежда": true, "обувь": true}
		if !validTypes[requestBody.Type] {
			http.Error(w, "Invalid product type", http.StatusBadRequest)
			return
		}

		product, err := repo.AddProduct(requestBody.Type, requestBody.PVZID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(product)
	}
}
