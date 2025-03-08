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
			zap.S().Debug("executing create_table", zap.String("query", args.Query))

			// Verify query starts with CREATE TABLE
			if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(args.Query)), "CREATE TABLE") {
				zap.S().Warn("invalid query type for create_table", zap.String("query", args.Query))
				return nil, errors.New("create_table only supports CREATE TABLE statements")
			}

			// Execute query
			zap.S().Debug("creating table", zap.String("query", args.Query))
			_, err := db.Exec(args.Query)
			if err != nil {
				zap.S().Error("failed to create table",
					zap.String("query", args.Query),
					zap.Error(err))
				return nil, errors.Wrap(err, "failed to create table")
			}

			// Extract table name (simple implementation)
			parts := strings.Split(args.Query, "CREATE TABLE")
			if len(parts) < 2 {
				zap.S().Error("could not extract table name", zap.String("query", args.Query))
				return nil, errors.New("could not extract table name")
			}
			tablePart := strings.TrimSpace(parts[1])
			tableName := strings.Split(tablePart, " ")[0]
			tableName = strings.Trim(tableName, "`[]\"' ")
			zap.S().Info("table created successfully", zap.String("table_name", tableName))

			return mcp.NewToolResponse(mcp.NewTextContent(fmt.Sprintf("Table '%s' was successfully created", tableName))), nil
		})

	if err != nil {
		zap.S().Error("failed to register create_table tool", zap.Error(err))
		return errors.Wrap(err, "failed to register create_table tool")
	}

	return nil
}
