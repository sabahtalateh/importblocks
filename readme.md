# Import Blocks

Order imports into blocks defined in config file

## Installation

```shell
go install github.com/sabahtalateh/importblocks@latest
```

You may also need to modify your `~/.[bash|zsh|fish..]rc` with 

```shell
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

So modifications to take effect run
```shell
source ~/.[bash|zsh|fish..]rc
```

## Syntax
```
importblocks [-config ..] [-lookup ..] [path ..]
```

- `-config` **relative** or **absolute** path to config file. (see: [Config File](#config-file))
- `-lookup` lookup config file name. (see: [Lookup Config File](#lookup-config-file))
- `path` file or directory **relative** or **absolute** path. If ends with `/...` then inner files and dirs will be formatted recursively. Ex: `code/go/project1/...`, `./...`. Multiple paths may be specified

## Quickstart

Suppose you have

```go
import (
	"a/b/c"
	"fmt"
	"x/y/z"
	"os"
	"github.com/my/module/pkg2" // your module package
	"math"
	"github.com/my/module/pkg1" // your module package
	"github.com/thirdparty/pkg1"
	"golang.org/x/exp"
	"a/b/c2"
)
```

To order imports in your project root create `importblocks.yaml` file with content

```yaml
importblocks:
  - [ "!std" ]                                    # standard library
  - [ "github.com/thirdparty", "golang.org/x" ]   # some package prefixes
  - [ "*" ]                                       # everything else
  - [ "!mod" ]                                    # your module packages
```

Then run program from your module dir
```shell
importblocks -config imports.yaml ./...
```

Imports will be grouped by blocks. Imports within block sorted alphabetically

```go
import (
	"fmt"
	"os"
	"math"
	
	"github.com/thirdparty/pkg1"
	"golang.org/x/exp"

	"a/b/c"
	"a/b/c2"
	"x/y/z"

	"github.com/my/module/pkg1"
	"github.com/my/module/pkg2"
)
```

## Config file

Config file is `yaml` with 2-dimensional array of specifiers where specifier can be one of
- `!std` to mention any standard library package
- `!mod` to mention your module package
- Any string. Package prefix to mention package path starting with it
- `*` to mention any package path which not specified by other specifiers

## Lookup Config File

**Multiple config files may exist on different levels of directories hierarchy.** With this feature you may have common config for different projects inside one directory and specific config for each project. **Nested config content will be appended to topmost**

Example

```shell
root
├──project1
│  └──imports.yaml
├──project2
│  └──imports.yaml
└──imports.yaml
```


```yaml
# root/imports.yaml
importblocks:
  - [ "!std" ]
  - [ "*" ]
```

```yaml
# root/project1/imports.yaml
importblocks:
  - [ "monorepo/project1" ]
```

Then
```shell
$ cd root/project1
$ importblocks -lookup imports.yaml ./...
```

`.go`-files will be formatted with computed config `root/imports.yaml` + `root/project1/imports.yaml`
```yaml
importblocks:
  - [ "!std" ]
  - [ "*" ]
  - [ "monorepo/project1" ]
```

