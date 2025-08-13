package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"metricapp/internal/logger"
	models "metricapp/internal/model"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

var (
	psqlHandler *PsqlHandler
	once        sync.Once
)

var (
	ErrNoConnection = errors.New("there is no connection to db")
)

type PsqlHandler struct {
	pool *pgxpool.Pool
}

func NewPsqlHandler(dsn string, mPath string) {
	once.Do(func() {
		pool, err := pgxpool.New(context.Background(), dsn)
		if err != nil {
			logger.Error("failed to connect to db", zap.Error(err))
			return
		}

		if err := pool.Ping(context.Background()); err != nil {
			logger.Error("connection to db was not established", zap.Error(err))
			return
		}

		psqlHandler = &PsqlHandler{
			pool: pool,
		}

		err = migration(dsn, mPath)
		if err != nil {
			logger.Error("failed to make migration", zap.Error(err))
			return
		}
	})
}

// Миграция с помощью goose
func migration(dsn string, mPath string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect for migration: %w", err)
	}
	defer db.Close()

	// goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, mPath); err != nil {
		return fmt.Errorf("failed to update sceme: %w", err)
	}

	return nil
}

func Ping() error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return psqlHandler.pool.Ping(ctx)
}

func UpdateGauge(key string, value float64, opt ...transactionInfo) error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	query := `INSERT INTO
			metrics (id, mtype, value)
			VALUES
			($1, $2, $3)
			ON CONFLICT (id) DO
			UPDATE
			SET
			value = EXCLUDED.value;`

	var (
		rows pgconn.CommandTag
		err  error
	)

	if len(opt) > 0 {
		tInfo := opt[0]
		rows, err = tInfo.tx.Exec(
			tInfo.ctx,
			query,
			key, models.Gauge, value,
		)
	} else {
		rows, err = psqlHandler.pool.Exec(context.Background(),
			query,
			key, models.Gauge, value,
		)
	}

	if err != nil {
		return fmt.Errorf("failed to make query: %w", err)
	}

	if rows.RowsAffected() == 0 {
		return fmt.Errorf("rows is not affected")
	}

	return nil
}

func IncrementCounter(key string, delta int64, opt ...transactionInfo) error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	const query = `INSERT INTO metrics (id, mtype, delta)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE
		SET delta = metrics.delta + EXCLUDED.delta;`

	var (
		rows pgconn.CommandTag
		err  error
	)

	if len(opt) > 0 {
		tInfo := opt[0]
		rows, err = tInfo.tx.Exec(tInfo.ctx, query, key, models.Counter, delta)
	} else {
		rows, err = psqlHandler.pool.Exec(context.Background(), query, key, models.Counter, delta)
	}

	if err != nil {
		return fmt.Errorf("failed to update counter: %w", err)
	}
	if rows.RowsAffected() == 0 {
		return fmt.Errorf("rows is not affected")
	}
	return nil
}

type transactionInfo struct {
	tx  pgx.Tx
	ctx context.Context
}

func InsertBatch(ctx context.Context, metrics []models.Metrics) error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	tx, err := psqlHandler.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	tInfo := transactionInfo{
		tx:  tx,
		ctx: ctx,
	}

	for _, m := range metrics {
		var err error

		switch m.MType {
		case models.Gauge:
			err = UpdateGauge(m.ID, *m.Value, tInfo)
		case models.Counter:
			err = IncrementCounter(m.ID, *m.Delta, tInfo)
		}

		if err != nil {
			return fmt.Errorf("transaction aborted: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
