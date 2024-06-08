package fake

import (
	"go/ast"
	"go/token"
	"os"
	"path"

	"github.com/sonalys/fake/internal/files"
	"github.com/sonalys/fake/internal/imports"
	"golang.org/x/mod/modfile"
)

// Generator is the controller for the whole module, caching files and holding metadata.
type Generator struct {
	FileSet         *token.FileSet
	MockPackageName string

	cachedPackageInfo func(f *ast.File) (nameMap, pathMap map[string]*imports.ImportEntry)
	goModFilename     string
	goMod             *modfile.File
}

// NewGenerator will create a new mock generator for the specified module.
func NewGenerator(pkgName, baseDir string) (*Generator, error) {
	goModPath, err := files.FindFile(baseDir, "go.mod")
	if err != nil {
		return nil, err
	}
	// Read the contents of the go.mod file
	modFileContent, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	// Parse the go.mod file
	modFile, err := modfile.Parse(goModPath, modFileContent, nil)
	if err != nil {
		return nil, err
	}

	return &Generator{
		FileSet:           token.NewFileSet(),
		goModFilename:     goModPath,
		goMod:             modFile,
		MockPackageName:   pkgName,
		cachedPackageInfo: imports.CachedImportInformation(path.Dir(goModPath)),
	}, nil
}
