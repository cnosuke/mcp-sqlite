package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"

	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// ListTablesArgs - Arguments for list_tables tool (empty struct)
type ListTablesArgs struct {
	// No arguments needed
}

// RegisterListTablesTools - Register the list_tables tool
func RegisterListTablesTools(server *mcp.Server, db *sql.DB) error {
	zap.S().Debug("registering list_tables tool")
	err := server.RegisterTool("list_tables", "Get a list of all tables in the database",
		func(args ListTablesArgs) (*mcp.ToolResponse, error) {
			zap.S().Debug("executing list_tables")

			// Get table list from SQLite system tables
			const query = "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
			zap.S().Debug("querying for tables", zap.String("query", query))
			rows, err := db.Query(query)
			if err != nil {
				zap.S().Error("failed to get table list", zap.Error(err))
				return nil, fmt.Errorf("failed to get table list: %w", err)
			}
			defer rows.Close()

			// Slice to store table names
			var tables []string

			// Process each row
			tableCount := 0
			for rows.Next() {
				var tableName string
				if err := rows.Scan(&tableName); err != nil {
					zap.S().Error("failed to scan table name", zap.Error(err))
					return nil, fmt.Errorf("failed to scan table name: %w", err)
				}
				tables = append(tables, tableName)
				tableCount++
			}
			zap.S().Debug("found tables", zap.Int("count", tableCount), zap.Strings("tables", tables))

			// Convert result to JSON
			jsonResult, err := json.Marshal(tables)
			if err != nil {
				zap.S().Error("failed to convert result to JSON", zap.Error(err))
				return nil, fmt.Errorf("failed to convert result to JSON: %w", err)
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Error("failed to register list_tables tool", zap.Error(err))
		return fmt.Errorf("failed to register list_tables tool: %w", err)
	}

	return nil
}
