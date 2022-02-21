package cli

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func NewApp() (*cli.App, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	defaultSSHDir := filepath.Join(homeDir, ".ssh", "config")
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:      "config",
				Aliases:   []string{"c"},
				Value:     filepath.Join(defaultSSHDir, "config"),
				Usage:     "Path to the SSH config file",
				TakesFile: true,
			},
		},
		Commands: []*cli.Command{
			NewAddCommand(),
		},
	}
	return app, nil
}
