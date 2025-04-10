package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/nspcc-dev/neo-go/pkg/wallet"
	cid "github.com/nspcc-dev/neofs-sdk-go/container/id"
	"github.com/nspcc-dev/neofs-sdk-go/user"
)

var (
	privKeyWif string
	cID        string
	network    string

	testnetStorageNodes = []string{"grpcs://st1.t5.fs.neo.org:8082", "grpcs://st2.t5.fs.neo.org:8082", "grpcs://st3.t5.fs.neo.org:8082", "grpcs://st4.t5.fs.neo.org:8082"}
	mainnetStorageNodes = []string{"grpcs://st1.storage.fs.neo.org:8082", "grpcs://st2.storage.fs.neo.org:8082", "grpcs://st3.storage.fs.neo.org:8082", "grpcs://st4.storage.fs.neo.org:8082"}
)

func setEnvVars() {
	privKeyWif = os.Getenv("PRIVATE_KEY_WIF")
	cID = os.Getenv("CONTAINER_ID")
	network = os.Getenv("NETWORK")
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func storageNodes() []string {
	if network == "testnet" {
		return testnetStorageNodes
	} else {
		return mainnetStorageNodes
	}
}

func containerId() cid.ID {
	var c cid.ID

	err := c.DecodeString((cID))
	panicOnErr(err)

	return c
}

func signer() user.Signer {
	acc, err := wallet.NewAccountFromWIF(privKeyWif)
	panicOnErr(err)

	return user.NewAutoIDSignerRFC6979(acc.PrivateKey().PrivateKey)
}

func findFilesRecursively(root string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and node_modules
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if !info.IsDir() && !strings.HasPrefix(d.Name(), ".") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
