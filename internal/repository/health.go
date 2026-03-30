package repository

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lguilherme/contas/internal/domain"
)

type dbHealthChecker struct {
	db *pgxpool.Pool
}

func NewHealthChecker(db *pgxpool.Pool) domain.HealthChecker {
	return &dbHealthChecker{db: db}
}

func (h *dbHealthChecker) Ping(ctx context.Context) error {
	slog.Debug("repo: pinging database")
	return h.db.Ping(ctx)
}
