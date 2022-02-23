//go:build darwin

package secrets

import (
	"github.com/keybase/go-keychain"
)

type KeychainManager struct{}

func (k KeychainManager) SaveSSHKeyfilePassword(path string, passphrase string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("OpenSSH")
	item.SetAccount(path)
	item.SetLabel("SSH: " + path)
	item.SetAccessGroup("com.apple.ssh.passphrases")
	item.SetData([]byte(passphrase))
	item.SetSynchronizable(keychain.SynchronizableNo)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	return keychain.AddItem(item)
}

func (k KeychainManager) Close() error {
	return nil
}

func NewManager() (Manager, error) {
	return KeychainManager{}, nil
}
