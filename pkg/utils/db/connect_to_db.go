package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nedokyrill/avito-pr-api/pkg/utils/logger"
)

func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, os.Getenv("DB_URL"))
	if err != nil {
		logger.Logger.Errorf("unable to connect to database: %v\n", err)
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		logger.Logger.Errorf("unable to ping database: %v\n", err)
		return nil, err
	}

	logger.Logger.Info("connected to Postgres successfully")
	return conn, nil
}
