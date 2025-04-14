package handlers

import (
	"pvz-service/internal/repository"
	"time"

	"github.com/stretchr/testify/mock"
)

func parseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(err)
	}
	return t
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) RegisterUser(email, password, role string) (*repository.User, error) {
	args := m.Called(email, password, role)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockRepository) LoginUser(email, password string) (*repository.User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockRepository) CreatePVZ(city string) (*repository.PVZ, error) {
	args := m.Called(city)
	return args.Get(0).(*repository.PVZ), args.Error(1)
}

func (m *MockRepository) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]repository.PVZWithReceptions, error) {
	args := m.Called(startDate, endDate, page, limit)
	return args.Get(0).([]repository.PVZWithReceptions), args.Error(1)
}

func (m *MockRepository) CreateReception(pvzID string) (*repository.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*repository.Reception), args.Error(1)
}

func (m *MockRepository) CloseReception(pvzID string) (*repository.Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*repository.Reception), args.Error(1)
}

func (m *MockRepository) AddProduct(productType, pvzID string) (*repository.Product, error) {
	args := m.Called(productType, pvzID)
	return args.Get(0).(*repository.Product), args.Error(1)
}

func (m *MockRepository) DeleteLastProduct(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *MockRepository) GetProductsForReception(receptionID string) ([]repository.Product, error) {
	args := m.Called(receptionID)
	return args.Get(0).([]repository.Product), args.Error(1)
}

func (m *MockRepository) GetReceptionsForPVZ(pvzID string) ([]repository.ReceptionWithProducts, error) {
	args := m.Called(pvzID)
	return args.Get(0).([]repository.ReceptionWithProducts), args.Error(1)
}
