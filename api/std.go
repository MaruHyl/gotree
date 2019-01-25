package api

import (
	"golang.org/x/tools/go/packages"
	"strings"
)

var GoStdList string

func GetStdMap() (map[string]struct{}, error) {
	if GoStdList != "" {
		// fast-path
		var strSlice = strings.Split(GoStdList, "\n")
		var stdMap = make(map[string]struct{}, len(strSlice))
		for _, s := range strSlice {
			stdMap[s] = struct{}{}
		}
		return stdMap, nil
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadImports}, "std")
	if err != nil {
		return nil, err
	}
	var stdMap = make(map[string]struct{}, len(pkgs))
	for _, pkg := range pkgs {
		if strings.HasPrefix(pkg.ID, "vendor") {
			continue
		}
		stdMap[pkg.ID] = struct{}{}
	}
	return stdMap, nil
}
