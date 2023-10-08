package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Formatter struct {
	o  *Ordering
	wd string
}

func NewFormatter(o *Ordering, wd string) *Formatter {
	return &Formatter{o: o, wd: wd}
}

func (f *Formatter) Walk(path string) {
	_ = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			f.File(path)
		}

		return nil
	})
}

func (f *Formatter) Path(path string) {
	stat, err := os.Stat(path)
	if err != nil {
		return
	}
	if stat.IsDir() {
		f.Dir(path)
	} else {
		f.File(path)
	}
}

func (f *Formatter) Dir(dir string) {
	dd, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range dd {
		if !entry.IsDir() {
			f.File(filepath.Join(dir, entry.Name()))
		}
	}
}

func (f *Formatter) File(path string) {
	if !strings.HasSuffix(path, ".go") {
		return
	}

	fileShort := strings.TrimPrefix(path, f.wd+string(os.PathSeparator))
	f.o.OrderImports(path, fileShort)
}
