package repository

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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
			log.Fatalf("failed to connect to db: %v", err)
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
