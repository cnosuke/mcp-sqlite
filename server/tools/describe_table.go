package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.uber.org/zap"
)

// DescribeTableArgs - Arguments for describe_table tool (kept for testing compatibility)
type DescribeTableArgs struct {
	TableName string `json:"table_name" jsonschema:"description=Name of table to describe"`
}

// RegisterDescribeTableTool - Register the describe_table tool
func RegisterDescribeTableTool(mcpServer *server.MCPServer, db *sql.DB) error {
	zap.S().Debug("registering describe_table tool")

	// Define the tool
	tool := mcp.NewTool("describe_table",
		mcp.WithDescription("View schema information for a specific table"),
		mcp.WithString("table_name",
			mcp.Description("Name of table to describe"),
			mcp.Required(),
		),
	)

	// Add the tool handler
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract table_name parameter
		tableName, ok := request.Params.Arguments["table_name"].(string)
		if !ok || tableName == "" {
			return mcp.NewToolResultError("table_name parameter is required"), nil
		}

		zap.S().Debugw("executing describe_table", "table_name", tableName)

		// Get table schema information
		query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		zap.S().Debugw("querying table schema", "query", query)
		rows, err := db.QueryContext(ctx, query)
		if err != nil {
			zap.S().Errorw("failed to get table information",
				"table_name", tableName,
				"error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer rows.Close()

		// Slice to store column information
		var columns []map[string]interface{}

		// Process each row
		columnCount := 0
		for rows.Next() {
			var cid int
			var name, dataType string
			var notNull, pk int
			var dfltValue interface{}

			if err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk); err != nil {
				zap.S().Errorw("failed to scan column information", "error", err)
				return mcp.NewToolResultError(err.Error()), nil
			}

			column := map[string]interface{}{
				"name":        name,
				"type":        dataType,
				"not_null":    notNull == 1,
				"default":     dfltValue,
				"primary_key": pk == 1,
			}
			columns = append(columns, column)
			columnCount++
			zap.S().Debugw("column found",
				"name", name,
				"type", dataType,
				"not_null", notNull == 1,
				"primary_key", pk == 1)
		}
		zap.S().Debugw("table schema retrieved",
			"table_name", tableName,
			"column_count", columnCount)

		// Convert result to JSON
		jsonResult, err := json.Marshal(columns)
		if err != nil {
			zap.S().Errorw("failed to convert result to JSON", "error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	return nil
}
