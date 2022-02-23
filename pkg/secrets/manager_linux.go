//go:build linux

package secrets

import (
	"errors"
	"os"
	"strings"
)

func NewManager() (Manager, error) {
	switch strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP")) {
	case "kde":
		return NewKWalletManager()
	}
	return nil, errors.New("unsupported platform")
}
