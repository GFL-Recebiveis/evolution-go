package whatsmeow_service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Antes, cada StartClient chamava sqlstore.New, abrindo um sql.DB novo (pool sem limite) que
// nunca era fechado: cada conexão/reconexão/QR vazava conexões ociosas no Postgres até
// estourar "too many connections". Agora há um único container (e um único pool limitado)
// compartilhado por todas as instâncias.
const (
	authMaxOpenConns    = 10
	authMaxIdleConns    = 2
	authConnMaxLifetime = 5 * time.Minute
	authConnMaxIdleTime = 1 * time.Minute
)

var (
	authContainerMu sync.Mutex
	authContainer   *sqlstore.Container
	authContainerDB *sql.DB
)

func getAuthContainer(ctx context.Context, dialect, address string, log waLog.Logger) (*sqlstore.Container, error) {
	authContainerMu.Lock()
	defer authContainerMu.Unlock()

	if authContainer != nil {
		return authContainer, nil
	}

	db, err := sql.Open(dialect, address)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if dialect == "postgres" {
		db.SetMaxOpenConns(authMaxOpenConns)
		db.SetMaxIdleConns(authMaxIdleConns)
		db.SetConnMaxLifetime(authConnMaxLifetime)
		db.SetConnMaxIdleTime(authConnMaxIdleTime)
	}

	container := sqlstore.NewWithDB(db, dialect, log)
	if err := container.Upgrade(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to upgrade database: %w", err)
	}

	authContainer = container
	authContainerDB = db
	return authContainer, nil
}
