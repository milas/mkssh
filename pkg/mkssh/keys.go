package mkssh

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
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
}

func NewKeyPair(keyType KeyType) (KeyPair, error) {
	switch keyType {
	case KeyTypeEd25519:
		return NewEd25519KeyPair()
	case KeyTypeRSA:
		return NewRSAKeyPair()
	default:
		return KeyPair{}, fmt.Errorf("unsupported key type")
	}
}

func NewEd25519KeyPair() (KeyPair, error) {
	ret := KeyPair{Type: KeyTypeEd25519}
	var err error
	ret.Public, ret.Private, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, err
	}
	return ret, nil
}

func NewRSAKeyPair() (KeyPair, error) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, rsaKeySize)
	if err != nil {
		return KeyPair{}, err
	}

	return KeyPair{
		Type:    KeyTypeRSA,
		Public:  rsaKey.PublicKey,
		Private: rsaKey,
	}, nil
}

type SaveOptions struct {
	Comment string
}

func (k KeyPair) Save(basePath string, name string, opts SaveOptions) (err error) {
	privKeyFile, err := OpenTruncate(filepath.Join(basePath, name))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := privKeyFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	pubKeyFile, err := OpenTruncate(filepath.Join(basePath, name+".pub"))
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := pubKeyFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	if err := k.writePrivateKeyPEM(privKeyFile); err != nil {
		return err
	}

	if err := k.writePublicKeyOpenSSH(pubKeyFile, opts.Comment); err != nil {
		return err
	}

	return nil
}

func (k KeyPair) writePrivateKeyPEM(w io.Writer) error {
	var block pem.Block
	switch k.Type {
	case KeyTypeRSA:
		// if RSA is being used, it's likely for compatibility, so let's be
		// a bit more conservative and go with PKCS #1
		block = pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(k.Private.(*rsa.PrivateKey)),
		}
	default:
		enc, err := x509.MarshalPKCS8PrivateKey(k.Private)
		if err != nil {
			return err
		}
		block = pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: enc,
		}
	}

	if err := pem.Encode(w, &block); err != nil {
		return err
	}
	return nil
}

// writePublicKeyOpenSSH marshals the public key to the OpenSSH format.
//
// Adapted from https://cs.opensource.google/go/x/crypto/+/master:ssh/keys.go;l=279-290;drc=b4de73f9ece8163b492578e101e4ef8923ac2c5c.
// (Adds comment support and uses io.Writer directly.)
func (k KeyPair) writePublicKeyOpenSSH(w io.Writer, comment string) error {
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

	if comment != "" {
		if _, err := w.Write([]byte(" " + strings.TrimSpace(comment))); err != nil {
			return err
		}
	}

	if _, err := w.Write([]byte("\n")); err != nil {
		return err
	}
	return nil
}
