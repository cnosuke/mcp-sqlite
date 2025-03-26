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
			zap.S().Debugw("executing read_query", "query", args.Query)

			// Verify query starts with SELECT
			if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(args.Query)), "SELECT") {
				zap.S().Warnw("invalid query type for read_query", "query", args.Query)
				return nil, errors.New("read_query only supports SELECT queries")
			}

			// Execute query
			zap.S().Debugw("executing SELECT query", "query", args.Query)
			rows, err := db.Query(args.Query)
			if err != nil {
				zap.S().Errorw("failed to execute query",
					"query", args.Query,
					"error", err)
				return nil, errors.Wrap(err, "failed to execute query")
			}
			defer rows.Close()

			// Get results
			columns, err := rows.Columns()
			if err != nil {
				zap.S().Errorw("failed to get column names", "error", err)
				return nil, errors.Wrap(err, "failed to get column names")
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
			zap.S().Debugw("query completed", "rows_returned", rowCount)

			// Convert results to JSON
			jsonResult, err := json.Marshal(results)
			if err != nil {
				zap.S().Errorw("failed to convert results to JSON", "error", err)
				return nil, errors.Wrap(err, "failed to convert results to JSON")
			}

			return mcp.NewToolResponse(mcp.NewTextContent(string(jsonResult))), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register read_query tool", "error", err)
		return errors.Wrap(err, "failed to register read_query tool")
	}

	return nil
}
