package tools

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// ReadQueryArgs - Arguments for read_query tool
type ReadQueryArgs struct {
	Query string `json:"query" jsonschema:"description=The SELECT SQL query to execute"`
}

// RegisterReadQueryTool - Register the read_query tool
func RegisterReadQueryTool(server *mcp.Server, db *sql.DB) error {
	zap.S().Debug("registering read_query tool")
	err := server.RegisterTool("read_query", "Execute SELECT queries to read data from the database",
		func(args ReadQueryArgs) (*mcp.ToolResponse, error) {
			zap.S().Debug("executing read_query", zap.String("query", args.Query))

			// Verify query starts with SELECT
			if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(args.Query)), "SELECT") {
				zap.S().Warn("invalid query type for read_query", zap.String("query", args.Query))
				return nil, errors.New("read_query only supports SELECT queries")
			}

			// Execute query
			zap.S().Debug("executing SELECT query", zap.String("query", args.Query))
			rows, err := db.Query(args.Query)
			if err != nil {
				zap.S().Error("failed to execute query",
					zap.String("query", args.Query),
					zap.Error(err))
				return nil, errors.Wrap(err, "failed to execute query")
			}
			defer rows.Close()

			// Get results
			columns, err := rows.Columns()
			if err != nil {
				zap.S().Error("failed to get column names", zap.Error(err))
				return nil, errors.Wrap(err, "failed to get column names")
			}
			zap.S().Debug("query columns", zap.Strings("columns", columns))

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
					zap.S().Error("failed to scan row",
						zap.Int("row", rowCount),
						zap.Error(err))
					return nil, errors.Wrap(err, "failed to scan row")
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
			zap.S().Debug("query completed", zap.Int("rows_returned", rowCount))

			// Convert results to JSON
			jsonResult, err := json.Marshal(results)
			if err != nil {
				zap.S().Error("failed to convert results to JSON", zap.Error(err))
				return nil, errors.Wrap(err, "failed to convert results to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Error("failed to register read_query tool", zap.Error(err))
		return errors.Wrap(err, "failed to register read_query tool")
	}

	return nil
}
