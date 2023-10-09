package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var DefaultBlocks = Blocks{
	[]string{"!std"},
	[]string{"*"},
	[]string{"!mod"},
}

type Config struct {
	Blocks Blocks `yaml:"importblocks"`
}

type Blocks [][]string

func ReadConfig(path string, verbose bool) (Config, error) {
	var (
		c   Config
		err error
	)

	bb, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	err = yaml.Unmarshal(bb, &c)
	if err != nil {
		return Config{}, err
	}

	if verbose {
		fmt.Printf("config file found: %s\n", path)
	}

	return c, nil
}

func Lookup(dir, fileName string) Config {
	var (
		bb      = Blocks{}
		prevDir = ""
	)

	for {
		if prevDir == dir {
			break
		}
		lookupFile := filepath.Join(dir, fileName)
		dirBlocks := blocksFromFile(lookupFile)
		if len(dirBlocks) != 0 {
			bb = append(dirBlocks, bb...)
			fmt.Printf("lookup: %s\n", lookupFile)
		}
		prevDir = dir
		dir = filepath.Dir(dir)
	}

	if len(bb) == 0 {
		fmt.Println("empty config after lookup. default config will be used")
		bb = DefaultBlocks
	}

	return Config{Blocks: bb}
}

func blocksFromFile(path string) Blocks {
	var (
		c   Config
		err error
	)

	bb, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		fmt.Printf("skip config. errors reading: %s\n", path)
		return nil
	}

	err = yaml.Unmarshal(bb, &c)
	if err != nil {
		fmt.Printf("skip config. errors unmarshal: %s\n", path)
		return nil
	}

	return c.Blocks
}
