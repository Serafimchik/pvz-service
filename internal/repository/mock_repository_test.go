package repository

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"pvz-service/config"

	"github.com/stretchr/testify/mock"
)

type MockRow struct {
	mock.Mock
	ScanData []interface{}
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

type MockDBConnection struct {
	mock.Mock
}

func (m *MockDBConnection) QueryRow(ctx context.Context, sql string, args ...interface{}) config.Row {
	argsCalled := m.Called(ctx, sql, args)
	return argsCalled.Get(0).(config.Row)
}

func (m *MockDBConnection) Query(ctx context.Context, sql string, args ...interface{}) (config.Rows, error) {
	argsCalled := m.Called(ctx, sql, args)
	return argsCalled.Get(0).(config.Rows), argsCalled.Error(1)
}

func (m *MockDBConnection) Exec(ctx context.Context, sql string, arguments ...interface{}) (config.CommandTag, error) {
	argsCalled := m.Called(ctx, sql, arguments)
	return argsCalled.Get(0).(config.CommandTag), argsCalled.Error(1)
}

type MockHasher struct {
	Err error
}

func (m *MockHasher) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return []byte("mocked-hash"), nil
}

type MockCommandTag struct {
	mock.Mock
}

func (m *MockCommandTag) RowsAffected() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

type MockRows struct {
	mock.Mock
	ColumnsData    [][]interface{}
	CurrentRow     int
	IterationError error
}

func (m *MockRows) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) Next() bool {
	if m.CurrentRow >= len(m.ColumnsData) {
		return false
	}
	m.CurrentRow++
	return true
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.CurrentRow == 0 || m.CurrentRow > len(m.ColumnsData) {
		return fmt.Errorf("no rows to scan")
	}
	row := m.ColumnsData[m.CurrentRow-1]
	if len(dest) != len(row) {
		return fmt.Errorf("number of arguments in Scan does not match number of columns")
	}
	for i, col := range row {
		val := reflect.ValueOf(dest[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("non-pointer passed to Scan")
		}
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	}
	return nil
}
func (m *MockRows) Err() error {
	return m.IterationError
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreatePVZ(city string) (*PVZ, error) {
	args := m.Called(city)
	return args.Get(0).(*PVZ), args.Error(1)
}

func (m *MockRepository) GetPVZList(startDate, endDate *time.Time, page, limit int) ([]PVZWithReceptions, error) {
	args := m.Called(startDate, endDate, page, limit)
	return args.Get(0).([]PVZWithReceptions), args.Error(1)
}

func (m *MockRepository) GetReceptionsForPVZ(pvzID string) ([]ReceptionWithProducts, error) {
	args := m.Called(pvzID)
	return args.Get(0).([]ReceptionWithProducts), args.Error(1)
}

func (m *MockRepository) GetProductsForReception(receptionID string) ([]Product, error) {
	args := m.Called(receptionID)
	return args.Get(0).([]Product), args.Error(1)
}

func (m *MockRepository) CreateReception(pvzID string) (*Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*Reception), args.Error(1)
}

func (m *MockRepository) CloseReception(pvzID string) (*Reception, error) {
	args := m.Called(pvzID)
	return args.Get(0).(*Reception), args.Error(1)
}

func (m *MockRepository) AddProduct(productType, pvzID string) (*Product, error) {
	args := m.Called(productType, pvzID)
	return args.Get(0).(*Product), args.Error(1)
}

func (m *MockRepository) DeleteLastProduct(pvzID string) error {
	args := m.Called(pvzID)
	return args.Error(0)
}

func (m *MockRepository) RegisterUser(email, password, role string) (*User, error) {
	args := m.Called(email, password, role)
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) LoginUser(email, password string) (*User, error) {
	args := m.Called(email, password)
	return args.Get(0).(*User), args.Error(1)
}
