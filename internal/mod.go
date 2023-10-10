package internal

import (
	"os"

	"github.com/sabahtalateh/mod"
	"golang.org/x/mod/modfile"
)

func modPath(dir string) (string, error) {
	gmp, err := mod.ModFilePath(dir)
	if err != nil {
		return "", err
	}

	bb, err := os.ReadFile(gmp)
	if err != nil {
		return "", err
	}

	modFile, err := modfile.Parse("go.mod", bb, nil)
	if err != nil {
		return "", err
	}

	return modFile.Module.Mod.Path, nil
}
