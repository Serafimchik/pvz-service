package main

import (
	"log"
	"net/http"

	"pvz-service/config"
	"pvz-service/internal/handlers"
	"pvz-service/internal/middleware"
	"pvz-service/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	dbConfig := config.DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "pvz",
		SSLMode:  "disable",
	}
	poolCreator := &config.RealPoolCreator{}

	db := config.NewDBConnection(dbConfig, poolCreator)

	repo := repository.NewRepository(db)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)

	auth := middleware.NewJWTAuthenticator()
	r.Post("/dummyLogin", handlers.DummyLoginHandler(auth))
	r.Post("/register", handlers.RegisterHandler(repo))
	r.Post("/login", handlers.LoginHandler(repo, auth))

	r.Group(func(r chi.Router) {
		auth := middleware.NewJWTAuthenticator()
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

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
