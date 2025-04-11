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
	db := config.NewDBConnection(dbConfig)

	repo := repository.NewRepository(db)

	r := chi.NewRouter()

	r.Use(middleware.LoggingMiddleware)

	r.Post("/dummyLogin", handlers.DummyLoginHandler)
	r.Post("/register", handlers.RegisterHandler(repo))
	r.Post("/login", handlers.LoginHandler(repo))

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Route("/pvz", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("moderator"))
			r.Post("/", handlers.CreatePVZHandler(repo))
		})

		r.Route("/receptions", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("employee"))
			r.Post("/", handlers.CreateReceptionHandler(repo))
		})

		r.Route("/products", func(r chi.Router) {
			r.Use(middleware.RoleMiddleware("employee"))
			r.Post("/", handlers.AddProductHandler(repo))
		})
	})

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
