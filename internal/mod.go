package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func modPath(dir string) string {
	if dir == "/" || dir == "." || dir == "" {
		fmt.Println("skip !module. module not found")
		return "*"
	}

	bb, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if os.IsNotExist(err) {
		return modPath(filepath.Dir(dir))
	}
	if err != nil {
		fmt.Printf("skip !module. error reading: %s\n", filepath.Join(dir, "go.mod"))
		return "*"
	}

	mod, err := modfile.Parse("go.mod", bb, nil)
	if err != nil {
		fmt.Printf("skip !module. error parsing: %s\n", filepath.Join(dir, "go.mod"))
		return "*"
	}

	fmt.Printf("module set to %s\n", mod.Module.Mod.Path)
	return mod.Module.Mod.Path
}
