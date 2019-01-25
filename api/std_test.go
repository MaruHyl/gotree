package api

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetStdMap_Fast(t *testing.T) {
	m, err := GetStdMap()
	require.NoError(t, err)
	fmt.Println(len(m))
}

func TestGetStdMap_Slow(t *testing.T) {
	GoStdList = ""
	m, err := GetStdMap()
	require.NoError(t, err)
	fmt.Println(len(m))
}
