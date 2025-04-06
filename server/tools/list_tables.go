package tools

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// ListTablesArgs - Arguments for list_tables tool (kept for testing compatibility)
type ListTablesArgs struct {
	// No arguments needed
}

// RegisterListTablesTools - Register the list_tables tool
func RegisterListTablesTools(mcpServer *server.MCPServer, db *sql.DB) error {
	zap.S().Debug("registering list_tables tool")

	// Define the tool (no parameters needed)
	tool := mcp.NewTool("list_tables",
		mcp.WithDescription("Get a list of all tables in the database"),
	)

	// Add the tool handler
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		zap.S().Debug("executing list_tables")

		// Get table list from SQLite system tables
		const query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
		zap.S().Debugw("querying for tables", "query", query)
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			zap.S().Errorw("failed to get table list", "error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer rows.Close()

		// Slice to store table names
		var tables []string

		// Process each row
		tableCount := 0
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				zap.S().Errorw("failed to scan table name", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}
			tables = append(tables, tableName)
			tableCount++
		}
		zap.S().Debugw("found tables", "count", tableCount, "tables", tables)

		// Convert result to JSON
		jsonResult, err := json.Marshal(tables)
		if err != nil {
			zap.S().Errorw("failed to convert result to JSON", "error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	return nil
}
