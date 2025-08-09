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

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
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
	conn *pgx.Conn
}

func NewPsqlHandler(dsn string, mPath string) {
	once.Do(func() {
		conn, err := pgx.Connect(context.Background(), dsn)
		if err != nil {
			logger.Error("failed to connect to db", zap.Error(err))
			return
		}

		if err := conn.Ping(context.Background()); err != nil {
			logger.Error("connection to db was not established", zap.Error(err))
			return
		}

		psqlHandler = &PsqlHandler{
			conn: conn,
		}

		err = migrationV2(dsn, mPath)
		if err != nil {
			logger.Error("failed to make migration", zap.Error(err))
			return
		}
	})
}

func migration(dsn string, mPath string) error {
	c, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect for migration: %w", err)
	}

	conn, err := c.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to db: %w", err)
	}

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file:///%s", mPath), "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	err = m.Up()
	if err != nil {
		return fmt.Errorf("failed to update scheme: %w", err)
	}

	return nil
}

// Миграция с помощью goose
func migrationV2(dsn string, mPath string) error {
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

	return psqlHandler.conn.Ping(ctx)
}

func UpdateGauge(key string, value float64) error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	rows, err := psqlHandler.conn.Exec(context.Background(),
		`INSERT INTO
		metrics (id, mtype, value)
		VALUES
		  ($1, $2, $3)
		ON CONFLICT (id, mtype) DO
		UPDATE
		SET
  		value = EXCLUDED.value;`,
		key, models.Gauge, value,
	)

	if err != nil {
		return fmt.Errorf("failed to make query: %w", err)
	}

	if rows.RowsAffected() == 0 {
		return fmt.Errorf("rows is not affected")
	}

	return nil
}

func IncrementCounter(key string, delta int64) error {
	if psqlHandler == nil {
		return ErrNoConnection
	}

	rows, err := psqlHandler.conn.Exec(context.Background(),
		`INSERT INTO public.metrics (id, mtype, delta)
		VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET delta = metrics.delta + EXCLUDED.delta`,
		key, models.Counter, delta,
	)

	if err != nil {
		return fmt.Errorf("failed to update counter: %w", err)
	}

	if rows.RowsAffected() == 0 {
		return fmt.Errorf("rows is not affected: %w", err)
	}

	return nil
}
