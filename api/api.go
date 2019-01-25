package api

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/tools/go/packages"
	"regexp"
	"sort"
	"strings"
)

type Option struct {
	Test           bool
	Mode           string
	BuildFlags     []string
	Args           []string
	SkipStd        bool
	Level          int
	IncludePattern string
	ExcludePattern string
	IgnoreCase     bool
	cfg            *packages.Config
	includeRegexp  *regexp.Regexp
	excludeRegexp  *regexp.Regexp
	stdMap         map[string]struct{}
}

func (opt *Option) Init() error {
	var err error
	// build std map
	if opt.SkipStd {
		opt.stdMap, err = GetStdMap()
		if err != nil {
			return err
		}
	}
	// build cfg
	opt.cfg, err = opt.parseConfig()
	if err != nil {
		return err
	}
	var caseInsensitive = ""
	// build regexp
	if opt.IgnoreCase {
		caseInsensitive += "(?i)"
	}
	if len(opt.IncludePattern) > 0 {
		opt.includeRegexp, err = regexp.Compile(caseInsensitive + opt.IncludePattern)
		if err != nil {
			return err
		}
	}
	if len(opt.ExcludePattern) > 0 {
		opt.excludeRegexp, err = regexp.Compile(caseInsensitive + opt.ExcludePattern)
		if err != nil {
			return err
		}
	}
	return nil
}

func (opt *Option) parseConfig() (*packages.Config, error) {
	pcfg := &packages.Config{
		Mode:       packages.LoadImports,
		Tests:      opt.Test,
		BuildFlags: opt.BuildFlags,
	}

	switch strings.ToLower(opt.Mode) {
	case "files":
		pcfg.Mode = packages.LoadFiles
	case "imports":
		pcfg.Mode = packages.LoadImports
	case "types":
		pcfg.Mode = packages.LoadTypes
	case "syntax":
		pcfg.Mode = packages.LoadSyntax
	case "allsyntax":
		pcfg.Mode = packages.LoadAllSyntax
	default:
		return nil, fmt.Errorf("unavailable mode %s", opt.Mode)
	}

	return pcfg, nil
}

func (opt *Option) Load() ([]*packages.Package, error) {
	pkgs, err := packages.Load(opt.cfg, opt.Args...)
	if err != nil {
		return nil, err
	}
	return pkgs, CheckError(pkgs)
}

func CheckError(pkgs []*packages.Package) error {
	var mErr *multierror.Error
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			mErr = multierror.Append(mErr, err)
		}
	})
	return mErr.ErrorOrNil()
}

func (opt *Option) GetDeps(pkgs []*packages.Package) []*packages.Package {
	var loadedPkgs = make([]*packages.Package, 0)
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if !opt.filterPackages(pkg) {
			loadedPkgs = append(loadedPkgs, pkg)
		}
	})
	sort.Slice(loadedPkgs, func(i, j int) bool {
		return loadedPkgs[i].ID < loadedPkgs[j].ID
	})
	return loadedPkgs
}

func (opt *Option) filterPackages(pkg *packages.Package) bool {
	if opt.SkipStd {
		_, ok := opt.stdMap[pkg.ID]
		if ok {
			return true
		}
	}
	if opt.includeRegexp != nil && !opt.includeRegexp.MatchString(pkg.ID) {
		return true
	} else if opt.excludeRegexp != nil && opt.excludeRegexp.MatchString(pkg.ID) {
		return true
	}
	return false
}

type Result struct {
	Error       error
	TreeContent string
	JsonContent []byte
}

type jsonPkg struct {
	ID      string    `json:"pkg"`
	Imports []jsonPkg `json:"imports"`
}

type graph struct {
	nodes map[string]*node // key: pkg.PkgPath
}

type node struct {
	pkg *packages.Package
	in  []*node
	out []*node
}

func (opt *Option) build(pkg *packages.Package) *graph {
	var g = new(graph)
	g.nodes = make(map[string]*node)
	var queue = make([]*packages.Package, 0)
	queue = append(queue, pkg)
	for len(queue) > 0 {
		var pkg = queue[0]
		queue = queue[1:]
		n, visited := g.nodes[pkg.ID]
		if !visited {
			n = new(node)
			n.pkg = pkg
			g.nodes[pkg.ID] = n
		}
		for _, importPkg := range pkg.Imports {
			importNode, importVisited := g.nodes[importPkg.PkgPath]
			if !importVisited {
				importNode = new(node)
				importNode.pkg = importPkg
				g.nodes[importPkg.ID] = importNode
				queue = append(queue, importPkg)
			}
			n.out = append(n.out, importNode)
			importNode.in = append(importNode.in, n)
		}
	}
	for _, node := range g.nodes {
		sort.Slice(node.in, func(i, j int) bool {
			return node.in[i].pkg.ID < node.in[j].pkg.ID
		})
		sort.Slice(node.out, func(i, j int) bool {
			return node.out[i].pkg.ID < node.out[j].pkg.ID
		})
	}
	return g
}

// WARN: import cycle will cause stack overflow
func (opt *Option) Tree(pkg *packages.Package) Result {
	var result Result
	var g = opt.build(pkg)
	// visit all nodes
	var tree = true
	var treeStr = make([][]string, 0)
	var treeFlag = make([]bool, 0)
	const treeLast = "`-- "
	const treeNotLast = "|-- "
	const treeAncestorLast = "    "
	const treeAncestorNotLast = "|   "
	var level = 0
	var matchMap = make(map[string]bool, len(g.nodes))
	var visit func(n *node) bool
	visit = func(n *node) bool {
		//
		if opt.Level > 0 && level > opt.Level {
			return false
		}
		//
		match, ok := matchMap[n.pkg.ID]
		if !ok {
			match = !opt.filterPackages(n.pkg)
			matchMap[n.pkg.ID] = match
		}
		//
		if tree {
			var prefix = make([]string, 0, len(treeFlag)+1)
			for i, flag := range treeFlag {
				if i == len(treeFlag)-1 {
					if flag {
						prefix = append(prefix, treeNotLast)
					} else {
						prefix = append(prefix, treeLast)
					}
				} else {
					if flag {
						prefix = append(prefix, treeAncestorNotLast)
					} else {
						prefix = append(prefix, treeAncestorLast)
					}
				}
			}
			var pkgId = n.pkg.ID
			if match {
				pkgId = color.BlueString(pkgId)
			}
			prefix = append(prefix, pkgId)
			treeStr = append(treeStr, prefix)
		}
		var inPath = false
		for i, out := range n.out {
			if tree {
				treeFlag = append(treeFlag, i < len(n.out)-1)
			}
			level++
			inPath = visit(out) || inPath
			level--
			if tree {
				treeFlag = treeFlag[:len(treeFlag)-1]
			}
		}
		if !inPath && !match {
			if tree {
				str := treeStr[len(treeStr)-1]
				treeStr = treeStr[:len(treeStr)-1]
				if len(str) > 1 {
					var index = len(str) - 2
					var flag = str[index]
					switch flag {
					case treeLast:
					loop:
						for i := len(treeStr) - 1; i >= 0; i-- {
							str := treeStr[i]
							upFlag := str[index]
							switch upFlag {
							case treeLast:
								panic("no possible1")
							case treeNotLast:
								str[index] = treeLast
								break loop
							case treeAncestorLast:
								panic("no possible2")
							case treeAncestorNotLast:
								str[index] = treeAncestorLast
							default:
								break loop
							}
						}
					case treeNotLast:
						// do nothing
					}
				}
			}
			return false
		}
		return true
	}
	visit(g.nodes[pkg.ID])
	var treeStrCombine = make([]string, 0, len(treeStr))
	for _, str := range treeStr {
		treeStrCombine = append(treeStrCombine, strings.Join(str, ""))
	}
	result.TreeContent = strings.Join(treeStrCombine, "\n")
	return result
}
