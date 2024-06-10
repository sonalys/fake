package fake

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Generate(t *testing.T) {
	output := t.TempDir()
	// output := "out"
	os.RemoveAll(output) // no caching
	Run([]string{"testdata"}, output, nil)
	g, err := NewGenerator("mocks", "testdata")
	require.NoError(t, err)
	_, err = g.ParseFile(path.Join(output, "testdata", "stub.gen.go"))
	require.NoError(t, err)
}
