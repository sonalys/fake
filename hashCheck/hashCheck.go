package hashCheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
		files, err := fake.ListGoFiles(dir, ignore)
		if err != nil {
			return nil, fmt.Errorf("getFiles: %w", err)
		}

		groups := groupByDirectory(files)

		for dir, files := range groups {
			CheckDependenciesChanges(files[0])

			hashes, err := dirHash(files)
			if err != nil {
				return nil, fmt.Errorf("dirHash: %w", err)
			}

			if _, err := os.Stat(filepath.Join(dir, fileName)); !os.IsNotExist(err) {
				jsonHashes := make(map[string]string)
				tmp, err := os.ReadFile(filepath.Join(dir, fileName))

				if err != nil {
					return nil, err
				}

				err = json.Unmarshal(tmp, &jsonHashes)
				if err != nil {
					return nil, fmt.Errorf("json.Unmarshal: %w", err)
				}
				res = append(res, cmpHashes(jsonHashes, hashes)...)
			} else {
				res = append(res, files...)
			}

			err = saveHashesToFile(hashes, filepath.Join(dir, fileName))
			if err != nil {
				return nil, fmt.Errorf("saveHashesToFile: %w", err)
			}
		}

	}
	return res, nil
}

// groupByDirectory takes a list of absolute paths files and groups them by directory
func groupByDirectory(files []string) map[string][]string {
	groups := make(map[string][]string)
	for _, file := range files {
		dir := filepath.Dir(file)
		groups[dir] = append(groups[dir], file)
	}
	return groups
}

// dirHash generates a hash for each file returning a map of fileName : hash
func dirHash(files []string) (map[string]string, error) {
	hashes := make(map[string]string)
	for _, file := range files {
		hash, err := hashFile(file)
		if err != nil {
			return nil, err
		}
		hashes[file] = hash
	}

	return hashes, nil
}

// saveHashesToFile saves the hashes to a "fake.lock.json" file
func saveHashesToFile(hashes map[string]string, filePath string) error {
	data, err := json.Marshal(hashes)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// hashFile returns the SHA256 hash of a single file
func hashFile(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// cmpHashes compares the hashes from the json file with the calculated hashes
// returns a list of files that have changed
func cmpHashes(jsonHashes, calcHashes map[string]string) []string {
	res := make([]string, 0)

	for k, v := range calcHashes {
		if jsonHashes[k] != v {
			res = append(res, k)
		}
	}

	return res
}

// CheckDependenciesChanges проверяет, меняются ли зависимости для файла .go
func CheckDependenciesChanges(file string) (bool, error) {
	// Загрузить информацию о пакете до изменения файла
	before, err := loadPackageInfo(file)
	if err != nil {
		return false, err
	}

	fmt.Printf("file: %s, before: %v\n", file, before.Imports)
	// Сравнить списки зависимостей
	return true, nil
}

// loadPackageInfo загружает информацию о пакете для данного файла
func loadPackageInfo(file string) (*packages.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedImports,
	}
	pkgs, err := packages.Load(cfg, file)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("encountered errors while loading package information")
	}
	return pkgs[0], nil
}
