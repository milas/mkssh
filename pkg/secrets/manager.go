package secrets

import (
	"errors"
)

var ErrNotSupported = errors.New("not supported")

type Manager interface {
	SavePrivateKeyPassphrase(path string, passphrase string) error

	Close() error
}
