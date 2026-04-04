package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"github.com/retlehs/ansimotd/internal/api"
	"github.com/retlehs/ansimotd/internal/config"
)

var DownloadCommand = &cli.Command{
	Name:      "download",
	Usage:     "Download art packs from 16colo.rs",
	ArgsUsage: "<year>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "group",
			Aliases: []string{"g"},
			Usage:   "Filter packs by group name",
		},
		&cli.StringFlag{
			Name:    "pack",
			Aliases: []string{"p"},
			Usage:   "Download a specific pack by name",
		},
	},
	Action: runDownload,
}

func runDownload(_ context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		return fmt.Errorf("year argument is required (e.g. ansimotd download 1996)")
	}

	year, err := strconv.Atoi(cmd.Args().First())
	if err != nil {
		return fmt.Errorf("invalid year: %s", cmd.Args().First())
	}

	group := cmd.String("group")
	pack := cmd.String("pack")

	if err := config.EnsureDirs(); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Fetching pack list for %d...\n", year)

	packs, err := api.FetchPacks(year, group, pack)
	if err != nil {
		return err
	}

	if len(packs) == 0 {
		fmt.Fprintln(os.Stderr, "No packs found matching your criteria.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Found %d pack(s)\n", len(packs))

	downloaded := 0
	skipped := 0
	failed := 0
	var failures []string

	for i, p := range packs {
		fmt.Fprintf(os.Stderr, "[%d/%d] %s... ", i+1, len(packs), p.Name)

		ok, err := api.DownloadAndExtract(p, config.ArtDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			failures = append(failures, fmt.Sprintf("%s: %s", p.Name, err))
			failed++
			continue
		}
		if ok {
			fmt.Fprintln(os.Stderr, "done")
			downloaded++
		} else {
			fmt.Fprintln(os.Stderr, "skipped (exists)")
			skipped++
		}
	}

	fmt.Fprintf(os.Stderr, "\nComplete: %d downloaded, %d skipped, %d failed\n", downloaded, skipped, failed)
	fmt.Fprintf(os.Stderr, "Art saved to: %s\n", config.ArtDir())

	if len(failures) > 0 {
		fmt.Fprintln(os.Stderr, "\nFailed packs:")
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  - %s\n", f)
		}
	}

	return nil
}
