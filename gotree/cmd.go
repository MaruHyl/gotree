package main

import (
	"fmt"

	"github.com/MaruHyl/gotree"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cmd.Flags().IntVarP(
		&maxLevel, "max_level", "l", 0, "Set max level of tree")
	cmd.Flags().BoolVar(&noReport, "noreport", false, "Turn off dep/direct/indirect count at end of tree listing")
	cmd.Flags().StringVarP(
		&pattern, "pattern", "p", "", "List only those deps that match the pattern given")
	cmd.Flags().BoolVarP(&json, "json", "j", false, "Prints out an JSON representation of the tree")
	cmd.Flags().BoolVar(&noStd, "nostd", false, "Filter out std packages")
	cmd.Flags().BoolVar(&noInternal, "nointernal", false, "Filter out internal packages")
}

var maxLevel int
var noReport bool
var pattern string
var json bool
var noStd bool
var noInternal bool

var cmd = &cobra.Command{
	Use:   "gotree",
	Short: "Show deps like tree.",
	Long:  "Show deps like tree.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		opts := []gotree.Option{
			gotree.WithMaxLevel(maxLevel),
			gotree.WithNoReport(noReport),
			gotree.WithNoStd(noStd),
			gotree.WithNoInternal(noInternal),
		}
		if pattern != "" {
			patternFilter, err := gotree.NewRegexpFilter(pattern)
			if err != nil {
				panic(fmt.Errorf("compile pattern error: %v", err))
			}
			opts = append(opts, gotree.WithFilter(gotree.NewReverseFilter(patternFilter)))
		}
		root, err := gotree.LoadPackages()
		if err != nil {
			panic(err)
		}
		var str string
		if json {
			str, err = gotree.JSONTree(dep{root}, opts...)
			if err != nil {
				panic(fmt.Errorf("build json error: %v", err))
			}
		} else {
			str, err = gotree.Tree(dep{root}, opts...)
			if err != nil {
				panic(fmt.Errorf("build tree error: %v", err))
			}
		}
		fmt.Println(str)
	},
}

type dep struct {
	pkg *packages.Package
}

func (d dep) Name() string {
	return d.pkg.PkgPath
}

func (d dep) Deps() []gotree.Dep {
	imports := d.pkg.Imports
	deps := make([]gotree.Dep, 0, len(imports))
	for _, i := range imports {
		deps = append(deps, dep{i})
	}
	return deps
}
