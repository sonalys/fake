package files

import (
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ListGoFiles lists all Go files under a directory.
func ListGoFiles(paths, ignore []string) ([]string, error) {
	var goFiles []string
	for _, path := range paths {
		err := filepath.Walk(path, func(filename string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			for _, entry := range ignore {
				if _, ok := strings.CutPrefix(filename, entry); ok {
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
	for {
		filePath := filepath.Join(childDir, fileName)
		if fileExists(filePath) {
			return filePath, nil
		}

		parentDir := filepath.Dir(childDir)
		if parentDir == childDir {
			break
		}
		childDir = parentDir
	}

	return "", fmt.Errorf("%s file not found", fileName)
}

// GetPackagePath returns the full package path from a given *ast.File.
func GetPackagePath(fset *token.FileSet, filename string) (string, error) {
	goModPath, err := FindFile(filepath.Dir(filename), "go.mod")
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

/*
GroupByDirectory groups files by their directory
Example:

Input:

	files := []string{
		"/home/user/documents/file1.txt",
		"/home/user/documents/file2.txt",
		"/home/user/images/image1.png",
		"/home/user/images/image2.png",
		"/home/user/images/image3.png",
	}

Output:

		{
		"/home/user/documents": []string{
			"/home/user/documents/file1.txt",
			"/home/user/documents/file2.txt",
		},
		"/home/user/images": []string{
			"/home/user/images/image1.png",
			"/home/user/images/image2.png",
			"/home/user/images/image3.png",
		},
	}
*/
func GroupByDirectory(files []string) map[string][]string {
	groups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}
