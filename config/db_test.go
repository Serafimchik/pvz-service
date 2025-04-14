package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPoolCreator struct {
	mock.Mock
}

func (m *MockPoolCreator) NewPool(config *pgxpool.Config) (DBConnection, error) {
	args := m.Called(config)
	return args.Get(0).(DBConnection), args.Error(1)
}

type MockPgxRows struct {
	mock.Mock
	ColumnsData [][]interface{}
	CurrentRow  int
	Closed      bool
}

func (m *MockPgxRows) Close() {
	m.Called()
	m.Closed = true
}

func (m *MockPgxRows) Next() bool {
	if m.CurrentRow >= len(m.ColumnsData) {
		return false
	}
	m.CurrentRow++
	return true
}

func (m *MockPgxRows) Scan(dest ...interface{}) error {
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
func (m *MockPgxRows) Values() ([]interface{}, error) {
	if m.CurrentRow == 0 || m.CurrentRow > len(m.ColumnsData) {
		return nil, fmt.Errorf("no rows to fetch values")
	}
	return m.ColumnsData[m.CurrentRow-1], nil
}
func (m *MockPgxRows) RawValues() [][]byte {
	if m.CurrentRow == 0 || m.CurrentRow > len(m.ColumnsData) {
		return nil
	}
	row := m.ColumnsData[m.CurrentRow-1]
	rawValues := make([][]byte, len(row))
	for i, col := range row {
		switch v := col.(type) {
		case string:
			rawValues[i] = []byte(v)
		case int:
			rawValues[i] = []byte(fmt.Sprintf("%d", v))
		default:
			rawValues[i] = []byte(fmt.Sprintf("%v", v))
		}
	}
	return rawValues
}
func (m *MockPgxRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockPgxRows) FieldDescriptions() []pgconn.FieldDescription {
	return []pgconn.FieldDescription{}
}

func (m *MockPgxRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPgxRows) Conn() *pgx.Conn {
	return nil
}

type MockRow struct {
	Columns []interface{}
}

func (m *MockRow) Scan(dest ...interface{}) error {
	if len(dest) != len(m.Columns) {
		return fmt.Errorf("number of arguments in Scan does not match number of columns")
	}
	for i, col := range m.Columns {
		val := reflect.ValueOf(dest[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("non-pointer passed to Scan")
		}
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(col))
	}
	return nil
}

type MockPool struct {
	mock.Mock
}

func (m *MockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgx.Row)
}

func (m *MockPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	argsMock := m.Called(ctx, sql, args)
	return argsMock.Get(0).(pgx.Rows), argsMock.Error(1)
}

func (m *MockPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	argsMock := m.Called(ctx, sql, arguments)
	return argsMock.Get(0).(pgconn.CommandTag), argsMock.Error(1)
}

func (m *MockPool) Close() {
	m.Called()
}

func TestQueryRow_Success(t *testing.T) {
	mockPool := new(MockPool)
	mockRow := &MockRow{
		Columns: []interface{}{1, "test@example.com"},
	}
	mockPool.On("QueryRow", mock.Anything, "SELECT id FROM users WHERE email = $1", mock.Anything).
		Return(mockRow)
	adapter := &PoolAdapter{
		pool: mockPool,
	}

	row := adapter.QueryRow(context.Background(), "SELECT id FROM users WHERE email = $1", "test@example.com")

	assert.NotNil(t, row)

	mockPool.AssertNumberOfCalls(t, "QueryRow", 1)
}
func TestPoolAdapter_QueryRow(t *testing.T) {
	mockPool := new(MockPool)

	mockRow := &MockRow{
		Columns: []interface{}{1, "test@example.com"},
	}

	mockPool.On("QueryRow", mock.Anything, "SELECT id, email FROM users WHERE email = $1", mock.Anything).
		Return(mockRow)

	adapter := &PoolAdapter{
		pool: mockPool,
	}

	row := adapter.QueryRow(context.Background(), "SELECT id, email FROM users WHERE email = $1", "test@example.com")

	var id int
	var email string
	err := row.Scan(&id, &email)

	assert.NoError(t, err)
	assert.Equal(t, 1, id)
	assert.Equal(t, "test@example.com", email)

	mockPool.AssertNumberOfCalls(t, "QueryRow", 1)
}
func TestPoolAdapter_Query(t *testing.T) {
	mockPool := new(MockPool)

	mockRows := &MockPgxRows{
		ColumnsData: [][]interface{}{
			{1, "test@example.com"},
			{2, "admin@example.com"},
		},
	}

	mockPool.On("Query", mock.Anything, "SELECT id, email FROM users", mock.Anything).
		Return(mockRows, nil)

	adapter := &PoolAdapter{
		pool: mockPool,
	}

	rows, err := adapter.Query(context.Background(), "SELECT id, email FROM users")

	assert.NoError(t, err)
	assert.NotNil(t, rows)

	var id int
	var email string
	for rows.Next() {
		err := rows.Scan(&id, &email)
		assert.NoError(t, err)
		t.Logf("Row: id=%d, email=%s", id, email)
	}

	mockPool.AssertNumberOfCalls(t, "Query", 1)
}

func TestPoolAdapter_Exec(t *testing.T) {
	mockPool := new(MockPool)

	mockPool.On("Exec", mock.Anything, "INSERT INTO users (email, name) VALUES ($1, $2)", mock.Anything).
		Return(pgconn.NewCommandTag("INSERT 1"), nil)

	adapter := &PoolAdapter{
		pool: mockPool,
	}

	tag, err := adapter.Exec(context.Background(), "INSERT INTO users (email, name) VALUES ($1, $2)", "test@example.com", "John Doe")
	assert.NoError(t, err)
	assert.NotNil(t, tag)
	assert.Equal(t, int64(1), tag.RowsAffected())

	mockPool.AssertNumberOfCalls(t, "Exec", 1)
}
func TestNewDBConnection_Success(t *testing.T) {
	mockPoolCreator := new(MockPoolCreator)

	mockPoolCreator.On("NewPool", mock.Anything).
		Return(&PoolAdapter{}, nil)

	config := DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "dbname",
		SSLMode:  "disable",
	}

	dbConn := NewDBConnection(config, mockPoolCreator)

	assert.NotNil(t, dbConn)

	mockPoolCreator.AssertNumberOfCalls(t, "NewPool", 1)
}
func TestNewDBConnection_Error(t *testing.T) {
	mockPoolCreator := new(MockPoolCreator)

	mockPoolCreator.On("NewPool", mock.Anything).
		Return(nil, errors.New("connection error"))

	config := DBConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "user",
		Password: "password",
		DBName:   "dbname",
		SSLMode:  "disable",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, but no panic occurred")
		}
	}()
	NewDBConnection(config, mockPoolCreator)

	mockPoolCreator.AssertNumberOfCalls(t, "NewPool", 1)
}
func TestRealPoolCreator_NewPool_Success(t *testing.T) {
	poolCreator := &RealPoolCreator{}

	poolConfig, err := pgxpool.ParseConfig("host=localhost port=5432 user=user password=password dbname=pvz sslmode=disable")
	assert.NoError(t, err)

	dbConn, err := poolCreator.NewPool(poolConfig)

	assert.NoError(t, err)
	assert.NotNil(t, dbConn)

	dbConn.(*PoolAdapter).pool.Close()
}
func TestRealPoolCreator_NewPool_Error(t *testing.T) {
	poolCreator := &RealPoolCreator{}
	poolConfig, err := pgxpool.ParseConfig("invalid connection string")
	assert.Error(t, err, "Expected error from ParseConfig")

	if err != nil {
		return
	}

	dbConn, err := poolCreator.NewPool(poolConfig)

	assert.Error(t, err, "Expected error from NewPool")
	assert.Nil(t, dbConn, "Expected dbConn to be nil")
}
func TestPgxRowsWrapper_Close(t *testing.T) {
	mockRows := new(MockPgxRows)
	mockRows.On("Close").Return()

	wrapper := &PgxRowsWrapper{rows: mockRows}

	err := wrapper.Close()

	assert.NoError(t, err)

	mockRows.AssertNumberOfCalls(t, "Close", 1)
}
