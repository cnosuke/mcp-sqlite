package server

import (
	"database/sql"

	"github.com/cnosuke/mcp-sqlite/config"
	"github.com/cockroachdb/errors"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// SQLiteServer - SQLite server structure
type SQLiteServer struct {
	DB  *sql.DB
	cfg *config.Config
}

// NewSQLiteServer - Create a new SQLite server
func NewSQLiteServer(cfg *config.Config) (*SQLiteServer, error) {
	zap.S().Infow("creating new SQLite server",
		"database_path", cfg.SQLite.Path)

	db, err := sql.Open("sqlite3", cfg.SQLite.Path)
	if err != nil {
		zap.S().Errorw("failed to open SQLite database",
			"database_path", cfg.SQLite.Path,
			"error", err)
		return nil, errors.Wrap(err, "failed to open SQLite database")
	}

	// Connection test
	zap.S().Debug("testing database connection")
	if err := db.Ping(); err != nil {
		zap.S().Errorw("failed to connect to SQLite database",
			"database_path", cfg.SQLite.Path,
			"error", err)
		return nil, errors.Wrap(err, "failed to connect to SQLite database")
	}
	zap.S().Info("successfully connected to SQLite database")

	return &SQLiteServer{
		DB:  db,
		cfg: cfg,
	}, nil
}

// Close - Close the server
func (s *SQLiteServer) Close() error {
	zap.S().Info("closing SQLite server")
	return s.DB.Close()
}
