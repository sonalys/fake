package fake

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Generate(t *testing.T) {
	output := t.TempDir()
	Run([]string{"testdata"}, output, []string{})
	g := NewGenerator("mocks")
	_, err := g.ParseFile(path.Join(output, "testdata", "stub.gen.go"))
	require.NoError(t, err)
}
