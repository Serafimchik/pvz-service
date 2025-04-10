package repository

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	DB *pgx.Conn
}

func NewRepository(db *pgx.Conn) *Repository {
	return &Repository{DB: db}
}

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

func (r *Repository) CreatePVZ(city string) (*PVZ, error) {
	newPVZ := &PVZ{
		ID:               uuid.New().String(),
		RegistrationDate: time.Now(),
		City:             city,
	}

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Insert("pvz").
		Columns("id", "registration_date", "city").
		Values(newPVZ.ID, newPVZ.RegistrationDate, newPVZ.City).
		Suffix("RETURNING id, registration_date, city").
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&newPVZ.ID, &newPVZ.RegistrationDate, &newPVZ.City)
	if err != nil {
		return nil, err
	}

	return newPVZ, nil
}
