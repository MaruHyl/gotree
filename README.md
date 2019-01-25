# gopkg
> Tools for go pkgs's inspection and analysis.
> Base on golang.org/x/tools/go/packages

## Installation

```sh
go get github.com/MaruHyl/gopkg
```

## Usage example

### Show current dir pkg

```sh
gopkg
```

### Find pkg

```sh
gopkg -a "strings"
```

### Show current pkg deps

```sh
gopkg deps
```

### Show pkg deps

```sh
gopkg deps -a "strings"
```

### Show pkg deps like tree

```sh
gopkg deps tree -a "strings"
```

### Show pkg deps like tree with pattern

```sh
gopkg deps tree -a "strings" -p "unsafe"
```

### More usage

Please refer to the help detail.

```sh
gopkg -h
gopkg deps -h
gopkg deps tree -h
```
