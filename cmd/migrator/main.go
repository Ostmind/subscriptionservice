package main

import (
	"github.com/Ostmind/subscriptionservice/internal/subscription/config"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.MustNew()

	connStr := config.GetConnStr(cfg.DB.Host, cfg.DB.Port,
		cfg.DB.DBUser, cfg.DB.DBPassword,
		cfg.DB.DBName, cfg.DB.DBSSLMode)

	connConfig, err := pgx.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("migrator: failed to parse conn config: %s", err)
	}

	db := stdlib.OpenDB(*connConfig)
	defer db.Close()

	if err = goose.SetDialect("postgres"); err != nil {
		log.Fatalf("migrator: goose error: %s", err)
	}

	if err = goose.Up(db, cfg.Srv.MigrationPath); err != nil {
		log.Fatalf("migrator: goose error: %s", err)
	}
}
