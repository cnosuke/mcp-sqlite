package tools

import (
	"database/sql"
	"encoding/json"

	"github.com/cockroachdb/errors"
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
			zap.S().Debugw("querying for tables", "query", query)
			rows, err := db.Query(query)
			if err != nil {
				zap.S().Errorw("failed to get table list", "error", err)
				return nil, errors.Wrap(err, "failed to get table list")
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
					return nil, errors.Wrap(err, "failed to scan table name")
				}
				tables = append(tables, tableName)
				tableCount++
			}
			zap.S().Debugw("found tables", "count", tableCount, "tables", tables)

			// Convert result to JSON
			jsonResult, err := json.Marshal(tables)
			if err != nil {
				zap.S().Errorw("failed to convert result to JSON", "error", err)
				return nil, errors.Wrap(err, "failed to convert result to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register list_tables tool", "error", err)
		return errors.Wrap(err, "failed to register list_tables tool")
	}

	return nil
}
