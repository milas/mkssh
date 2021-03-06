package mkssh

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/youmark/pkcs8"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const rsaKeySize = 4096

type KeyType int

const (
	KeyTypeEd25519 KeyType = iota + 1
	KeyTypeRSA
)

type KeyPair struct {
	Public  crypto.PublicKey
	Private crypto.PrivateKey
	Type    KeyType
	Comment string
}

func NewKeyPair(keyType KeyType, comment string) (KeyPair, error) {
	switch keyType {
	case KeyTypeEd25519:
		return NewEd25519KeyPair(comment)
	case KeyTypeRSA:
		return NewRSAKeyPair(comment)
	default:
		return KeyPair{}, fmt.Errorf("unsupported key type")
	}
}

func NewEd25519KeyPair(comment string) (KeyPair, error) {
	ret := KeyPair{Type: KeyTypeEd25519, Comment: comment}
	var err error
	ret.Public, ret.Private, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}
	return ret, nil
}

func NewRSAKeyPair(comment string) (KeyPair, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		Type:    KeyTypeRSA,
		Comment: comment,
		// N.B. x/crypto/ssh helpers expect *rsa.PublicKey
		Public:  &rsaKey.PublicKey,
		Private: rsaKey,
	}, nil
}

func (k KeyPair) Save(path, passphrase string) (err error) {
	var privKeyFile io.WriteCloser

	if passphrase != "" {
		privKeyFile, err = OpenTruncate(path)
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := privKeyFile.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()
	}
	pubKeyFile, err := OpenTruncate(path + ".pub")
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := pubKeyFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if privKeyFile != nil {
		if err := k.WritePrivateKey(privKeyFile, passphrase); err != nil {
			return err
		}
	}

	if err := k.WritePublicKey(pubKeyFile); err != nil {
		return err
	}

	return nil
}

func (k KeyPair) MarshalPublicKey() ([]byte, error) {
	var buf bytes.Buffer
	if err := k.WritePublicKey(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (k KeyPair) MarshalPrivateKey(passphrase string) ([]byte, error) {
	var buf bytes.Buffer
	if err := k.WritePrivateKey(&buf, passphrase); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (k KeyPair) AddToAgent() error {
	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return fmt.Errorf("could not connect to SSH agent: %v", err)
	}

	cli := agent.NewClient(conn)

	key := agent.AddedKey{
		PrivateKey:       k.Private,
		Comment:          k.Comment,
		ConfirmBeforeUse: false,
	}
	if err := cli.Add(key); err != nil {
		return fmt.Errorf("could not add key to SSH agent: %v", err)
	}
	return nil
}

func (k KeyPair) WritePrivateKey(w io.Writer, passphrase string) error {
	if passphrase == "" {
		return errors.New("passphrase is required")
	}

	var block *pem.Block
	switch k.Type {
	case KeyTypeRSA:
		// if RSA is being used, it's likely for compatibility, so let's be
		// a bit more conservative and go with PKCS #1 + 3DES
		var err error
		//goland:noinspection GoDeprecation
		block, err = x509.EncryptPEMBlock(
			rand.Reader,
			"RSA PRIVATE KEY",
			x509.MarshalPKCS1PrivateKey(k.Private.(*rsa.PrivateKey)),
			[]byte(passphrase),
			x509.PEMCipher3DES,
		)
		if err != nil {
			return err
		}
	default:
		data, err := pkcs8.ConvertPrivateKeyToPKCS8(k.Private, []byte(passphrase))
		if err != nil {
			return err
		}
		block = &pem.Block{
			Type:  "ENCRYPTED PRIVATE KEY",
			Bytes: data,
		}
	}

	if err := pem.Encode(w, block); err != nil {
		return err
	}
	return nil
}

// WritePublicKey marshals the public key to the OpenSSH format.
//
// Adapted from https://cs.opensource.google/go/x/crypto/+/master:ssh/keys.go;l=279-290;drc=b4de73f9ece8163b492578e101e4ef8923ac2c5c.
// (Adds comment support and uses io.Writer directly.)
func (k KeyPair) WritePublicKey(w io.Writer) error {
	/*
		Copyright (c) 2009 The Go Authors. All rights reserved.

		Redistribution and use in source and binary forms, with or without
		modification, are permitted provided that the following conditions are
		met:

		   * Redistributions of source code must retain the above copyright
		notice, this list of conditions and the following disclaimer.
		   * Redistributions in binary form must reproduce the above
		copyright notice, this list of conditions and the following disclaimer
		in the documentation and/or other materials provided with the
		distribution.
		   * Neither the name of Google Inc. nor the names of its
		contributors may be used to endorse or promote products derived from
		this software without specific prior written permission.

		THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
		"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
		LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
		A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
		OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
		SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
		LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
		DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
		THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
		(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
		OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
	*/
	pubKey, err := ssh.NewPublicKey(k.Public)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte(pubKey.Type() + " ")); err != nil {
		return err
	}

	e := base64.NewEncoder(base64.StdEncoding, w)
	if _, err := e.Write(pubKey.Marshal()); err != nil {
		return err
	}
	if err := e.Close(); err != nil {
		return err
	}

	comment := strings.TrimSpace(k.Comment)
	if comment != "" {
		if _, err := w.Write([]byte(" " + comment)); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}
