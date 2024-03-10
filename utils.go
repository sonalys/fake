package fake

import (
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/mod/modfile"
)

// GetPackagePath returns the full package path from a given *ast.File.
func GetPackagePath(fset *token.FileSet, filename string) (string, error) {
	goModPath, err := FindParentFile(filepath.Dir(filename), "go.mod")
	if err != nil {
		return "", err
	}
	// Read the contents of the go.mod file
	modFileContent, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	// Parse the go.mod file
	modFile, err := modfile.Parse(goModPath, modFileContent, nil)
	if err != nil {
		return "", err
	}
	// Retrieve the module path
	modulePath := modFile.Module.Mod.Path
	pkgPath, _ := GetRelativePath(goModPath, path.Dir(filename))
	return path.Join(modulePath, pkgPath), nil
}

func generateOutputFile(input, output string) *os.File {
	filename, _ := strings.CutSuffix(path.Base(input), ".go")
	outputFile := path.Join(output, fmt.Sprintf("%s.gen.go", filename))
	outFile, err := CreateFileAndFolders(outputFile)
	if err != nil {
		log.Panic().Msgf("Error creating mock file: %v\n", err)
	}
	return outFile
}
