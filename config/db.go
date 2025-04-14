package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}
type PoolCreator interface {
	NewPool(config *pgxpool.Config) (DBConnection, error)
}

type DBConnection interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) Row
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (CommandTag, error)
}

type PgxRowsWrapper struct {
	rows pgx.Rows
}

func (w *PgxRowsWrapper) Close() error {
	w.rows.Close()
	return nil
}

func (w *PgxRowsWrapper) Next() bool {
	return w.rows.Next()
}

func (w *PgxRowsWrapper) Scan(dest ...interface{}) error {
	return w.rows.Scan(dest...)
}
func (p *PgxRowsWrapper) Err() error {
	return p.rows.Err()
}

type Row interface {
	Scan(dest ...interface{}) error
}

type Rows interface {
	Close() error
	Next() bool
	Scan(dest ...interface{}) error
	Err() error
}

type CommandTag interface {
	RowsAffected() int64
}

type PoolAdapter struct {
	pool PoolInterface
}
type PoolInterface interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Close()
}

func (p *PoolAdapter) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return p.pool.QueryRow(ctx, sql, args...)
}

func (p *PoolAdapter) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	pgxRows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &PgxRowsWrapper{rows: pgxRows}, nil
}

func (p *PoolAdapter) Exec(ctx context.Context, sql string, arguments ...interface{}) (CommandTag, error) {
	tag, err := p.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func NewDBConnection(config DBConfig, poolCreator PoolCreator) DBConnection {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	dbConn, err := poolCreator.NewPool(poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	log.Println("Successfully connected to the database!")
	return dbConn
}

type RealPoolCreator struct{}

func (r *RealPoolCreator) NewPool(config *pgxpool.Config) (DBConnection, error) {
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &PoolAdapter{pool: pool}, nil
}
