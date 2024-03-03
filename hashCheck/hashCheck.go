package hashCheck

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	fileName = "fake.lock.json"
)

// CompareFileHashes identifies changed files in directories by comparing current to stored hashes.
// Returns a list of changed files and any errors encountered.
func CompareFileHashes(inputDirs []string) ([]string, error) {
	res := make([]string, 0)

	for _, dir := range inputDirs {
		files, err := getFiles(dir)
		if err != nil {
			return nil, fmt.Errorf("getFiles: %w", err)
		}

		groups := groupByDirectory(files)

		for dir, files := range groups {
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

// getFiles returns a list of all .go files and fake.lock.json in a directory with absolute paths
func getFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") && !strings.HasSuffix(info.Name(), ".gen.go")) && !slices.Contains(files, path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
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
