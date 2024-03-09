package hashCheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sonalys/fake"
	"go/build"
	"go/parser"
	"go/token"
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

		model, err := parseJsonModel(dir)
		if err != nil {
			return nil, fmt.Errorf("parseJsonModel: %w", err)
		}

		for _, file := range files {

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

		//fmt.Printf("model: %v\n", model)
		fmt.Printf("dir: %s\n", dir)
		err = saveHashToFile(dir, model)
		if err != nil {
			return nil, fmt.Errorf("saveHashToFile: %w", err)
		}

	}

	return res, nil
}

func parseJsonModel(path string) (Hashes, error) {
	data, err := os.ReadFile(filepath.Join(path, fileName))
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
	fmt.Printf("imports: %v\n", imports)
	if err != nil {
		return nil, fmt.Errorf("loadPackageInfo: %w", err)
	}

	for _, importName := range imports {
		path, err := getPackagePath(importName)
		fmt.Printf("path: %s\n", path)

		if err != nil {
			return nil, fmt.Errorf("getPackagePath: %w", err)
		}
		goSumPath := filepath.Join(path, "go.sum")

		res = append(res, goSumPath)
	}

	return res, nil
}

func loadPackageInfo(file string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	imports := make([]string, len(f.Imports))
	for i, imp := range f.Imports {
		imports[i] = imp.Path.Value
	}

	return imports, nil
}

func getPackagePath(importPath string) (string, error) {
	pkg, err := build.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", err
	}

	return pkg.Dir, nil
}

// TODO: remove
func getGoFiles(dir string, ignore []string) ([]string, error) {
	files, err := fake.ListGoFiles(dir, ignore)
	if err != nil {
		return nil, fmt.Errorf("getFiles: %w", err)
	}

	return files, nil
}

// TODO: use group by dir
func saveHashToFile(dir string, hash map[string]FileHashData) error {
	data, err := json.Marshal(hash)
	if err != nil {
		return err
	}

	fmt.Printf("hash: %v\n", hash)
	err = os.MkdirAll(filepath.Join("mocks", dir), os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join("mocks", dir, fileName), data, 0644)
}

// hashFiles returns the SHA256 hash of files
func hashFiles(files ...string) (string, error) {
	hasher := sha256.New()

	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				//fmt.Printf("File %s does not exist\n", file)
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
