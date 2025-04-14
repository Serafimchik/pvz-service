package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser_Success(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*string) = "mocked-id"
			*args.Get(1).(*string) = "test@example.com"
			*args.Get(3).(*string) = "employee"
		}).Return(nil)

	repo := &PostgresRepository{
		DB:     mockDB,
		Hasher: &MockHasher{},
	}

	user, err := repo.RegisterUser("test@example.com", "password123", "employee")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "mocked-id", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "employee", user.Role)

	mockDB.AssertCalled(t, "QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything)
}

func TestRegisterUser_InvalidRole(t *testing.T) {
	repo := &PostgresRepository{}

	user, err := repo.RegisterUser("test@example.com", "password123", "invalid-role")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestLoginUser_Success(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*string) = "mocked-id"
			*args.Get(1).(*string) = "test@example.com"
			*args.Get(2).(*string) = string(hashedPassword)
			*args.Get(3).(*string) = "employee"
		}).Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	user, err := repo.LoginUser("test@example.com", "password123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "mocked-id", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "employee", user.Role)

	mockDB.AssertCalled(t, "QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything)
}

func TestLoginUser_InvalidCredentials(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			*args.Get(0).(*string) = "mocked-id"
			*args.Get(1).(*string) = "test@example.com"
			*args.Get(2).(*string) = "wrong-hash"
			*args.Get(3).(*string) = "employee"
		}).Return(nil)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	user, err := repo.LoginUser("test@example.com", "wrong-password")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestLoginUser_UserNotFound(t *testing.T) {
	mockDB := new(MockDBConnection)
	mockRow := new(MockRow)

	mockDB.On("QueryRow", context.Background(), mock.AnythingOfType("string"), mock.Anything).
		Return(mockRow)

	mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pgx.ErrNoRows)

	repo := &PostgresRepository{
		DB: mockDB,
	}

	user, err := repo.LoginUser("nonexistent@example.com", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "invalid credentials")
}
