package repository

import (
	"context"
	"database/sql"
	"log"
	"metricapp/internal/logger"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var (
	psqlHandler *PsqlHandler
	once        sync.Once
)

type PsqlHandler struct {
	dbClient *sql.DB
}

func NewPsqlHandler(dsn string) {
	once.Do(func() {
		db, err := sql.Open("pgx", dsn)
		if err != nil {
			logger.Error("failed to connect to db", zap.Error(err))
			return
		}

		psqlHandler = &PsqlHandler{
			dbClient: db,
		}

		err = migration()
		if err != nil {
			log.Fatalf("failed to make migration: %v", err)
		}
	})
}

func migration() error {
	log.Println(psqlHandler.dbClient)
	conn, err := psqlHandler.dbClient.Conn(context.Background())
	if err != nil {
		return err
	}

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance("file:///app/migrations", "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	return nil
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return psqlHandler.dbClient.PingContext(ctx)
}
