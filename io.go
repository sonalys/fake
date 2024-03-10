package fake

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

// ListGoFiles lists all Go files under a directory.
func ListGoFiles(dirPath string, ignore []string) ([]string, error) {
	var goFiles []string
	err := filepath.Walk(dirPath, func(filename string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		for _, entry := range ignore {
			if _, ok := strings.CutPrefix(filename, entry); ok {
				return nil
			}
		}
		// Check if the file has a ".go" extension
		if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
			goFiles = append(goFiles, filename)
		}
		return nil
	})
	return goFiles, err
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

// FileExists checks if a file exists at the given path.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

// FindParentFile searches for the go.mod file in the given directory and its parent directories.
func FindParentFile(childDir, filename string) (string, error) {
	for {
		goModPath := filepath.Join(childDir, filename)
		if FileExists(goModPath) {
			return goModPath, nil
		}
		parentDir := filepath.Dir(childDir)
		if parentDir == childDir {
			break
		}
		childDir = parentDir
	}
	return "", fmt.Errorf("%s file not found", filename)
}
