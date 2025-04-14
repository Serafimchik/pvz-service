package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pvz-service/internal/handlers"
	"pvz-service/internal/middleware"
	"pvz-service/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func MockServer(repo repository.Repository) *httptest.Server {
	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)

	auth := middleware.NewJWTAuthenticator()
	r.Post("/dummyLogin", handlers.DummyLoginHandler(auth))
	r.Post("/register", handlers.RegisterHandler(repo))
	r.Post("/login", handlers.LoginHandler(repo, auth))

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(auth))

		r.Route("/pvz", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					allowedRoles := []string{}
					switch r.Method {
					case http.MethodPost:
						allowedRoles = []string{"moderator"}
					case http.MethodGet:
						allowedRoles = []string{"employee", "moderator"}
					default:
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
						return
					}
					middleware.RoleMiddleware(allowedRoles...)(next).ServeHTTP(w, r)
				})
			})
			r.Post("/", handlers.CreatePVZHandler(repo))
			r.Get("/", handlers.GetPVZListHandler(repo))
		})

		r.Route("/receptions", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("employee"))
			r.Post("/", handlers.CreateReceptionHandler(repo))
		})

		r.Route("/products", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("employee"))
			r.Post("/", handlers.AddProductHandler(repo))
		})

		r.Route("/pvz/{pvzId}", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("employee"))
			r.Post("/delete_last_product", handlers.DeleteLastProductHandler(repo))
			r.Post("/close_last_reception", handlers.CloseReceptionHandler(repo))
		})
	})

	return httptest.NewServer(r)
}

type MockRepository struct{}

func (m *MockRepository) CreatePVZ(city string) (*repository.PVZ, error) {
	return &repository.PVZ{
		ID:               "mocked-pvz-id",
		RegistrationDate: time.Now(),
		City:             city,
	}, nil
}

func (m *MockRepository) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]repository.PVZWithReceptions, error) {
	return []repository.PVZWithReceptions{}, nil
}

func (m *MockRepository) GetReceptionsForPVZ(pvzID string) ([]repository.ReceptionWithProducts, error) {
	return []repository.ReceptionWithProducts{}, nil
}

func (m *MockRepository) GetProductsForReception(receptionID string) ([]repository.Product, error) {
	return []repository.Product{}, nil
}

func (m *MockRepository) CreateReception(pvzID string) (*repository.Reception, error) {
	return &repository.Reception{
		ID:       "mocked-reception-id",
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   "in_progress",
	}, nil
}

func (m *MockRepository) CloseReception(pvzID string) (*repository.Reception, error) {
	return &repository.Reception{
		ID:       "mocked-reception-id",
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   "close",
	}, nil
}

func (m *MockRepository) AddProduct(productType, pvzID string) (*repository.Product, error) {
	return &repository.Product{
		ID:          "mocked-product-id",
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: "mocked-reception-id",
	}, nil
}

func (m *MockRepository) DeleteLastProduct(pvzID string) error {
	return nil
}

func (m *MockRepository) RegisterUser(email, password, role string) (*repository.User, error) {
	return &repository.User{
		ID:    "mocked-user-id",
		Email: email,
		Role:  role,
	}, nil
}

func (m *MockRepository) LoginUser(email, password string) (*repository.User, error) {
	return &repository.User{
		ID:    "mocked-user-id",
		Email: email,
		Role:  "employee",
	}, nil
}
func TestIntegrationScenario(t *testing.T) {
	mockRepo := &MockRepository{}
	server := MockServer(mockRepo)
	defer server.Close()

	moderatorToken := getTestToken(t, server.URL, "moderator")
	pvzID := createPVZ(t, server.URL, moderatorToken)

	employeeToken := getTestToken(t, server.URL, "employee")

	receptionID := createReception(t, server.URL, employeeToken, pvzID)

	for i := 0; i < 50; i++ {
		addProduct(t, server.URL, employeeToken, pvzID, receptionID)
	}

	closeReception(t, server.URL, employeeToken, pvzID)
}

func getTestToken(t *testing.T, baseURL, role string) string {
	reqBody := map[string]string{"role": role}
	reqJSON, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewBuffer(reqJSON))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])

	return response["token"]
}

func createPVZ(t *testing.T, baseURL, token string) string {
	reqBody := map[string]string{"city": "Москва"}
	reqJSON, _ := json.Marshal(reqBody)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+"/pvz", bytes.NewBuffer(reqJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	return response["id"].(string)
}

func createReception(t *testing.T, baseURL, token, pvzID string) string {
	reqBody := map[string]string{"pvzId": pvzID}
	reqJSON, _ := json.Marshal(reqBody)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+"/receptions", bytes.NewBuffer(reqJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&response)
	return response["id"].(string)
}

func addProduct(t *testing.T, baseURL, token, pvzID, receptionID string) {
	reqBody := map[string]string{"type": "электроника", "pvzId": pvzID}
	reqJSON, _ := json.Marshal(reqBody)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+"/products", bytes.NewBuffer(reqJSON))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func closeReception(t *testing.T, baseURL, token, pvzID string) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+"/pvz/"+pvzID+"/close_last_reception", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
