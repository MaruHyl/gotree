package gotree_test

import (
	"strings"
	"testing"

	"github.com/MaruHyl/gotree"
	"github.com/stretchr/testify/require"
)

type mockDep struct {
	name string
	deps []gotree.Dep
}

func (n *mockDep) Name() string {
	return n.name
}

func (n *mockDep) Deps() []gotree.Dep {
	return n.deps
}

func getCompleteDep() gotree.Dep {
	// init nodes
	nodeMap := make(map[string]*mockDep)
	nodeMap["a"] = &mockDep{name: "a"}
	nodeMap["b"] = &mockDep{name: "b"}
	nodeMap["c"] = &mockDep{name: "c"}
	nodeMap["d"] = &mockDep{name: "d"}
	nodeMap["e"] = &mockDep{name: "e"}
	nodeMap["f"] = &mockDep{name: "f"}
	nodeMap["g"] = &mockDep{name: "g"}
	nodeMap["h"] = &mockDep{name: "h"}
	nodeMap["i"] = &mockDep{name: "i"}
	// build graph
	nodeMap["c"].deps = []gotree.Dep{nodeMap["d"], nodeMap["e"]}
	nodeMap["f"].deps = []gotree.Dep{nodeMap["g"], nodeMap["h"]}
	nodeMap["b"].deps = []gotree.Dep{nodeMap["c"]}
	nodeMap["a"].deps = []gotree.Dep{nodeMap["b"], nodeMap["f"]}
	nodeMap["h"].deps = []gotree.Dep{nodeMap["i"]}
	return nodeMap["a"]
}

func TestTree_Json_Empty(t *testing.T) {
	//
	json, err := gotree.JSONTree(nil)
	require.NoError(t, err)
	require.Equal(t, "", json)
	//
	json, err = gotree.JSONTree(&mockDep{
		name: "root",
	})
	require.NoError(t, err)
	require.Equal(t, `
[
 {
  "Type": "root",
  "Name": "root",
  "Matched": true
 },
 {
  "Type": "report",
  "Deps": 0,
  "Direct": 0,
  "Indirect": 0
 }
]`, "\n"+json)
}

func TestTree_Json(t *testing.T) {
	//
	json, err := gotree.JSONTree(getCompleteDep())
	require.NoError(t, err)
	require.Equal(t, `
[
 {
  "Type": "root",
  "Name": "a",
  "Matched": true,
  "Deps": [
   {
    "Type": "direct",
    "Name": "b",
    "Matched": true,
    "Deps": [
     {
      "Type": "indirect",
      "Name": "c",
      "Matched": true,
      "Deps": [
       {
        "Type": "indirect",
        "Name": "d",
        "Matched": true
       },
       {
        "Type": "indirect",
        "Name": "e",
        "Matched": true
       }
      ]
     }
    ]
   },
   {
    "Type": "direct",
    "Name": "f",
    "Matched": true,
    "Deps": [
     {
      "Type": "indirect",
      "Name": "g",
      "Matched": true
     },
     {
      "Type": "indirect",
      "Name": "h",
      "Matched": true,
      "Deps": [
       {
        "Type": "indirect",
        "Name": "i",
        "Matched": true
       }
      ]
     }
    ]
   }
  ]
 },
 {
  "Type": "report",
  "Deps": 8,
  "Direct": 2,
  "Indirect": 6
 }
]`, "\n"+json)
	//
	json, err = gotree.JSONTree(getCompleteDep(), gotree.WithMaxLevel(2))
	require.NoError(t, err)
	require.Equal(t, `
[
 {
  "Type": "root",
  "Name": "a",
  "Matched": true,
  "Deps": [
   {
    "Type": "direct",
    "Name": "b",
    "Matched": true,
    "Deps": [
     {
      "Type": "indirect",
      "Name": "c",
      "Matched": true
     }
    ]
   },
   {
    "Type": "direct",
    "Name": "f",
    "Matched": true,
    "Deps": [
     {
      "Type": "indirect",
      "Name": "g",
      "Matched": true
     },
     {
      "Type": "indirect",
      "Name": "h",
      "Matched": true
     }
    ]
   }
  ]
 },
 {
  "Type": "report",
  "Deps": 5,
  "Direct": 2,
  "Indirect": 3
 }
]`, "\n"+json)
}

func TestTree_Empty(t *testing.T) {
	//
	tree, err := gotree.Tree(nil)
	require.NoError(t, err)
	require.Equal(t, "", tree)
	//
	tree, err = gotree.Tree(&mockDep{
		name: "root",
	})
	require.NoError(t, err)
	require.Equal(t, `
root
0 deps, 0 direct, 0 indirect`, "\n"+tree)
}

func TestTree(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		const result = `
a
├── b
│   └── c
│       ├── d
│       └── e
└── f
    ├── g
    └── h
        └── i
8 deps, 2 direct, 6 indirect`
		tree, err := gotree.Tree(getCompleteDep())
		require.NoError(t, err)
		require.Equal(t, strings.TrimPrefix(result, "\n"), tree)
	})

	t.Run("maxLevel", func(t *testing.T) {
		const result = `
a
├── b
│   └── c
└── f
    ├── g
    └── h
5 deps, 2 direct, 3 indirect`
		tree, err := gotree.Tree(getCompleteDep(), gotree.WithMaxLevel(2))
		require.NoError(t, err)
		require.Equal(t, strings.TrimPrefix(result, "\n"), tree)
	})
}

type fixedFilter struct {
	target string
}

func (f fixedFilter) Filter(name string) bool {
	return f.target != name
}

func TestFilter(t *testing.T) {
	json, err := gotree.JSONTree(getCompleteDep(), gotree.WithFilter(fixedFilter{"c"}))
	require.NoError(t, err)
	require.Equal(t, `
[
 {
  "Type": "root",
  "Name": "a",
  "Matched": false,
  "Deps": [
   {
    "Type": "direct",
    "Name": "b",
    "Matched": false,
    "Deps": [
     {
      "Type": "indirect",
      "Name": "c",
      "Matched": true
     }
    ]
   }
  ]
 },
 {
  "Type": "report",
  "Deps": 2,
  "Direct": 1,
  "Indirect": 1
 }
]`, "\n"+json)

	tree, err := gotree.Tree(getCompleteDep(), gotree.WithFilter(fixedFilter{"c"}))
	require.NoError(t, err)
	require.Equal(t, `
a
└── b
    └── c
2 deps, 1 direct, 1 indirect`, "\n"+tree)
}
