package tools

import (
	"database/sql"

	mcp "github.com/metoro-io/mcp-golang"
)

// RegisterAllTools - Register all tools with the server
func RegisterAllTools(server *mcp.Server, db *sql.DB) error {
	// Register read_query tool
	if err := RegisterReadQueryTool(server, db); err != nil {
		return err
	}

	// Register write_query tool
	if err := RegisterWriteQueryTool(server, db); err != nil {
		return err
	}

	// Register create_table tool
	if err := RegisterCreateTableTool(server, db); err != nil {
		return err
	}

	// Register list_tables tool
	if err := RegisterListTablesTools(server, db); err != nil {
		return err
	}

	// Register describe_table tool
	if err := RegisterDescribeTableTool(server, db); err != nil {
		return err
	}

	return nil
}
