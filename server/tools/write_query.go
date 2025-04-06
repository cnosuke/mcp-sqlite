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

// WriteQueryArgs - Arguments for write_query tool (kept for testing compatibility)
type WriteQueryArgs struct {
	Query string `json:"query" jsonschema:"description=The SQL write query (INSERT, UPDATE, DELETE) to execute. Supports multiple statements separated by semicolons and transactions (BEGIN, COMMIT, ROLLBACK)"`
}

// statementResult - 各ステートメントの実行結果
type statementResult struct {
	Statement    string
	Operation    string
	Success      bool
	RowsAffected int64
	LastInsertID int64
	Error        error
}

// splitQueries - クエリを複数のステートメントに分割
func splitQueries(query string) []string {
	// セミコロンで分割
	statements := strings.Split(query, ";")

	// 空のステートメントを除去
	var result []string
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// isValidWriteOperation - 有効な書き込み操作かどうかを確認し、操作タイプを返す
func isValidWriteOperation(stmt string) (bool, string) {
	upper := strings.ToUpper(strings.TrimSpace(stmt))

	validTypes := []string{"INSERT", "UPDATE", "DELETE", "BEGIN", "COMMIT", "ROLLBACK"}
	for _, validType := range validTypes {
		if strings.HasPrefix(upper, validType) {
			return true, validType
		}
	}

	return false, ""
}

// executeStatements - 複数のステートメントを実行
func executeStatements(ctx context.Context, db *sql.DB, stmts []string) ([]statementResult, error) {
	var results []statementResult
	var tx *sql.Tx
	inTransaction := false

	for i, stmt := range stmts {
		valid, operationType := isValidWriteOperation(stmt)
		if !valid {
			return results, fmt.Errorf("statement %d is not a valid write operation: %s", i+1, stmt)
		}

		result := statementResult{
			Statement: stmt,
			Operation: operationType,
			Success:   false,
		}

		// トランザクション処理
		if operationType == "BEGIN" {
			if inTransaction {
				return results, fmt.Errorf("nested transactions are not supported")
			}

			var err error
			tx, err = db.BeginTx(ctx, nil)
			if err != nil {
				result.Error = err
				results = append(results, result)
				return results, fmt.Errorf("failed to begin transaction: %w", err)
			}

			inTransaction = true
			result.Success = true
			results = append(results, result)
			continue
		} else if operationType == "COMMIT" {
			if !inTransaction {
				return results, fmt.Errorf("COMMIT without BEGIN")
			}

			err := tx.Commit()
			if err != nil {
				result.Error = err
				results = append(results, result)
				return results, fmt.Errorf("failed to commit transaction: %w", err)
			}

			inTransaction = false
			tx = nil
			result.Success = true
			results = append(results, result)
			continue
		} else if operationType == "ROLLBACK" {
			if !inTransaction {
				return results, fmt.Errorf("ROLLBACK without BEGIN")
			}

			err := tx.Rollback()
			if err != nil {
				result.Error = err
				results = append(results, result)
				return results, fmt.Errorf("failed to rollback transaction: %w", err)
			}

			inTransaction = false
			tx = nil
			result.Success = true
			results = append(results, result)
			continue
		}

		// 通常のステートメント実行
		var sqlResult sql.Result
		var err error

		if inTransaction {
			sqlResult, err = tx.ExecContext(ctx, stmt)
		} else {
			sqlResult, err = db.ExecContext(ctx, stmt)
		}

		if err != nil {
			result.Error = err
			results = append(results, result)

			if inTransaction {
				tx.Rollback()
				inTransaction = false
				tx = nil
			}

			return results, fmt.Errorf("failed to execute statement %d: %w", i+1, err)
		}

		// 結果の処理
		rowsAffected, _ := sqlResult.RowsAffected()
		result.RowsAffected = rowsAffected

		if operationType == "INSERT" {
			lastInsertID, _ := sqlResult.LastInsertId()
			result.LastInsertID = lastInsertID
		}

		result.Success = true
		results = append(results, result)
	}

	// トランザクションが閉じられていない場合
	if inTransaction {
		return results, fmt.Errorf("transaction was not committed or rolled back")
	}

	return results, nil
}

// formatResponse - 実行結果をフォーマット
func formatResponse(results []statementResult) string {
	totalStatements := len(results)
	successfulStatements := 0
	totalRowsAffected := int64(0)
	lastInsertID := int64(0)

	for _, result := range results {
		if result.Success {
			successfulStatements++
			totalRowsAffected += result.RowsAffected

			if result.Operation == "INSERT" && result.LastInsertID > 0 {
				lastInsertID = result.LastInsertID
			}
		}
	}

	var response string
	if successfulStatements == totalStatements {
		response = fmt.Sprintf("Successfully executed %d statements. Total rows affected: %d",
			totalStatements, totalRowsAffected)

		if lastInsertID > 0 {
			response += fmt.Sprintf(", last insert ID: %d", lastInsertID)
		}
	} else {
		response = fmt.Sprintf("Executed %d out of %d statements. Total rows affected: %d",
			successfulStatements, totalStatements, totalRowsAffected)

		if lastInsertID > 0 {
			response += fmt.Sprintf(", last insert ID: %d", lastInsertID)
		}

		// エラー情報を追加
		for i, result := range results {
			if !result.Success {
				response += fmt.Sprintf("\nError in statement %d: %s", i+1, result.Error)
			}
		}
	}

	return response
}

// RegisterWriteQueryTool - Register the write_query tool
func RegisterWriteQueryTool(mcpServer *server.MCPServer, db *sql.DB) error {
	zap.S().Debug("registering write_query tool")

	// Define the tool
	tool := mcp.NewTool("write_query",
		mcp.WithDescription("Execute write queries (INSERT, UPDATE, DELETE) to modify data in the database. Supports multiple statements separated by semicolons and transactions (BEGIN, COMMIT, ROLLBACK)"),
		mcp.WithString("query",
			mcp.Description("The SQL write query (INSERT, UPDATE, DELETE) to execute"),
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

		zap.S().Debugw("executing write_query", "query", query)

		// Split query into multiple statements
		statements := splitQueries(query)
		if len(statements) == 0 {
			zap.S().Warnw("empty query", "query", query)
			return mcp.NewToolResultError("empty query"), nil
		}

		// Execute statements
		zap.S().Debugw("executing statements", "count", len(statements))
		results, err := executeStatements(ctx, db, statements)
		if err != nil {
			zap.S().Errorw("failed to execute statements",
				"error", err)
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format response
		responseMessage := formatResponse(results)
		zap.S().Infow("statements executed",
			"total", len(statements),
			"successful", len(results))

		return mcp.NewToolResultText(responseMessage), nil
	})

	return nil
}
