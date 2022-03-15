//go:build windows

package secrets

type WinCredManager struct{}

func (w WinCredManager) SavePrivateKeyPassphrase(path string, passphrase string) error {
	return nil
}

func (w WinCredManager) Close() error {
	return nil
}

func NewManager() (Manager, error) {
	return WinCredManager{}, nil
}
