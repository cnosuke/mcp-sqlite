# MCP SQLite Server

MCP SQLite Server is a Go-based MCP server implementation that wraps an SQLite database, allowing MCP clients (e.g., Claude Desktop) to interact with SQLite via a standardized JSON‑RPC protocol.

## Features

- **MCP Compliance:** Provides a JSON‑RPC based interface for tool execution according to the MCP specification.
- **SQLite Operations:** Supports operations such as creating tables, describing table schemas, listing tables, and executing both read and write queries.

## Requirements

- Go 1.24 or later
- SQLite (the database file will be created if it does not exist)

## Configuration

The server is configured via a YAML file (default: `config.yml`). For example:

```yaml
sqlite:
  path: './sqlite.db'
```

**Note:** The SQLite file path can also be injected via an environment variable `SQLITE_PATH`. If this environment variable is set, it will override the value in the configuration file.

## Logging

Adjust logging behavior using the following command-line flags:

- `--no-logs`: Suppress non-critical logs.
- `--log`: Specify a file path to write logs.

**Important:** When using the MCP server with a stdio transport, logging must not be directed to standard output because it would interfere with the MCP protocol communication. Therefore, you should always use `--no-logs` along with `--log` to ensure that all logs are written exclusively to a log file.

## MCP Server Usage

MCP clients interact with the server by sending JSON‑RPC requests to execute various tools. The following MCP tools are supported:

- **create_table:** Executes a `CREATE TABLE` statement.
- **describe_table:** Retrieves schema details for a specific table.
- **list_tables:** Returns a list of all tables in the SQLite database.
- **read_query:** Executes `SELECT` queries and returns the result in JSON format.
- **write_query:** Executes write queries (such as `INSERT`, `UPDATE`, or `DELETE`).

### Using with Claude Desktop

To integrate with Claude Desktop, add an entry to your `claude_desktop_config.json` file. **Because MCP uses stdio for communication, you must redirect logs away from stdio by using the `--no-logs` and `--log` flags.** Below is an example configuration that injects the SQLite file path via an environment variable:

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "./bin/mcp-sqlite",
      "args": ["server", "--no-logs", "--log", "path_to_log_file"],
      "env": {
        "SQLITE_PATH": "/path/to/your/sqlite.db"
      }
    }
  }
}
```

This configuration registers the MCP SQLite Server with Claude Desktop, ensuring that all logs are directed to the specified log file rather than interfering with the MCP protocol messages transmitted over stdio.

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

**Author:** cnosuke ( [x.com/cnosuke](https://x.com/cnosuke) )
