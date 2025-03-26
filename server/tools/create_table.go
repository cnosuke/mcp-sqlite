package tools

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// CreateTableArgs - Arguments for create_table tool
type CreateTableArgs struct {
	Query string `json:"query" jsonschema:"description=CREATE TABLE SQL statement"`
}

// RegisterCreateTableTool - Register the create_table tool
func RegisterCreateTableTool(server *mcp.Server, db *sql.DB) error {
	zap.S().Debug("registering create_table tool")
	err := server.RegisterTool("create_table", "Create new tables in the database",
		func(args CreateTableArgs) (*mcp.ToolResponse, error) {
			zap.S().Debugw("executing create_table", "query", args.Query)

			// Verify query starts with CREATE TABLE
			if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(args.Query)), "CREATE TABLE") {
				zap.S().Warnw("invalid query type for create_table", "query", args.Query)
				return nil, errors.New("create_table only supports CREATE TABLE statements")
			}

			// Execute query
			zap.S().Debugw("creating table", "query", args.Query)
			_, err := db.Exec(args.Query)
			if err != nil {
				zap.S().Errorw("failed to create table",
					"query", args.Query,
					"error", err)
				return nil, errors.Wrap(err, "failed to create table")
			}

			// Extract table name (simple implementation)
			parts := strings.Split(args.Query, "CREATE TABLE")
			if len(parts) < 2 {
				zap.S().Errorw("could not extract table name", "query", args.Query)
				return nil, errors.New("could not extract table name")
			}
			tablePart := strings.TrimSpace(parts[1])
			tableName := strings.Split(tablePart, " ")[0]
			tableName = strings.Trim(tableName, "`[]\"' ")
			zap.S().Infow("table created successfully", "table_name", tableName)

			return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Table '%s' was successfully created", tableName))), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register create_table tool", "error", err)
		return errors.Wrap(err, "failed to register create_table tool")
	}

	return nil
}
