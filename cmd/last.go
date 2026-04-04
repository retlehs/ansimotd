package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/retlehs/ansimotd/internal/config"
)

var LastCommand = &cli.Command{
	Name:  "last",
	Usage: "Print the path of the last displayed ANSI art file",
	Action: func(_ context.Context, _ *cli.Command) error {
		data, err := os.ReadFile(config.LastFile())
		if err != nil {
			return fmt.Errorf("no last file recorded (have you run ansimotd display?)")
		}
		fmt.Println(strings.TrimSpace(string(data)))
		return nil
	},
}
