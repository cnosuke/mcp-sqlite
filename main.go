package main

import (
	"fmt"
	"os"

	"github.com/cnosuke/mcp-notion/config"
	"github.com/cnosuke/mcp-notion/logger"
	"github.com/cnosuke/mcp-notion/server"
	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
)

var (
	// Version and Revision are replaced when building.
	// To set specific version, edit Makefile.
	Version  = "0.0.1"
	Revision = "xxx"

	Name  = "mcp-notion"
	Usage = "A simple MCP server implementation"
)

func main() {
	app := cli.NewApp()
	app.Version = fmt.Sprintf("%s (%s)", Version, Revision)
	app.Name = Name
	app.Usage = Usage

	app.Commands = []*cli.Command{
		{
			Name:    "server",
			Aliases: []string{"s"},
			Usage:   "A simple MCP server implementation",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "config",
					Aliases: []string{"c"},
					Value:   "config.yml",
					Usage:   "path to the configuration file",
				},
				&cli.BoolFlag{
					Name:  "no-logs",
					Usage: "suppress all logs except fatal errors",
				},
				&cli.StringFlag{
					Name:  "log",
					Usage: "path to log file (if specified, logs will be written to this file)",
				},
			},
			Action: func(c *cli.Context) error {
				configPath := c.String("config")
				noLogs := c.Bool("no-logs")
				logPath := c.String("log")

				// Read the configuration file
				cfg, err := config.LoadConfig(configPath)
				if err != nil {
					return errors.Wrap(err, "failed to load configuration file")
				}

				// Initialize logger
				if err := logger.InitLogger(true, noLogs, logPath); err != nil {
					return errors.Wrap(err, "failed to initialize logger")
				}
				defer logger.Sync()

				// Start the server
				return server.Run(cfg)
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
