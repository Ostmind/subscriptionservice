package postgres

import (
	"context"
	"fmt"
	"github.com/Ostmind/subscriptionservice/internal/subscription/config"
	"github.com/Ostmind/subscriptionservice/internal/subscription/models"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(dbConfig config.DatabaseConfig) (*Storage, error) {
	db := &Storage{}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.DBUser, dbConfig.DBPassword, dbConfig.DBName, dbConfig.DBSSLMode)

	err := db.connect(psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error creating connection DB %w", models.ErrDBConnectionCreation)
	}

	return db, nil
}

func (store *Storage) Close() {
	store.DB.Close()
}

func (store *Storage) connect(connStr string) error {
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return fmt.Errorf("db.connect: %w", err)
	}

	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("db.connect pool ping: %w", err)
	}

	store.DB = pool

	return nil
}
