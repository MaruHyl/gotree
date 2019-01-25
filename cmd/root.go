package cmd

import (
	"fmt"
	"github.com/MaruHyl/gopkg/api"
	"github.com/spf13/cobra"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

type Config struct {
	api.Option
}

var cfg = new(Config)

func init() {
	rootCmd.PersistentFlags().BoolVarP(
		&cfg.Test, "test", "t", false,
		"include any tests implied by the patterns")
	rootCmd.PersistentFlags().StringVarP(
		&cfg.Mode, "mode", "m", "imports",
		"mode (one of files, imports, types, syntax, allsyntax)")
	rootCmd.PersistentFlags().StringSliceVarP(
		&cfg.BuildFlags, "buildflags", "b", nil,
		"pass argument to underlying build system (may be repeated)")
	rootCmd.PersistentFlags().StringSliceVarP(
		&cfg.Args, "args", "a", nil,
		"Args are passed to go/packages directly (may be repeated)")
	rootCmd.AddCommand(depsCmd)

	depsCmd.PersistentFlags().BoolVarP(&cfg.SkipStd, "skip-std", "s", false,
		"Skip std library")
	depsCmd.PersistentFlags().IntVarP(&cfg.Level, "level", "l", 0,
		"Max deep level")
	depsCmd.PersistentFlags().StringVarP(&cfg.IncludePattern, "include", "p", "",
		"Only list packages that match the pattern")
	depsCmd.PersistentFlags().StringVarP(&cfg.ExcludePattern, "exclude", "i", "",
		"Only list packages that not match the pattern")
	depsCmd.PersistentFlags().BoolVarP(&cfg.IgnoreCase, "case-insensitive", "c", false,
		"default(false): case insensitive")
	depsCmd.AddCommand(treeCmd)
}

func getNonErrorPkgs() ([]*packages.Package, error) {
	err := cfg.Init()
	if err != nil {
		return nil, err
	}
	pkgs, err := cfg.Load()
	if err != nil {
		return nil, err
	}
	err = api.CheckError(pkgs)
	if err != nil {
		return nil, err
	}
	return pkgs, nil
}

var rootCmd = &cobra.Command{
	Use:   "gopkg",
	Short: "Show package list.",
	Long:  `Show package list.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pkgs, err := getNonErrorPkgs()
		if err != nil {
			log.Fatal(err)
		}
		var pkgIDs = make([]string, 0, len(pkgs))
		for _, pkg := range pkgs {
			pkgIDs = append(pkgIDs, pkg.ID)
		}
		fmt.Println(strings.Join(pkgIDs, "\n"))
	},
}

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Show dep list.",
	Long:  `Show dep list.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pkgs, err := getNonErrorPkgs()
		if err != nil {
			log.Fatal(err)
		}
		allPkgs := cfg.GetDeps(pkgs)
		var pkgIDs = make([]string, 0, len(allPkgs))
		for _, pkg := range allPkgs {
			pkgIDs = append(pkgIDs, pkg.ID)
		}
		fmt.Println(strings.Join(pkgIDs, "\n"))
	},
}

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Show deps like tree",
	Long: `Show deps like tree. 
Support regexp(include and exclude), max deep level, json(todo), svg(todo).`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		pkgs, err := getNonErrorPkgs()
		if err != nil {
			log.Fatal(err)
		}
		var sb = new(strings.Builder)
		for _, pkg := range pkgs {
			fmt.Fprintf(sb, "==== %s\n", pkg.ID)
			ret := cfg.Tree(pkg)
			if ret.Error != nil {
				fmt.Fprintf(sb, "occurer error %v\n", ret.Error)
				continue
			}
			fmt.Fprintln(sb, ret.TreeContent)
		}
		fmt.Println(sb.String())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
