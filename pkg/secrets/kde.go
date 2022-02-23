//go:build linux

package secrets

import (
	"fmt"

	"r00t2.io/gokwallet"
)

type KWalletManager struct {
	wm *gokwallet.WalletManager
}

func NewKWalletManager() (KWalletManager, error) {
	wm, err := gokwallet.NewWalletManager(&gokwallet.RecurseOpts{}, "example")
	if err != nil {
		return KWalletManager{}, err
	}
	return KWalletManager{
		wm: wm,
	}, nil
}

func (k KWalletManager) SaveSSHKeyfilePassword(path string, passphrase string) error {
	w, err := gokwallet.NewWallet(k.wm, gokwallet.DefaultWalletName, &gokwallet.RecurseOpts{})
	if err != nil {
		return fmt.Errorf("could not open kwallet: %v", err)
	}
	defer w.Close()

	const folderName = "ksshaskpass"
	if hasFolder, err := w.HasFolder(folderName); err != nil {
		return fmt.Errorf("could not open kwallet: %v", err)
	} else if !hasFolder {
		if err := w.CreateFolder(folderName); err != nil {
			return fmt.Errorf("could not create %q kwallet folder: %v", folderName, err)
		}
	}

	f, err := gokwallet.NewFolder(w, folderName, &gokwallet.RecurseOpts{Passwords: true})
	if err != nil {
		return fmt.Errorf("could not open %q kwallet folder: %v", folderName, err)
	}

	if _, err := f.WritePassword(path, passphrase); err != nil {
		return fmt.Errorf("could not update kwallet: %v", err)
	}
	return nil
}

func (k KWalletManager) Close() error {
	return k.wm.Close()
}
