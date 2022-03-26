package cli

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/sethvargo/go-password/password"
	"github.com/urfave/cli/v2"

	"github.com/milas/mkssh/pkg/mkssh"
	"github.com/milas/mkssh/pkg/secrets"
	"github.com/milas/mkssh/pkg/sshconfig"
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

			k, err := mkssh.NewKeyPair(keyType, getKeyComment(c))
			if err != nil {
				return err
			}

			passphrase, err := password.Generate(64, 10, 10, false, false)
			if err != nil {
				return err
			}

			keyPath := filepath.Join(keyDir, name)
			if err := k.Save(keyPath, passphrase); err != nil {
				return err
			}

			secretsManager, err := secrets.NewManager()
			if err != nil {
				return err
			}
			// extra func() so defer secrets manager close always happens here
			err = func() (err error) {
				defer func() {
					if closeErr := secretsManager.Close(); closeErr != nil && err == nil {
						err = closeErr
					}
				}()
				return secretsManager.SavePrivateKeyPassphrase(filepath.Join(keyDir, name), passphrase)
			}()
			if err != nil {
				return err
			}

			if err := k.AddToAgent(); err != nil {
				return err
			}

			err = sshconfig.UpdateHostInConfig(c.Path("config"), name, name, keyPath)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func getKeyComment(c *cli.Context) string {
	if c.IsSet("comment") {
		return c.String("comment")
	}

	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	u, err := user.Current()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s@%s (mkssh)", u.Username, hostname)
}
