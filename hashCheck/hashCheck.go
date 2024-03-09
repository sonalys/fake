package hashCheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sonalys/fake"
	"golang.org/x/tools/go/packages"
	"io"
	"os"
	"path/filepath"
)

const (
	fileName = "fake.lock.json"
)

func CompareFileHashes(inputDirs, ignore []string) ([]string, error) {
	res := make([]string, 0)

	for _, dir := range inputDirs {

		files, err := fake.ListGoFiles(dir, ignore)
		if err != nil {
			return nil, fmt.Errorf("getFiles: %w", err)
		}

		groups := groupByDirectory(files)

		for group, data := range groups {
			model, err := parseJsonModel(group)

			if err != nil {
				return nil, fmt.Errorf("parseJsonModel: %w", err)
			}

			for _, file := range data {

				imports, err := getPackagesGosum(file)
				if err != nil {
					return nil, fmt.Errorf("getImportPath: %w", err)
				}

				goSum, err := hashFiles(imports...)
				if err != nil {
					return nil, fmt.Errorf("hashFiles: %w", err)
				}

				hash, err := hashFiles(file)
				if err != nil {
					return nil, fmt.Errorf("hashFiles: %w", err)
				}

				if model[file].Hash != hash || model[file].GoSum != goSum {
					res = append(res, file)
				}

				model[file] = FileHashData{
					Hash:  hash,
					GoSum: goSum,
				}

			}
			err = saveHashToFile(dir, group, model)
			if err != nil {
				return nil, fmt.Errorf("saveHashToFile: %w", err)
			}

		}

	}

	return res, nil
}

/*
groupByDirectory groups files by their directory
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
func groupByDirectory(files []string) map[string][]string {
	groups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

// parseJsonModel reads and parses the json model from the fake.lock.json file
func parseJsonModel(path string) (Hashes, error) {
	data, err := os.ReadFile(filepath.Join("mocks", path, fileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Hashes{}, nil
		}
		return nil, err
	}

	var model Hashes
	err = json.Unmarshal(data, &model)
	if err != nil {
		return nil, err
	}

	return model, nil
}

/*
getPackagesGosum takes path to a .go file
returns a list of go.sum files for the given file's dependencies
*/
func getPackagesGosum(file string) ([]string, error) {
	res := make([]string, 0)

	imports, err := loadPackageImports(file)
	if err != nil {
		return nil, fmt.Errorf("loadPackageImports: %w", err)
	}

	for _, importName := range imports {
		path := getPackagePath(importName)

		res = append(res, path)
	}

	return res, nil
}

// loadPackageImports returns a list of imports for a given .go file
func loadPackageImports(file string) ([]string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, file)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		return nil, errors.New("packages.Load: errors while loading package")
	}

	var imports []string
	for _, pkg := range pkgs {
		for imp := range pkg.Imports {
			imports = append(imports, imp)
		}
	}

	return imports, nil
}

// getPackagePath returns the path to the go.sum file for a given import path
func getPackagePath(importPath string) string {
	gopath := os.Getenv("GOPATH")
	goSumPath := filepath.Join(gopath, "pkg", "mod", importPath, "go.sum")
	return goSumPath
}

/*
The saveHashToFile function takes two strings representing the root directory (root) from user input
and the target directory (dir), as well as a hash map (hash).
This function saves the hash map to a fake.lock.json file in the specified directory.
*/
func saveHashToFile(root, dir string, hash map[string]FileHashData) error {
	fmt.Printf("root: %s\n", root)
	fmt.Printf("dir: %s\n", dir)
	data, err := json.Marshal(hash)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join(root, "mocks", dir), os.ModePerm)
	if err != nil {
		return err
	}

	os.WriteFile(filepath.Join(root, "mocks", dir, fileName), data, 0644)

	return nil
}

// hashFiles returns the SHA256 hash of files
func hashFiles(files ...string) (string, error) {
	hasher := sha256.New()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", err
		}

		if _, err := io.Copy(hasher, f); err != nil {
			f.Close()
			return "", err
		}

		f.Close()
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
