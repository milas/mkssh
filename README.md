# mkssh
`mkssh` is a CLI tool to generate and configure SSH keys.
It's designed to enable easily using per-host (or per-service) keys without compromising usability or security.

* [x] Generate modern/secure keys (`ed25519`)
* [x] Encrypt private key with randomly generated passphrase
* [x] Store passphrase in system credential store
  * [x] KDE ([KWallet + `ksshaskpass`][kde-kwallet-ssh])
  * [ ] GNOME ([GNOME Keyring][gnome-keyring-ssh])
  * [x] macOS ([Keychain][macos-keychain-ssh])
  * [ ] Windows 10+ ([Credential Manager][windows-credential-manager-ssh])
* [x] Create host-specific entry in `~/.ssh/config`
* [ ] Add public key to GitHub, GitLab, etc.
* [x] Compatibility mode for legacy/nonstandard systems

[gnome-keyring-ssh]: https://wiki.gnome.org/Projects/GnomeKeyring/Ssh
[kde-kwallet-ssh]: https://invent.kde.org/plasma/ksshaskpass
[macos-keychain-ssh]: https://developer.apple.com/library/archive/technotes/tn2449/_index.html
[windows-credential-manager-ssh]: https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_keymanagement
