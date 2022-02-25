//go:build linux

package secrets

import (
	"errors"
	"os"
	"strings"
)

func NewManager() (Manager, error) {
	desktop := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))

	if strings.Contains(desktop, "kde") {
		return NewKWalletManager()
	} else if strings.Contains(desktop, "gnome") {
		return NewSecretServiceManager()
	}
	return nil, errors.New("unsupported platform")
}
