package secrets

type Manager interface {
	SaveSSHKeyfilePassword(path string, passphrase string) error

	Close() error
}
