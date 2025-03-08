package server

import (
	"database/sql"
	"fmt"

	"github.com/cnosuke/mcp-notion/config"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// SQLiteServer - SQLite server structure
type SQLiteServer struct {
	DB     *sql.DB // 公開フィールドに変更
	cfg    *config.Config
	logger *zap.Logger
}

// NewSQLiteServer - Create a new SQLite server
func NewSQLiteServer(cfg *config.Config) (*SQLiteServer, error) {
	logger := zap.L() // グローバルロガーを使用
	zap.S().Info("creating new SQLite server",
		zap.String("database_path", cfg.SQLite.Path))

	db, err := sql.Open("sqlite3", cfg.SQLite.Path)
	if err != nil {
		zap.S().Error("failed to open SQLite database",
			zap.String("database_path", cfg.SQLite.Path),
			zap.Error(err))
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Connection test
	zap.S().Debug("testing database connection")
	if err := db.Ping(); err != nil {
		zap.S().Error("failed to connect to SQLite database",
			zap.String("database_path", cfg.SQLite.Path),
			zap.Error(err))
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}
	zap.S().Info("successfully connected to SQLite database")

	return &SQLiteServer{
		DB:     db,
		cfg:    cfg,
		logger: logger,
	}, nil
}

// Close - Close the server
func (s *SQLiteServer) Close() error {
	zap.S().Info("closing SQLite server")
	return s.DB.Close()
}
