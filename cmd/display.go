package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/retlehs/ansimotd/internal/config"
	"github.com/retlehs/ansimotd/internal/picker"
	"github.com/retlehs/ansimotd/internal/render"
	"github.com/retlehs/ansimotd/internal/sauce"
)

var DisplayCommand = &cli.Command{
	Name:  "display",
	Usage: "Display random or specific ANSI art",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "Display a specific ANSI file instead of random",
		},
	},
	Action: runDisplay,
}

func runDisplay(_ context.Context, cmd *cli.Command) error {
	if err := config.EnsureDirs(); err != nil {
		return err
	}

	file := cmd.String("file")
	if file == "" {
		var err error
		file, err = picker.Pick(config.ArtDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "ansimotd: %s\n", err)
			fmt.Fprintf(os.Stderr, "Art directory: %s\n", config.ArtDir())
			return nil
		}
	}

	raw, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file, err)
	}

	rec := sauce.ParseBytes(raw)

	if err := render.Render(os.Stdout, raw, rec); err != nil {
		return fmt.Errorf("rendering %s: %w", file, err)
	}

	// Write state file (atomic)
	writeLastFile(file)

	return nil
}

func writeLastFile(path string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	tmp := config.LastFile() + ".tmp"
	if err := os.WriteFile(tmp, []byte(abs+"\n"), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, config.LastFile())
}
