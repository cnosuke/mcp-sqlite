package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// ReadQueryArgs - Arguments for read_query tool (kept for testing compatibility)
type ReadQueryArgs struct {
	Query string `json:"query" jsonschema:"description=The SELECT SQL query to execute"`
}

// RegisterReadQueryTool - Register the read_query tool
func RegisterReadQueryTool(mcpServer *server.MCPServer, db *sql.DB) error {
	zap.S().Debug("registering read_query tool")

	// Define the tool
	tool := mcp.NewTool("read_query",
		mcp.WithDescription("Execute SELECT queries to read data from the database"),
		mcp.WithString("query",
			mcp.Description("The SELECT SQL query to execute"),
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

		zap.S().Debugw("executing read_query", "query", query)

		// Verify query starts with SELECT
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
			zap.S().Warnw("invalid query type for read_query", "query", query)
			return mcp.NewToolResultError("read_query only supports SELECT queries"), nil
		}

		// Execute query
		zap.S().Debugw("executing SELECT query", "query", query)
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			zap.S().Errorw("failed to execute query",
				"query", query,
				"error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer rows.Close()

		// Get results
		columns, err := rows.Columns()
		if err != nil {
			zap.S().Errorw("failed to get column names", "error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		zap.S().Debugw("query columns", "columns", columns)

		// Slice to store results
		var results []map[string]interface{}

		// Process each row
		rowCount := 0
		for rows.Next() {
			rowCount++
			// Create scan destinations dynamically based on column count
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			// Scan the row
			if err := rows.Scan(valuePtrs...); err != nil {
				zap.S().Errorw("failed to scan row",
					"row", rowCount,
					"error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			// Convert row to map
			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				// Convert SQLite values to Go types
				switch v := val.(type) {
				case []byte:
					row[col] = string(v)
				default:
					row[col] = v
				}
			}
			results = append(results, row)
		}
		zap.S().Debugw("query completed", "rows_returned", rowCount)

		// Convert results to JSON
		jsonResult, err := json.Marshal(results)
		if err != nil {
			zap.S().Errorw("failed to convert results to JSON", "error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	return nil
}
