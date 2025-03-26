package tools

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// DescribeTableArgs - Arguments for describe_table tool
type DescribeTableArgs struct {
	TableName string `json:"table_name" jsonschema:"description=Name of table to describe"`
}

// RegisterDescribeTableTool - Register the describe_table tool
func RegisterDescribeTableTool(server *mcp.Server, db *sql.DB) error {
	zap.S().Debug("registering describe_table tool")
	err := server.RegisterTool("describe_table", "View schema information for a specific table",
		func(args DescribeTableArgs) (*mcp.ToolResponse, error) {
			zap.S().Debugw("executing describe_table", "table_name", args.TableName)

			// Get table schema information
			query := fmt.Sprintf("PRAGMA table_info(%s)", args.TableName)
			zap.S().Debugw("querying table schema", "query", query)
			rows, err := db.Query(query)
			if err != nil {
				zap.S().Errorw("failed to get table information",
					"table_name", args.TableName,
					"error", err)
				return nil, errors.Wrap(err, "failed to get table information")
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
					return nil, errors.Wrap(err, "failed to scan column information")
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
				"table_name", args.TableName,
				"column_count", columnCount)

			// Convert result to JSON
			jsonResult, err := json.Marshal(columns)
			if err != nil {
				zap.S().Errorw("failed to convert result to JSON", "error", err)
				return nil, errors.Wrap(err, "failed to convert result to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register describe_table tool", "error", err)
		return errors.Wrap(err, "failed to register describe_table tool")
	}

	return nil
}
