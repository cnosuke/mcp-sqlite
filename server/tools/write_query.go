package tools

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	mcp "github.com/metoro-io/mcp-golang"
	"go.uber.org/zap"
)

// WriteQueryArgs - Arguments for write_query tool
type WriteQueryArgs struct {
	Query string `json:"query" jsonschema:"description=The SQL write query (INSERT, UPDATE, DELETE) to execute"`
}

// RegisterWriteQueryTool - Register the write_query tool
func RegisterWriteQueryTool(server *mcp.Server, db *sql.DB) error {
	zap.S().Debug("registering write_query tool")
	err := server.RegisterTool("write_query", "Execute write queries (INSERT, UPDATE, DELETE) to modify data in the database",
		func(args WriteQueryArgs) (*mcp.ToolResponse, error) {
			zap.S().Debugw("executing write_query", "query", args.Query)

			// Verify query starts with appropriate write operation
			trimmedQuery := strings.ToUpper(strings.TrimSpace(args.Query))
			if !strings.HasPrefix(trimmedQuery, "INSERT") &&
				!strings.HasPrefix(trimmedQuery, "UPDATE") &&
				!strings.HasPrefix(trimmedQuery, "DELETE") {
				zap.S().Warnw("invalid query type for write_query", "query", args.Query)
				return nil, errors.New("write_query only supports INSERT, UPDATE, or DELETE queries")
			}

			// Execute query
			zap.S().Debugw("executing write query", "query", args.Query)
			result, err := db.Exec(args.Query)
			if err != nil {
				zap.S().Errorw("failed to execute write query",
					"query", args.Query,
					"error", err)
				return nil, errors.Wrap(err, "failed to execute write query")
			}

			// Get affected rows count
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				zap.S().Warnw("couldn't get rows affected", "error", err)
			}

			// Get last inserted ID for INSERT operations
			var lastInsertID int64
			if strings.HasPrefix(trimmedQuery, "INSERT") {
				lastInsertID, err = result.LastInsertId()
				if err != nil {
					zap.S().Warnw("couldn't get last insert id", "error", err)
				}
			}

			// Prepare response message
			var responseMessage string
			operation := "operation"
			if strings.HasPrefix(trimmedQuery, "INSERT") {
				operation = "insert"
				if lastInsertID > 0 {
					responseMessage = fmt.Sprintf("insert successful: %d rows affected, last insert id: %d", rowsAffected, lastInsertID)
				} else {
					responseMessage = fmt.Sprintf("insert successful: %d rows affected", rowsAffected)
				}
			} else if strings.HasPrefix(trimmedQuery, "UPDATE") {
				operation = "update"
				responseMessage = fmt.Sprintf("update successful: %d rows affected", rowsAffected)
			} else if strings.HasPrefix(trimmedQuery, "DELETE") {
				operation = "delete"
				responseMessage = fmt.Sprintf("delete successful: %d rows affected", rowsAffected)
			}

			zap.S().Infow("write query executed successfully",
				"operation", operation,
				"rows_affected", rowsAffected)

			return mcp.NewToolResponse(mcp.NewTextContent(responseMessage)), nil
		})

	if err != nil {
		zap.S().Errorw("failed to register write_query tool", "error", err)
		return errors.Wrap(err, "failed to register write_query tool")
	}

	return nil
}
