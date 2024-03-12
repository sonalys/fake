package hashCheck

import (
	"bufio"
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
	"strings"
)

const (
	fileName = "fake.lock.json"
)

func CompareFileHashes(inputDirs, ignore []string) ([]string, error) {
	res := make([]string, 0)

	dir, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("os.Getwd: %w", err)
	}

	goModPath, err := fake.FindFile(dir, "go.sum")

	if err != nil {
		return nil, fmt.Errorf("fake.FindFile: %w", err)
	}

	dependenciesParsed, err := parseGoSumFile(goModPath)

	if err != nil {
		return nil, fmt.Errorf("parseGoSumFile: %w", err)
	}

	dependencies := make(map[string]string)

	for k, v := range dependenciesParsed {
		dependencies[k], err = hashFiles(v...)
		if err != nil {
			return nil, fmt.Errorf("parseGoSumFile: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("parseGoSumFile: %w", err)
	}

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

				imports, err := loadPackageImports(file)

				importHash := make([]string, len(imports))

				if err != nil {
					return nil, fmt.Errorf("loadPackageImports: %w", err)
				}

				for _, importName := range imports {
					importHash = append(importHash, dependencies[importName])
				}

				goSum, err := hashFiles(importHash...)
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
			err = saveHashToFile(group, model)
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
// parses file from mocks/{path}/fake.lock.json
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

// parseGoSumFile reads and parses the go.sum file into a map
// import : []string{all related hashes}
func parseGoSumFile(path string) (map[string][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	dependencies := make(map[string][]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) == 3 {
			dependencies[parts[0]] = append(dependencies[parts[0]], parts[2])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return dependencies, nil

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

/*
saveHashToFile function takes dir string
and the target directory (dir), as well as a hash map (hash).
It saves file at path mocks/{dir}/fake.lock.json
*/
func saveHashToFile(dir string, hash map[string]FileHashData) error {
	data, err := json.Marshal(hash)
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Join("mocks", dir), os.ModePerm)
	if err != nil {
		return err
	}

	os.WriteFile(filepath.Join("mocks", dir, fileName), data, 0644)

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
