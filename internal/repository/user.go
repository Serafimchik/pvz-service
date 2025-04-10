package repository

import (
	"context"
	"errors"
	"log"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (r *Repository) RegisterUser(email, password, role string) (*User, error) {
	if role != "employee" && role != "moderator" {
		return nil, errors.New("invalid role")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, errors.New("failed to hash password")
	}

	newUser := &User{
		ID:       uuid.New().String(),
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
	}

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Insert("users").
		Columns("id", "email", "password", "role").
		Values(newUser.ID, newUser.Email, newUser.Password, newUser.Role).
		Suffix("RETURNING id, email, password, role").
		ToSql()
	if err != nil {
		log.Printf("Error generating SQL query: %v", err)
		return nil, err
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&newUser.ID, &newUser.Email, &newUser.Password, &newUser.Role)
	if err != nil {
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	return newUser, nil
}

func (r *Repository) LoginUser(email, password string) (*User, error) {
	var user User

	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query, args, err := queryBuilder.Select("id", "email", "password", "role").
		From("users").
		Where(squirrel.Eq{"email": email}).
		ToSql()
	if err != nil {
		log.Printf("Error generating SQL query: %v", err)
		return nil, err
	}

	err = r.DB.QueryRow(context.Background(), query, args...).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("User not found for email: %s", email)
			return nil, errors.New("invalid credentials")
		}
		log.Printf("Error executing query: %v", err)
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("Password comparison failed for email: %s", email)
		return nil, errors.New("invalid credentials")
	}

	log.Printf("Login successful for email: %s", email)
	return &user, nil
}
