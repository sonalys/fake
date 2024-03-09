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

// CompareFileHashes identifies changed files in directories by comparing current to stored hashes.
// Returns a list of changed files and any errors encountered.
func CompareFileHashes(inputDirs, ignore []string) ([]string, error) {
	res := make([]string, 0)

	for _, dir := range inputDirs {
		files, err := getGoFiles(dir, ignore)
		if err != nil {
			return nil, err
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

func groupByDirectory(files []string) map[string][]string { //V2
	groups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

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

func getPackagesGosum(file string) ([]string, error) {
	res := make([]string, 0)

	imports, err := loadPackageInfo(file)
	if err != nil {
		return nil, fmt.Errorf("loadPackageInfo: %w", err)
	}

	for _, importName := range imports {
		path := getPackagePath(importName)

		res = append(res, path)
	}

	return res, nil
}

func loadPackageInfo(file string) ([]string, error) {
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

func getPackagePath(importPath string) string {
	gopath := os.Getenv("GOPATH")
	goSumPath := filepath.Join(gopath, "pkg", "mod", importPath, "go.sum")
	return goSumPath
}

// TODO: remove
func getGoFiles(dir string, ignore []string) ([]string, error) {
	files, err := fake.ListGoFiles(dir, ignore)
	if err != nil {
		return nil, fmt.Errorf("getFiles: %w", err)
	}

	return files, nil
}

//func groupByDirectory(files Hashes) map[string]Hashes { //V1
//	groups := make(map[string]Hashes)
//	for file, data := range files {
//		dir := filepath.Dir(file)
//		if _, ok := groups[dir]; !ok {
//			groups[dir] = make(Hashes)
//		}
//		groups[dir][file] = data
//	}
//	return groups
//}

// TODO: use group by dir
func saveHashToFile(root, dir string, hash map[string]FileHashData) error {

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
