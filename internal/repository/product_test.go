package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddProduct_Success(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)
	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = "mocked-reception-id"
	}).Return(nil)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*string) = "mocked-product-id"
			*args.Get(1).(*time.Time) = time.Now()
			*args.Get(2).(*string) = "электроника"
			*args.Get(3).(*string) = "mocked-reception-id"
		}).Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	product, err := repo.AddProduct("электроника", "mocked-pvz-id")

	assert.NoError(t, err)
	assert.NotNil(t, product)
	assert.Equal(t, "mocked-product-id", product.ID)
	assert.Equal(t, "электроника", product.Type)
	assert.Equal(t, "mocked-reception-id", product.ReceptionID)

	mockDB.AssertNumberOfCalls(t, "QueryRow", 2)
	mockRow.AssertNumberOfCalls(t, "Scan", 2)
}
func TestAddProduct_NoActiveReception(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything).Return(errors.New("no rows"))

	repo := &PostgresRepository{
		DB: mockDB,
	}

	product, err := repo.AddProduct("электроника", "mocked-pvz-id")

	assert.Error(t, err)
	assert.Nil(t, product)
	assert.Contains(t, err.Error(), "no active reception found for this PVZ")

	mockDB.AssertCalled(t, "QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything)
	mockRow.AssertCalled(t, "Scan", mock.Anything)
}

func TestDeleteLastProduct_Success(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = "mocked-reception-id"
	}).Return(nil)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = "mocked-product-id"
	}).Return(nil)

	mockCommandTag := new(MockCommandTag)
	mockCommandTag.On("RowsAffected").Return(int64(1))

	mockDB.On("Exec", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockCommandTag, nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}
	err := repo.DeleteLastProduct("mocked-pvz-id")

	assert.NoError(t, err)

	mockDB.AssertNumberOfCalls(t, "QueryRow", 2)
	mockRow.AssertNumberOfCalls(t, "Scan", 2)
	mockDB.AssertCalled(t, "Exec", context.Background(), mock.AnythingOfType("string"), mock.Anything)
}

func TestDeleteLastProduct_NoActiveReception(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything).Return(errors.New("no rows"))

	repo := &PostgresRepository{
		DB: mockDB,
	}

	err := repo.DeleteLastProduct("mocked-pvz-id")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active reception found for this PVZ")

	mockDB.AssertCalled(t, "QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything)
	mockRow.AssertCalled(t, "Scan", mock.Anything)
}
