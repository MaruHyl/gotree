package gotree

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/MaruHyl/gotree/internal/std"
	"github.com/fatih/color"
)

type Type string

const (
	Root     Type = "root"
	Direct   Type = "direct"
	Indirect Type = "indirect"
	Report   Type = "report"
)

type Dep interface {
	Name() string
	Deps() []Dep
}

type dep struct {
	Type    Type
	Name    string
	Matched bool
	Deps    []dep `json:",omitempty"`
}

type report struct {
	Type     Type
	Deps     int
	Direct   int
	Indirect int
}

// Get dep graph(json)
func JSONTree(d Dep, options ...Option) (string, error) {
	if d == nil {
		return "", nil
	}
	opts, err := buildOpts(options...)
	if err != nil {
		return "", err
	}
	dep, r := visit(d, opts)
	val := []interface{}{dep}
	if !opts.noReport {
		val = append(val, r)
	}
	b, err := json.MarshalIndent(val, "", " ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

const prefixClose = "└── "
const prefixOpen = "├── "
const close = "    "
const open = "│   "

// Get dep graph(tree)
func Tree(d Dep, options ...Option) (string, error) {
	if d == nil {
		return "", nil
	}
	opts, err := buildOpts(options...)
	if err != nil {
		return "", err
	}
	nd, r := visit(d, opts)
	//
	sb := new(strings.Builder)
	var dfs func(d dep, flags []bool)
	dfs = func(nd dep, flags []bool) {
		// build prefix
		prefix := ""
		for i, flag := range flags {
			isPrefix := i == len(flags)-1
			if flag {
				if isPrefix {
					prefix += prefixOpen
				} else {
					prefix += open
				}
			} else {
				if isPrefix {
					prefix += prefixClose
				} else {
					prefix += close
				}
			}
		}
		name := nd.Name
		if nd.Matched {
			name = color.RedString(name)
		}
		sb.WriteString(prefix + name + "\n")
		// traversal children
		for i, d := range nd.Deps {
			isOpen := true
			if i == len(nd.Deps)-1 {
				isOpen = false
			}
			flags = append(flags, isOpen)
			dfs(d, flags)
			flags = flags[:len(flags)-1]
		}
	}
	dfs(nd, nil)
	if !opts.noReport {
		fmt.Fprintf(
			sb, "%d deps, %d direct, %d indirect",
			r.Deps, r.Direct, r.Indirect)
	}
	return sb.String(), nil
}

// dfs traversal
func visit(d Dep, opts options) (dep, report) {
	var dfs func(
		d Dep, level int) (nd dep, depsCount int, directCount int, indirectCount int, filtered bool)
	dfs = func(
		d Dep, level int) (nd dep, depsCount int, directCount int, indirectCount int, filtered bool) {
		// filter out std or internal packages
		if opts.noStd && isStd(d.Name()) {
			filtered = true
			return
		}
		if opts.noInternal && isInternal(d.Name()) {
			filtered = true
			return
		}
		//
		var t Type
		switch level {
		case 0:
			t = Root
		case 1:
			t = Direct
			depsCount++
			directCount++
		default:
			t = Indirect
			depsCount++
			indirectCount++
		}
		nd = dep{
			Type:    t,
			Name:    d.Name(),
			Matched: !opts.filter.Filter(d.Name()),
		}
		filtered = !nd.Matched
		if opts.maxLevel > 0 && level >= opts.maxLevel {
			return
		}
		_deps := d.Deps()
		sort.Slice(_deps, func(i, j int) bool {
			return _deps[i].Name() < _deps[j].Name()
		})
		for _, _dep := range _deps {
			_nd, _depsCount, _directCount, _indirectCount, _filtered :=
				dfs(_dep, level+1)
			if _filtered {
				continue
			}
			filtered = false
			depsCount += _depsCount
			directCount += _directCount
			indirectCount += _indirectCount
			nd.Deps = append(nd.Deps, _nd)
		}
		return
	}
	nd, depsCount, directCount, indirectCount, _ := dfs(d, 0)
	return nd, report{
		Type:     Report,
		Deps:     depsCount,
		Direct:   directCount,
		Indirect: indirectCount,
	}
}

func isStd(name string) bool {
	_, ok := std.StdLib[name]
	return ok
}

func isInternal(name string) bool {
	if strings.HasPrefix(name, "internal/") {
		return true
	}
	if strings.HasSuffix(name, "/internal") {
		return true
	}
	return strings.Contains(name, "/internal/")
}
