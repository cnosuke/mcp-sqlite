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
			zap.S().Debug("executing describe_table", zap.String("table_name", args.TableName))

			// Get table schema information
			query := fmt.Sprintf("PRAGMA table_info(%s)", args.TableName)
			zap.S().Debug("querying table schema", zap.String("query", query))
			rows, err := db.Query(query)
			if err != nil {
				zap.S().Error("failed to get table information",
					zap.String("table_name", args.TableName),
					zap.Error(err))
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
					zap.S().Error("failed to scan column information", zap.Error(err))
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
				zap.S().Debug("column found",
					zap.String("name", name),
					zap.String("type", dataType),
					zap.Bool("not_null", notNull == 1),
					zap.Bool("primary_key", pk == 1))
			}
			zap.S().Debug("table schema retrieved",
				zap.String("table_name", args.TableName),
				zap.Int("column_count", columnCount))

			// Convert result to JSON
			jsonResult, err := json.Marshal(columns)
			if err != nil {
				zap.S().Error("failed to convert result to JSON", zap.Error(err))
				return nil, errors.Wrap(err, "failed to convert result to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Error("failed to register describe_table tool", zap.Error(err))
		return errors.Wrap(err, "failed to register describe_table tool")
	}

	return nil
}
