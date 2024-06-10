package files

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ListGoFiles lists all Go files under a directory.
func ListGoFiles(dirs, ignore []string) ([]string, error) {
	var goFiles []string
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(filename string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			for _, entry := range ignore {
				if matched, _ := path.Match(entry, dir); matched {
					return nil
				}
			}
			// Check if the file has a ".go" extension
			if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") && !strings.HasSuffix(info.Name(), ".gen.go") {
				goFiles = append(goFiles, filename)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return goFiles, nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateFileAndFolders creates a file and the necessary folders if they don't exist.
func CreateFileAndFolders(filePath string) (*os.File, error) {
	// Get the directory path from the file path
	dir := filepath.Dir(filePath)
	// Create directories if they don't exist
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// FindFile searches for the specified file in the given directory and its parent directories.
func FindFile(childDir, fileName string) (string, error) {
	abs, err := filepath.Abs(childDir)
	if err != nil {
		return "", fmt.Errorf("could not use path %s: %w", childDir, err)
	}
	for {
		filePath := filepath.Join(abs, fileName)
		if fileExists(filePath) {
			return filePath, nil
		}

		parentDir := filepath.Dir(abs)
		if parentDir == abs {
			break
		}
		abs = parentDir
	}

	return "", fmt.Errorf("%s file not found", fileName)
}

// GetPackagePath returns the absolute path of the file package, including the module path.
// This function is not considering packages with different names from their respective folders,
// the reason is that this software is not made for psychopaths.
func GetPackagePath(goModPath string, modFile *modfile.File, filename string) (string, error) {
	// Retrieve the module path
	modulePath := modFile.Module.Mod.Path
	pkgPath, err := GetRelativePath(goModPath, path.Dir(filename))
	if err != nil {
		return "", err
	}
	return path.Join(modulePath, pkgPath), nil
}

// GetRelativePath returns the shared path between two paths.
// if they are in the same folder, they will return the same folder path.
// Example: /path1/folder1/file and /path1/folder2/file2 should return /path1.
func GetRelativePath(path1, path2 string) (string, error) {
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
