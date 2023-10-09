package internal

import (
	"strings"

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
