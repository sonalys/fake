package fake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Generate(t *testing.T) {
	Run([]string{"testdata"}, "testdata/out", []string{"testdata/out"})
	g := NewGenerator("mocks")
	_, err := g.ParseFile("testdata/out/testdata/stub.gen.go")
	require.NoError(t, err)
}
