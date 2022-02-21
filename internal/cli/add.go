package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/milas/mkssh/pkg/mkssh"
)

func NewAddCommand() *cli.Command {
	return &cli.Command{
		Name: "add",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Required: true,
				Usage:    "Name for the key-pair",
			},
			&cli.StringFlag{
				Name:  "comment",
				Usage: "Comment (e.g. email) to include in public key",
			},
			&cli.StringFlag{
				Name:  "user",
				Usage: "Default username for connections",
				Value: "git",
			},
			&cli.StringFlag{
				Name:  "key-type",
				Value: "ed25519",
				Usage: "Key algorithm, supported values: ed25519, rsa",
			},
			&cli.PathFlag{
				Name:      "key-directory",
				Aliases:   []string{"key-dir"},
				Usage:     "Directory to store generated keys in",
				TakesFile: true,
			},
		},
		Action: func(c *cli.Context) error {
			name := c.String("name")
			if name == "" {
				return errors.New("name is required")
			}

			keyDir := c.Path("key-directory")
			if keyDir == "" {
				keyDir = filepath.Dir(c.Path("config"))
			}
			if absDir, err := filepath.Abs(keyDir); err == nil {
				keyDir = absDir
			}

			if stat, err := os.Stat(keyDir); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("key directory does not exist: %s", keyDir)
				} else if os.IsPermission(err) {
					return fmt.Errorf("cannot access key directory: %s", keyDir)
				} else {
					return fmt.Errorf("unexpected error: %v", err)
				}
			} else if !stat.IsDir() {
				return fmt.Errorf("key directory path is not a directory: %s", keyDir)
			}

			var keyType mkssh.KeyType
			switch c.String("key-type") {
			case "ed25519":
				keyType = mkssh.KeyTypeEd25519
			case "rsa":
				keyType = mkssh.KeyTypeRSA
			}

			k, err := mkssh.NewKeyPair(keyType)
			if err != nil {
				return err
			}

			opts := mkssh.SaveOptions{
				Comment: c.String("comment"),
			}

			if err := k.Save(keyDir, name, opts); err != nil {
				return err
			}

			return nil
		},
	}
}
