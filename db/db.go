package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/goloop/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"DB_HOST" def:"NONE"`
	Login    string `env:"DB_USER" def:"NONE"`
	Password string `env:"DB_PASSWORD" def:"NONE"`
	Database string `env:"DB_NAME" def:"NONE"`
	Port     string `env:"DB_PORT" def:"NONE"`
}

func Connect() (*pgxpool.Pool, error) {
	if err := env.Load(".env"); err != nil {
		return nil, fmt.Errorf("Unable to get DB parameters from .env")
	}

	var cfg Config
	if err := env.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("Unable to parse DB parameters")
	}
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.Login, cfg.Password, cfg.Host, cfg.Port, cfg.Database,
	)

	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		slog.Error("Error while connecting database", "error", err)
		return nil, fmt.Errorf("Could not connect to database %s\nError %s", dsn, err)
	}

	return conn, nil
}
