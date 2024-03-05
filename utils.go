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

// getRelativePath returns the shared path between two paths.
// if they are in the same folder, they will return the same folder path.
// Example: /path1/folder1/file and /path1/folder2/file2 should return /path1.
func getRelativePath(path1, path2 string) (string, error) {
	// Make the paths absolute to ensure accurate relative path calculation
	absPath1, err := filepath.Abs(path1)
	if err != nil {
		return "", err
	}
	absPath2, err := filepath.Abs(path2)
	if err != nil {
		return "", err
	}
	// Retrieve the relative path
	relativePath, err := filepath.Rel(filepath.Dir(absPath1), absPath2)
	if err != nil {
		return "", err
	}
	return relativePath, nil
}

// findGoMod searches for the go.mod file in the given directory and its parent directories.
func findGoMod(dir string) (string, error) {
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if fileExists(goModPath) {
			return goModPath, nil
		}
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}
	return "", fmt.Errorf("go.mod file not found")
}

// GetPackagePath returns the full package path from a given *ast.File.
func GetPackagePath(fset *token.FileSet, filename string) (string, error) {
	goModPath, err := findGoMod(filepath.Dir(filename))
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
	pkgPath, _ := getRelativePath(goModPath, path.Dir(filename))
	return path.Join(modulePath, pkgPath), nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
