package repository

import (
	"context"
	"metricapp/internal/logger"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	psqlHandler *PsqlHandler
	once        sync.Once
)

type PsqlHandler struct {
	conn *pgx.Conn
}

func NewPsqlHandler(dsn string) {
	once.Do(func() {
		conn, err := pgx.Connect(context.Background(), dsn)
		if err != nil {
			logger.Error("failed to connect to db", zap.Error(err))
			return
		}

		psqlHandler = &PsqlHandler{
			conn: conn,
		}
	})
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return psqlHandler.conn.Ping(ctx)
}
