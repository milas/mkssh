//go:build linux

package secrets

import (
	"fmt"

	"github.com/keybase/go-keychain/secretservice"
	dbus "github.com/keybase/go.dbus"
)

type SecretServiceManager struct {
	service *secretservice.SecretService
	conn    *dbus.Conn
	session *secretservice.Session
}

func NewSecretServiceManager() (SecretServiceManager, error) {
	// HACK(milas): go-keychain doesn't expose the DBus connection but the go
	// 	dbus lib stores the session bus as a global, so retrieve it here to
	// 	disconnect cleanly
	conn, err := dbus.SessionBus()
	if err != nil {
		return SecretServiceManager{}, err
	}

	svc, err := secretservice.NewService()
	if err != nil {
		return SecretServiceManager{}, err
	}

	session, err := svc.OpenSession(secretservice.AuthenticationDHAES)
	if err != nil {
		return SecretServiceManager{}, fmt.Errorf("could not open secret service session: %v", err)
	}

	return SecretServiceManager{
		service: svc,
		session: session,
		conn:    conn,
	}, nil
}

func (s SecretServiceManager) SavePrivateKeyPassphrase(path string, passphrase string) error {
	secret, err := s.session.NewSecret([]byte(passphrase))
	if err != nil {
		return fmt.Errorf("could not make secret: %v", err)
	}

	props := secretservice.NewSecretProperties(
		fmt.Sprintf("Unlock password for: %s", path),
		map[string]string{
			"unique": fmt.Sprintf("ssh-store:%s", path),
		},
	)
	props["org.freedesktop.Secret.Item.Locked"] = dbus.MakeVariant(false)

	_, err = s.service.CreateItem(
		secretservice.DefaultCollection,
		props,
		secret,
		secretservice.ReplaceBehaviorReplace,
	)
	if err != nil {
		return fmt.Errorf("could not save secret: %v", err)
	}
	return nil
}

func (s SecretServiceManager) Close() error {
	s.service.CloseSession(s.session)
	if err := s.conn.Close(); err != nil {
		return err
	}
	return nil
}
