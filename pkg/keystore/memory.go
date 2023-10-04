// This implements an in memory keystore for JWT support.

package keystore

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// Error variables for key store.
var (
	ErrKidLookup = errors.New("kid lookup failed")
)

// PrivateKey represents key information.
type PrivateKey struct {
	PK  *rsa.PrivateKey
	PEM []byte
}

// Memory represents an in memory store implementation of the
// KeyLookup interface for use with the auth package.
type Memory struct {
	store map[string]PrivateKey
}

// NewMemory constructs an empty key store.
func NewMemory() *Memory {
	return &Memory{
		store: make(map[string]PrivateKey),
	}
}

// NewMemoryStore constructs a key store with an initial set of keys.
func NewMemoryStore(store map[string]PrivateKey) *Memory {
	return &Memory{
		store: store,
	}
}

// NewMemoryFS constructs a key store and fill it with keys from fs.
func NewMemoryFS(fsys fs.FS) (*Memory, error) {
	ks := NewMemory()

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if dirEntry.IsDir() {
			return nil
		}

		if path.Ext(fileName) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)
		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}
		defer file.Close()

		pem, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("feeding auth private key: %w", err)
		}

		pk, err := jwt.ParseRSAPrivateKeyFromPEM(pem)
		if err != nil {
			return fmt.Errorf("parsing auth private key: %w", err)
		}

		key := PrivateKey{
			PK:  pk,
			PEM: pem,
		}

		ks.store[strings.TrimSuffix(dirEntry.Name(), ".pem")] = key

		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return ks, nil
}

// PrivateKeyPEM searches the key store for a given kid and returns
// the private key.
func (ks *Memory) PrivateKeyPEM(kid string) (string, error) {
	pk, ok := ks.store[kid]
	if !ok {
		return "", ErrKidLookup
	}

	return string(pk.PEM), nil
}

// PublicKeyPEM searches the key store for a given kid and returns
// the public key.
func (ks *Memory) PublicKeyPEM(kid string) (string, error) {
	pk, ok := ks.store[kid]
	if !ok {
		return "", ErrKidLookup
	}

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&pk.PK.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshalling public key: %w", err)
	}

	block := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	var b bytes.Buffer
	if err := pem.Encode(&b, &block); err != nil {
		return "", fmt.Errorf("encoding to private file: %w", err)
	}

	return b.String(), nil
}
