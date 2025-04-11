package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{DB: db}
}

type PVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type PVZWithReceptions struct {
	PVZ        PVZ                     `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

type ReceptionWithProducts struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
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

func (r *Repository) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]PVZWithReceptions, error) {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := queryBuilder.Select("DISTINCT pvz.id", "pvz.registration_date", "pvz.city").
		From("pvz").
		LeftJoin("receptions ON pvz.id = receptions.pvz_id").
		OrderBy("pvz.registration_date DESC")

	if startDate != nil && endDate != nil {
		query = query.Where(squirrel.And{
			squirrel.GtOrEq{"receptions.date_time": startDate},
			squirrel.LtOrEq{"receptions.date_time": endDate},
		})
	}

	query = query.Offset(uint64((page - 1) * limit)).Limit(uint64(limit))

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Printf("Failed to generate SQL query: %v", err)
		return nil, fmt.Errorf("failed to generate query for PVZ list: %w", err)
	}

	log.Printf("Executing SQL query: %s with args: %v", sqlQuery, args)

	rows, err := r.DB.Query(context.Background(), sqlQuery, args...)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return nil, fmt.Errorf("failed to fetch PVZ list: %w", err)
	}
	defer rows.Close()

	var pvzList []PVZWithReceptions
	for rows.Next() {
		var pvz PVZWithReceptions
		err := rows.Scan(&pvz.PVZ.ID, &pvz.PVZ.RegistrationDate, &pvz.PVZ.City)
		if err != nil {
			log.Printf("Failed to scan PVZ data: %v", err)
			return nil, fmt.Errorf("failed to scan PVZ data: %w", err)
		}

		log.Printf("Loading receptions for PVZ ID: %s", pvz.PVZ.ID)

		pvz.Receptions, err = r.GetReceptionsForPVZ(pvz.PVZ.ID)
		if err != nil {
			log.Printf("Failed to load receptions for PVZ ID %s: %v", pvz.PVZ.ID, err)
			return nil, fmt.Errorf("failed to fetch receptions for PVZ: %w", err)
		}

		pvzList = append(pvzList, pvz)
	}

	log.Printf("Fetched %d PVZ records", len(pvzList))
	return pvzList, nil
}

func (r *Repository) GetReceptionsForPVZ(pvzID string) ([]ReceptionWithProducts, error) {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Select("id", "date_time", "pvz_id", "status").
		From("receptions").
		Where(squirrel.Eq{"pvz_id": pvzID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query for receptions: %w", err)
	}

	rows, err := r.DB.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch receptions: %w", err)
	}
	defer rows.Close()

	var receptions []ReceptionWithProducts
	for rows.Next() {
		var reception ReceptionWithProducts
		err := rows.Scan(&reception.Reception.ID, &reception.Reception.DateTime, &reception.Reception.PVZID, &reception.Reception.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reception data: %w", err)
		}

		reception.Products, err = r.GetProductsForReception(reception.Reception.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch products for reception: %w", err)
		}

		receptions = append(receptions, reception)
	}

	return receptions, nil
}

func (r *Repository) GetProductsForReception(receptionID string) ([]Product, error) {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Select("id", "date_time", "type", "reception_id").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query for products: %w", err)
	}

	rows, err := r.DB.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.DateTime, &product.Type, &product.ReceptionID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product data: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}
