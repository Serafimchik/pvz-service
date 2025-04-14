package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateReception_Success(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything).Return(errors.New("no rows"))

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*string) = "mocked-reception-id"
			*args.Get(1).(*time.Time) = time.Now()
			*args.Get(2).(*string) = "mocked-pvz-id"
			*args.Get(3).(*string) = "in_progress"
		}).Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	reception, err := repo.CreateReception("mocked-pvz-id")

	assert.NoError(t, err)
	assert.NotNil(t, reception)
	assert.Equal(t, "mocked-reception-id", reception.ID)
	assert.Equal(t, "mocked-pvz-id", reception.PVZID)
	assert.Equal(t, "in_progress", reception.Status)

	mockDB.AssertNumberOfCalls(t, "QueryRow", 2)
	mockRow.AssertNumberOfCalls(t, "Scan", 2)
}

func TestCreateReception_OpenReceptionExists(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(0).(*string) = "existing-reception-id"
		*args.Get(1).(*string) = "in_progress"
	}).Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	reception, err := repo.CreateReception("mocked-pvz-id")

	assert.Error(t, err)
	assert.Nil(t, reception)
	assert.Contains(t, err.Error(), "there is an open reception for this PVZ")

	mockDB.AssertCalled(t, "QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything)
	mockRow.AssertCalled(t, "Scan", mock.Anything, mock.Anything)
}
