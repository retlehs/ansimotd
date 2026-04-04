package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/retlehs/ansimotd/cmd"
)

var version = "dev"

func main() {
	app := &cli.Command{
		Name:    "ansimotd",
		Usage:   "Display ANSI art as your message of the day",
		Version: version,
		Commands: []*cli.Command{
			cmd.DisplayCommand,
			cmd.DownloadCommand,
			cmd.LastCommand,
		},
		// Running with no subcommand defaults to display
		Action: func(ctx context.Context, c *cli.Command) error {
			return cmd.DisplayCommand.Action(ctx, c)
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "ansimotd: %s\n", err)
		os.Exit(1)
	}
}
