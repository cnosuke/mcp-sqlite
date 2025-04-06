package tools

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// CreateTableArgs - Arguments for create_table tool (kept for testing compatibility)
type CreateTableArgs struct {
	Query string `json:"query" jsonschema:"description=CREATE TABLE SQL statement"`
}

// RegisterCreateTableTool - Register the create_table tool
func RegisterCreateTableTool(mcpServer *server.MCPServer, db *sql.DB) error {
	zap.S().Debug("registering create_table tool")

	// Define the tool
	tool := mcp.NewTool("create_table",
		mcp.WithDescription("Create new tables in the database"),
		mcp.WithString("query",
			mcp.Description("CREATE TABLE SQL statement"),
			mcp.Required(),
		),
	)

	// Add the tool handler
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract query parameter
		query, ok := request.Params.Arguments["query"].(string)
		if !ok || query == "" {
			return mcp.NewToolResultError("query parameter is required"), nil
		}

		zap.S().Debugw("executing create_table", "query", query)

		// Verify query starts with CREATE TABLE
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "CREATE TABLE") {
			zap.S().Warnw("invalid query type for create_table", "query", query)
			return mcp.NewToolResultError("create_table only supports CREATE TABLE statements"), nil
		}

		// Execute query
		zap.S().Debugw("creating table", "query", query)
		_, err := db.ExecContext(ctx, query)
		if err != nil {
			zap.S().Errorw("failed to create table",
				"query", query,
				"error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract table name (simple implementation)
		parts := strings.Split(query, "CREATE TABLE")
		if len(parts) < 2 {
			zap.S().Errorw("could not extract table name", "query", query)
			return mcp.NewToolResultError("could not extract table name"), nil
		}
		tablePart := strings.TrimSpace(parts[1])
		tableName := strings.Split(tablePart, " ")[0]
		tableName = strings.Trim(tableName, "`[]\"' ")
		zap.S().Infow("table created successfully", "table_name", tableName)

		return mcp.NewToolResultText(fmt.Sprintf("Table '%s' was successfully created", tableName)), nil
	})

	return nil
}
