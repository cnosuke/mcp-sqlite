# MCP SQLite Server

MCP SQLite Server is a Go-based MCP server implementation that wraps an SQLite database, allowing MCP clients (e.g., Claude Desktop) to interact with SQLite via a standardized JSON‑RPC protocol.

## Features

- **MCP Compliance:** Provides a JSON‑RPC based interface for tool execution according to the MCP specification.
- **SQLite Operations:** Supports operations such as creating tables, describing table schemas, listing tables, and executing both read and write queries.

## Requirements

- Docker (recommended)

For local development:

- Go 1.24 or later
- SQLite (the database file will be created if it does not exist)
- GCC and development tools (for CGO compilation)

## Using with Docker (Recommended)

```bash
docker pull cnosuke/mcp-sqlite:latest

# Run with default SQLite database
docker run -i --rm cnosuke/mcp-sqlite:latest

# Run with a mounted SQLite database
docker run -i --rm -v /path/to/your/db:/app/sqlite.db cnosuke/mcp-sqlite:latest
```

### Using with Claude Desktop (Docker)

To integrate with Claude Desktop using Docker, add an entry to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "cnosuke/mcp-sqlite:latest"]
    }
  }
}
```

For persistent storage with a volume mount:

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "-v", "/path/to/your/db:/app/sqlite.db", "cnosuke/mcp-sqlite:latest"]
    }
  }
}
```

## Building and Running (Go Binary)

Alternatively, you can build and run the Go binary directly:

```bash
# Build the server
make bin/mcp-sqlite

# Run the server
./bin/mcp-sqlite server --config=config.yml
```

### Using with Claude Desktop (Go Binary)

To integrate with Claude Desktop using the Go binary, add an entry to your `claude_desktop_config.json` file:

```json
{
  "mcpServers": {
    "sqlite": {
      "command": "./bin/mcp-sqlite",
      "args": ["server", "--config", "config.yml"],
      "env": {
        "LOG_PATH": "",
        "SQLITE_PATH": "/path/to/your/sqlite.db"
      }
    }
  }
}
```

This configuration registers the MCP SQLite Server with Claude Desktop. By setting `LOG_PATH` to an empty string, the server will only output fatal errors, ensuring that non-critical logs do not interfere with the MCP protocol messages transmitted over stdio.

## Configuration

The server is configured via a YAML file (default: `config.yml`). For example:

```yaml
log: 'mcp-sqlite.log'
debug: false

sqlite:
  path: './sqlite.db'
```

Configuration options can also be specified via environment variables:

- `LOG_PATH`: Path to log file (empty string disables file logging)
- `DEBUG`: Enable debug logging (true/false)
- `SQLITE_PATH`: Path to SQLite database file

## Logging

Logs are directed to the file specified by the `log` config option or `LOG_PATH` environment variable. When this value is empty, only fatal errors will be output to stderr and all other logs are suppressed.

**Important:** When using the MCP server with a stdio transport, logging must not be directed to standard output because it would interfere with the MCP protocol communication. For Claude Desktop usage, it's recommended to set `LOG_PATH` to an empty string.

**Note for Docker:** The Docker image disables file logging by default, as container logs are typically collected from stdout/stderr.

## MCP Server Usage

MCP clients interact with the server by sending JSON‑RPC requests to execute various tools. The following MCP tools are supported:

- **create_table:** Executes a `CREATE TABLE` statement.
- **describe_table:** Retrieves schema details for a specific table.
- **list_tables:** Returns a list of all tables in the SQLite database.
- **read_query:** Executes `SELECT` queries and returns the result in JSON format.
- **write_query:** Executes write queries (such as `INSERT`, `UPDATE`, or `DELETE`).

## Command-Line Parameters

When starting the server, you can specify various settings:

```bash
./bin/mcp-sqlite server [options]
```

Options:

- `--config`, `-c`: Path to the configuration file (default: "config.yml").

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests for improvements or bug fixes. For major changes, open an issue first to discuss your ideas.

## License

This project is licensed under the MIT License.

**Author:** cnosuke ( [x.com/cnosuke](https://x.com/cnosuke) )
