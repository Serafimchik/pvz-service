package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProductsForReception(t *testing.T) {
	mockDB := new(MockDBConnection)

	mockRows := &MockRows{
		ColumnsData: [][]interface{}{
			{"123e4567-e89b-12d3-a456-426614174000", time.Now(), "электроника", "123e4567-e89b-12d3-a456-426614174000"},
		},
	}

	mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

	mockRows.On("Close").Return(nil).Maybe()

	repo := NewRepository(mockDB)

	products, err := repo.GetProductsForReception("123e4567-e89b-12d3-a456-426614174000")

	assert.NoError(t, err)
	assert.Len(t, products, 1)
	assert.Equal(t, "электроника", products[0].Type)

	mockDB.AssertNumberOfCalls(t, "Query", 1)
	mockRows.AssertCalled(t, "Close")
}
func TestGetReceptionsForPVZ_Success(t *testing.T) {
	mockDB := new(MockDBConnection)

	mockRows := &MockRows{
		ColumnsData: [][]interface{}{
			{"rec-1", time.Now(), "pvz-1", "in_progress"},
		},
	}

	mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

	mockRows.On("Close").Return(nil)

	mockRows.On("Err").Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	receptions, err := repo.GetReceptionsForPVZ("pvz-1")

	assert.NoError(t, err)
	assert.Len(t, receptions, 1)

	assert.Equal(t, "rec-1", receptions[0].Reception.ID)
	assert.Equal(t, "pvz-1", receptions[0].Reception.PVZID)
	assert.Equal(t, "in_progress", receptions[0].Reception.Status)

	mockDB.AssertNumberOfCalls(t, "Query", 2)
	mockRows.AssertCalled(t, "Close")
}
