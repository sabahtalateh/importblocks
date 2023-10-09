package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Formatter struct {
	o       *Ordering
	wd      string
	verbose bool
}

func NewFormatter(o *Ordering, wd string, verbose bool) *Formatter {
	return &Formatter{o: o, wd: wd, verbose: verbose}
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

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("%s: skip. error parsing\n", fileShort)
		return
	}

	buckets := make([][]*ast.ImportSpec, f.o.Buckets())

	var startImportsLine int
	var endImportsLine int

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			if startImportsLine == 0 {
				startImportsLine = fset.Position(decl.Pos()).Line
			}

			endImportsLine = fset.Position(decl.End()).Line

			impVal, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return
			}

			bucket := f.o.Bucket(impVal)
			buckets[bucket] = append(buckets[bucket], importSpec)
		}
	}

	var importsLines []string

	for _, bucket := range buckets {
		if len(bucket) == 0 {
			continue
		}

		sort.Slice(bucket, func(i, j int) bool {
			iVal, err := strconv.Unquote(bucket[i].Path.Value)
			if err != nil {
				return false
			}

			jVal, err := strconv.Unquote(bucket[j].Path.Value)
			if err != nil {
				return false
			}

			return iVal < jVal
		})

		for _, spec := range bucket {
			bb := new(bytes.Buffer)
			err = printer.Fprint(bb, fset, spec)
			if err != nil {
				return
			}

			ll := strings.Split(bb.String(), "\n")
			for i := 0; i < len(ll); i++ {
				ll[i] = "\t" + ll[i]
			}

			importsLines = append(importsLines, ll...)
		}
		importsLines = append(importsLines, "")
	}

	// trim leading empty lines
	for i := 0; i < len(importsLines); i++ {
		if importsLines[i] != "" {
			break
		}
		importsLines = importsLines[i:]
	}

	// trim trailing empty lines
	for i := len(importsLines) - 1; i > -1; i-- {
		if importsLines[i] != "" {
			break
		}
		importsLines = importsLines[:i]
	}

	if len(importsLines) == 0 {
		return
	}

	bb, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%s: skip. error reading\n", fileShort)
		return
	}

	lines := strings.Split(string(bb), "\n")
	origLines := slices.Clone(lines)

	head := lines[:startImportsLine-1]
	var tail []string
	for i := endImportsLine; i < len(lines); i++ {
		tail = append(tail, lines[i])
	}
	lines = append(head, "import (")
	lines = append(lines, importsLines...)
	lines = append(lines, ")")
	lines = append(lines, tail...)

	if !linesEquals(origLines, lines) {
		err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), os.ModePerm)
		if err != nil {
			fmt.Printf("%s: skip. error writing\n", fileShort)
			return
		}
		fmt.Println(fileShort)
		return
	}

	if f.verbose {
		fmt.Printf("%s: no changes\n", fileShort)
	}
}

func linesEquals(ll1, ll2 []string) bool {
	if len(ll1) != len(ll2) {
		return false
	}

	for i := 0; i < len(ll1); i++ {
		if ll1[i] != ll2[i] {
			return false
		}
	}

	return true
}
