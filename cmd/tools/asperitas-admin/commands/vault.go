package commands

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/rocketb/asperitas/pkg/vault"
)

func Vault(conf vault.Config, keysFolder string) error {
	vaultSrv, err := vault.New(vault.Config{
		Address:   conf.Address,
		MountPath: conf.MountPath,
		Token:     conf.Token,
	})
	if err != nil {
		return fmt.Errorf("constructing vault client: %w", err)
	}

	return loadKeys(vaultSrv, os.DirFS(keysFolder))
}

func loadKeys(vault *vault.Vault, fsys fs.FS) error {
	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("WalkDir failure: %w", err)
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

		privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading auth private key: %w", err)
		}

		kid := strings.TrimSuffix(dirEntry.Name(), ".pem")
		fmt.Println("Loading kid:", kid)

		if err := vault.AddPrivateKey(context.Background(), kid, privatePEM); err != nil {
			return fmt.Errorf("put: %w", err)
		}

		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}
