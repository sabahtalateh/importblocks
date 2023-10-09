package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sabahtalateh/importblocks/internal"
)

var configFile string
var lookup string
var verbose bool

const lookupUsage = `file name to lookup. NOT path to file
upwards lookup will be performed starting from current work dir
config values will be appended from topmost to closest to current work dir
if no config file found then default config will be used:
importblocks:
  - ["!std"]
  - ["*"]
  - ["!mod"]

example:

when 
workdir = /a/b/c and config_file_name = importblocks.yaml

then 
/a/b/c/importblocks.yaml, /a/b/importblocks.yaml, /a/importblocks.yaml, /importblocks.yaml
paths will be attempted to parse looking for config

if
/a/importblocks.yaml exists and contains
importblocks:
  - ["*"]

and /a/b/c/importblocks.yaml exists and contains
importblocks:
  - ["github.com/a/b/c"]

then resulting config will be
importblocks:
  - ["*"]
  - ["github.com/a/b/c"]
`

func init() {
	flag.StringVar(&configFile, "config", "", "config file path")
	flag.StringVar(&configFile, "c", "", "shorthand for config")
	flag.StringVar(&lookup, "lookup", "", lookupUsage)
	flag.StringVar(&lookup, "l", "", "shorthand for lookup")
	flag.BoolVar(&verbose, "v", false, "verbose mode")
}

func isHelp(arg string) bool {
	if arg == "h" {
		return true
	}

	if arg == "-h" {
		return true
	}

	if arg == "--h" {
		return true
	}

	if arg == "help" {
		return true
	}

	if arg == "-help" {
		return true
	}

	if arg == "--help" {
		return true
	}

	return false
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	if isHelp(args[0]) {
		flag.Usage()
		os.Exit(0)
	}

	wd, err := os.Getwd()
	check(err)

	ordering, err := internal.NewOrdering(conf(wd, verbose), wd)
	check(err)

	formatter := internal.NewFormatter(ordering, wd, verbose)

	for _, arg := range args {
		if strings.HasSuffix(arg, "...") {
			path := strings.TrimSuffix(arg, "...")
			if !filepath.IsAbs(path) {
				path = filepath.Join(wd, path)
			}
			formatter.Walk(path)
		} else {
			path := arg
			if !filepath.IsAbs(path) {
				path = filepath.Join(wd, path)
			}
			formatter.Path(path)
		}
	}
}

func conf(wd string, verbose bool) internal.Config {
	if configFile == "" && lookup == "" {
		if verbose {
			fmt.Println("nor -config or -lookup passed. default config will be used")
		}
		return internal.Config{Blocks: internal.DefaultBlocks}
	}

	var (
		config internal.Config
		err    error
	)

	if configFile != "" {
		config, err = internal.ReadConfig(configFile, verbose)
		check(err)
	}

	if lookup != "" {
		config = internal.Lookup(wd, lookup)
	}

	return config
}

func check(err error) {
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
