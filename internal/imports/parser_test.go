package imports

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/sonalys/fake/internal/packages"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	// Create a new file set
	fset := token.NewFileSet()

	// Parse the file
	f, err := parser.ParseFile(fset, "../../testdata/stub.go", nil, 0)
	require.NoError(t, err)

	got, _ := FileListUsedImports(f)

	exp := []ImportEntry{
		{PackageInfo: packages.PackageInfo{Name: "anotherpkg", Path: "github.com/sonalys/fake/testdata/anotherpkg"}},
		{PackageInfo: packages.PackageInfo{Name: "time", Path: "time"}},
		{PackageInfo: packages.PackageInfo{Name: "testing", Path: "testing"}},
		{PackageInfo: packages.PackageInfo{Name: "require", Path: "github.com/stretchr/testify/require"}, Alias: "stub"},
	}
	require.ElementsMatch(t, exp, got)
}
