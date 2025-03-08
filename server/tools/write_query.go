package tools

import (
	"database/sql"
	"fmt"
	"strings"

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
			zap.S().Debug("executing write_query", zap.String("query", args.Query))

			// Verify query starts with appropriate write operation
			trimmedQuery := strings.ToUpper(strings.TrimSpace(args.Query))
			if !strings.HasPrefix(trimmedQuery, "INSERT") &&
				!strings.HasPrefix(trimmedQuery, "UPDATE") &&
				!strings.HasPrefix(trimmedQuery, "DELETE") {
				zap.S().Warn("invalid query type for write_query", zap.String("query", args.Query))
				return nil, fmt.Errorf("write_query only supports INSERT, UPDATE, or DELETE queries")
			}

			// Execute query
			zap.S().Debug("executing write query", zap.String("query", args.Query))
			result, err := db.Exec(args.Query)
			if err != nil {
				zap.S().Error("failed to execute write query",
					zap.String("query", args.Query),
					zap.Error(err))
				return nil, fmt.Errorf("failed to execute write query: %w", err)
			}

			// Get affected rows count
			rowsAffected, err := result.RowsAffected()
			if err != nil {
				zap.S().Warn("couldn't get rows affected", zap.Error(err))
			}

			// Get last inserted ID for INSERT operations
			var lastInsertID int64
			if strings.HasPrefix(trimmedQuery, "INSERT") {
				lastInsertID, err = result.LastInsertId()
				if err != nil {
					zap.S().Warn("couldn't get last insert id", zap.Error(err))
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

			zap.S().Info("write query executed successfully",
				zap.String("operation", operation),
				zap.Int64("rows_affected", rowsAffected))

			return mcp.NewToolResponse(mcp.NewTextContent(responseMessage)), nil
		})

	if err != nil {
		zap.S().Error("failed to register write_query tool", zap.Error(err))
		return fmt.Errorf("failed to register write_query tool: %w", err)
	}

	return nil
}
