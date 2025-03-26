package server

import (
	mcp "github.com/metoro-io/mcp-golang"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"go.uber.org/zap"

	"github.com/cnosuke/mcp-sqlite/config"
	"github.com/cnosuke/mcp-sqlite/server/tools"
	"github.com/cockroachdb/errors"
)

// Run - Execute the MCP server
func Run(cfg *config.Config) error {
	zap.S().Info("starting MCP SQLite Server")

	// Channel to prevent server from terminating
	done := make(chan struct{})

	// Create SQLite server
	zap.S().Debug("creating SQLite server")
	sqliteServer, err := NewSQLiteServer(cfg)
	if err != nil {
		zap.S().Errorw("failed to create SQLite server", "error", err)
		return err
	}
	defer sqliteServer.Close()

	// Create server with stdio transport
	zap.S().Debug("creating MCP server with stdio transport")
	transport := stdio.NewStdioServerTransport()
	server := mcp.NewServer(transport)

	// Register all tools
	zap.S().Debug("registering tools")
	if err := tools.RegisterAllTools(server, sqliteServer.DB); err != nil {
		zap.S().Errorw("failed to register tools", "error", err)
		return err
	}

	// Start the server
	zap.S().Info("starting MCP server")
	err = server.Serve()
	if err != nil {
		zap.S().Errorw("failed to start server", "error", err)
		return errors.Wrap(err, "failed to start server")
	}

	zap.S().Infow("mcp SQLite server started successfully",
		"database_path", cfg.SQLite.Path)

	// Block to prevent program termination
	zap.S().Info("waiting for requests...")
	<-done
	zap.S().Info("server shutting down")
	return nil
}
