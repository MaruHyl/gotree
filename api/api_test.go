package api

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/packages/packagestest"
	"testing"
)

func TestOption_Init(t *testing.T) {
	packagestest.TestAll(t, testOptionInit)
}

func testOptionInit(t *testing.T, exporter packagestest.Exporter) {
	var opt = &Option{
		Test:           true,
		Mode:           "imports",
		BuildFlags:     nil,
		Args:           []string{"std"},
		SkipStd:        false,
		Level:          0,
		IncludePattern: "",
		ExcludePattern: "",
		IgnoreCase:     true,
	}
	err := opt.Init()
	require.NoError(t, err)
}

func TestOption_Load(t *testing.T) { packagestest.TestAll(t, testOptionLoad) }
func testOptionLoad(t *testing.T, exporter packagestest.Exporter) {
	var opt = &Option{
		Test:           false,
		Mode:           "imports",
		BuildFlags:     nil,
		Args:           []string{"std"},
		SkipStd:        false,
		Level:          0,
		IncludePattern: "",
		ExcludePattern: "",
		IgnoreCase:     true,
	}
	err := opt.Init()
	require.NoError(t, err)
	pkgs, err := opt.Load()
	require.NoError(t, err)
	t.Log(len(pkgs))
}

func TestOption_GetDeps(t *testing.T) { packagestest.TestAll(t, testOptionGetDeps) }
func testOptionGetDeps(t *testing.T, exporter packagestest.Exporter) {
	exported := packagestest.Export(t, exporter, []packagestest.Module{{
		Name: "gopkg/fake",
		Files: map[string]interface{}{
			"a/a.go": `package a; import (_ "gopkg/fake/b"; _ "gopkg/fake/c")`,
			"b/b.go": `package b; import (_ "gopkg/fake/c"; _ "gopkg/fake/d")`,
			"c/c.go": `package c; import (_ "strings";)`,
			"d/d.go": `package d; import (_ "gopkg/fake/c";)`,
		}}})
	defer exported.Cleanup()
	exported.Config.Mode = packages.LoadImports

	var opt = &Option{
		Test:           false,
		Mode:           "imports",
		BuildFlags:     nil,
		Args:           []string{"gopkg/fake/a"},
		SkipStd:        true,
		Level:          0,
		IncludePattern: "",
		ExcludePattern: "",
		IgnoreCase:     true,
	}
	err := opt.Init()
	require.NoError(t, err)
	opt.cfg = exported.Config
	pkgs, err := opt.Load()
	require.NoError(t, err)
	deps := opt.GetDeps(pkgs)
	require.Equal(t, 4, len(deps))
	require.Equal(t, "gopkg/fake/a", deps[0].ID)
	require.Equal(t, "gopkg/fake/b", deps[1].ID)
	require.Equal(t, "gopkg/fake/c", deps[2].ID)
	require.Equal(t, "gopkg/fake/d", deps[3].ID)
}

func TestGraphOption_Tree(t *testing.T) { packagestest.TestAll(t, testGraphOptionTree) }
func testGraphOptionTree(t *testing.T, exporter packagestest.Exporter) {
	exported := packagestest.Export(t, exporter, []packagestest.Module{{
		Name: "gopkg/fake",
		Files: map[string]interface{}{
			"a/a.go":   `package a; import (_ "gopkg/fake/b"; _ "gopkg/fake/c")`,
			"b/b.go":   `package b; import (_ "gopkg/fake/d1"; _ "gopkg/fake/d2")`,
			"c/c.go":   `package c; import (_ "gopkg/fake/b"; _ "gopkg/fake/f")`,
			"d1/d1.go": `package d1;`,
			"d2/d2.go": `package d2;`,
			"f/f.go":   `package f;`,
		}}})
	defer exported.Cleanup()
	exported.Config.Mode = packages.LoadImports

	var opt = &Option{
		Test:           false,
		Mode:           "imports",
		BuildFlags:     nil,
		Args:           []string{"gopkg/fake/a"},
		SkipStd:        true,
		Level:          0,
		IncludePattern: "",
		ExcludePattern: "",
		IgnoreCase:     true,
	}
	err := opt.Init()
	require.NoError(t, err)
	opt.cfg = exported.Config
	pkgs, err := opt.Load()
	require.NoError(t, err)
	require.Equal(t, 1, len(pkgs))
	// default option
	r := opt.Tree(pkgs[0])
	require.NoError(t, r.Error)
	t.Log("\n" + r.TreeContent)
	//
	opt.IncludePattern = "fake/d"
	opt.ExcludePattern = "d1"
	err = opt.Init()
	require.NoError(t, err)
	r = opt.Tree(pkgs[0])
	require.NoError(t, r.Error)
	t.Log("\n" + r.TreeContent)
	//
	opt.Level = 2
	err = opt.Init()
	require.NoError(t, err)
	r = opt.Tree(pkgs[0])
	require.NoError(t, r.Error)
	t.Log("\n" + r.TreeContent)
}
