package gotree

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

func LoadPackages() (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.LoadImports,
	}
	root, err := packages.Load(cfg, "")
	if err != nil {
		return nil, fmt.Errorf("load packages error: %v", err)
	}
	if len(root) != 1 {
		return nil, fmt.Errorf("unsupported packages number: %d", len(root))
	}
	packages.PrintErrors(root)
	return root[0], nil
}
