package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"` // "электроника", "одежда", "обувь"
	ReceptionID string    `json:"receptionId"`
}

func (r *PostgresRepository) AddProduct(productType, pvzID string) (*Product, error) {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	var receptionID string
	query, args, err := queryBuilder.Select("id").
		From("receptions").
		Where(squirrel.Eq{"pvz_id": pvzID, "status": "in_progress"}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate query for active reception: %w", err)
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&receptionID)
	if err != nil {
		return nil, errors.New("no active reception found for this PVZ")
	}

	newProduct := &Product{
		ID:          uuid.New().String(),
		DateTime:    time.Now(),
		Type:        productType,
		ReceptionID: receptionID,
	}

	query, args, err = queryBuilder.Insert("products").
		Columns("id", "date_time", "type", "reception_id").
		Values(newProduct.ID, newProduct.DateTime, newProduct.Type, newProduct.ReceptionID).
		Suffix("RETURNING id, date_time, type, reception_id").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to generate insert query: %w", err)
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(
		&newProduct.ID, &newProduct.DateTime, &newProduct.Type, &newProduct.ReceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add product: %w", err)
	}

	return newProduct, nil
}

func (r *PostgresRepository) DeleteLastProduct(pvzID string) error {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	var receptionID string
	query, args, err := queryBuilder.Select("id").
		From("receptions").
		Where(squirrel.Eq{"pvz_id": pvzID, "status": "in_progress"}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate query for active reception: %w", err)
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&receptionID)
	if err != nil {
		return errors.New("no active reception found for this PVZ")
	}
	var productID string
	query, args, err = queryBuilder.Select("id").
		From("products").
		Where(squirrel.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate query for last product: %w", err)
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&productID)
	if err != nil {
		return errors.New("no products to delete in the current reception")
	}

	query, args, err = queryBuilder.Delete("products").
		Where(squirrel.Eq{"id": productID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate delete query: %w", err)
	}

	_, err = r.DB.Exec(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
