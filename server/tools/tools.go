package tools

import (
	"database/sql"

	"github.com/mark3labs/mcp-go/server"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(mcpServer *server.MCPServer, db *sql.DB) error {
	// Register read_query tool
	if err := RegisterReadQueryTool(mcpServer, db); err != nil {
		return err
	}

	// Register write_query tool
	if err := RegisterWriteQueryTool(mcpServer, db); err != nil {
		return err
	}

	// Register create_table tool
	if err := RegisterCreateTableTool(mcpServer, db); err != nil {
		return err
	}

	// Register list_tables tool
	if err := RegisterListTablesTools(mcpServer, db); err != nil {
		return err
	}

	// Register describe_table tool
	if err := RegisterDescribeTableTool(mcpServer, db); err != nil {
		return err
	}

	return nil
}
