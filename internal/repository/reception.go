package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type Reception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PVZID    string    `json:"pvzId"`
	Status   string    `json:"status"` // "in_progress" или "close"
}

func (r *Repository) CreateReception(pvzID string) (*Reception, error) {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Select("id", "status").
		From("receptions").
		Where(squirrel.Eq{"pvz_id": pvzID, "status": "in_progress"}).
		ToSql()
	if err != nil {
		return nil, err
	}

	var openReception Reception
	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&openReception.ID, &openReception.Status)
	if err == nil {
		return nil, errors.New("there is an open reception for this PVZ")
	}

	newReception := &Reception{
		ID:       uuid.New().String(),
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   "in_progress",
	}

	query, args, err = queryBuilder.Insert("receptions").
		Columns("id", "date_time", "pvz_id", "status").
		Values(newReception.ID, newReception.DateTime, newReception.PVZID, newReception.Status).
		Suffix("RETURNING id, date_time, pvz_id, status").
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(
		&newReception.ID, &newReception.DateTime, &newReception.PVZID, &newReception.Status,
	)
	if err != nil {
		return nil, err
	}

	return newReception, nil
}
