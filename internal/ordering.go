package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/packages"
)

type Ordering struct {
	stdOrder int
	anyOrder int
	order    map[string]int

	std map[string]struct{}
}

func NewOrdering(c Config, wd string) (*Ordering, error) {
	o := &Ordering{
		order: map[string]int{},
		std:   map[string]struct{}{},
	}

	pp, err := packages.Load(nil, "std")
	if err != nil {
		return nil, err
	}

	for _, p := range pp {
		o.std[p.ID] = struct{}{}
	}

	stdFound := false
	anyFound := false

	for i, vv := range c.Blocks {
		for _, v := range vv {
			if v == "!std" {
				stdFound = true
				o.stdOrder = i
			}

			if v == "!mod" {
				v = modPath(wd)
			}

			if v == "*" {
				o.anyOrder = i
				anyFound = true
			}

			o.order[v] = i
		}
	}

	if !anyFound {
		o.anyOrder = len(c.Blocks)
	}

	if !stdFound {
		for k := range o.order {
			o.order[k] += 1
		}
		o.anyOrder += 1
	}

	return o, nil
}

func (o *Ordering) Bucket(imp string) int {
	for pkgPrefix := range o.std {
		if strings.HasPrefix(imp, pkgPrefix) {
			return o.stdOrder
		}
	}

	orderCandidates := map[string]int{}
	for pkgPrefix, ord := range o.order {
		if strings.HasPrefix(imp, pkgPrefix) {
			orderCandidates[pkgPrefix] = ord
		}
	}

	if len(orderCandidates) == 0 {
		return o.anyOrder
	}

	var longest string
	for pkg := range orderCandidates {
		if len(pkg) > len(longest) {
			longest = pkg
		}
	}

	return orderCandidates[longest]
}

func (o *Ordering) Buckets() int {
	max := 0

	if o.stdOrder > max {
		max = o.stdOrder
	}

	if o.anyOrder > max {
		max = o.anyOrder
	}

	for _, ord := range o.order {
		if ord > max {
			max = ord
		}
	}

	return max + 1
}

func (o *Ordering) OrderImports(path, fileShort string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("skip. error parsing: %s\n", fileShort)
		return
	}

	buckets := make([][]*ast.ImportSpec, o.Buckets())

	var startImportsLine int
	var endImportsLine int

	for _, decl := range f.Decls {
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

			bucket := o.Bucket(impVal)
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
		fmt.Printf("skip. error reading: %s\n", fileShort)
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
			fmt.Printf("skip. error writing: %s\n", fileShort)
			return
		}
		fmt.Println(fileShort)
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
