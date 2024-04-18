package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/sonalys/fake/internal/files"
	"golang.org/x/tools/go/packages"
)

const (
	lockFilename = "fake.lock.json"
)

func getImportsHash(filePath string, dependencies map[string]string) (string, error) {
	imports, err := loadPackageImports(filePath)
	if err != nil {
		return "", fmt.Errorf("loading package imports: %w", err)
	}
	sort.Strings(imports)
	var b strings.Builder
	for _, importPath := range imports {
		if hash, ok := dependencies[importPath]; ok {
			b.WriteString(hash)
		}
	}
	return b.String(), nil
}

func GetUncachedFiles(inputDirs, ignore []string, outputDir string) (map[string]map[string]LockfileHandler, error) {
	dependencies, err := parseGoSum(inputDirs[0])
	if err != nil {
		return nil, fmt.Errorf("parsing go.sum file: %w", err)
	}
	goFiles, err := files.ListGoFiles(inputDirs, ignore)
	if err != nil {
		return nil, fmt.Errorf("listing *.go files: %w", err)
	}
	lockFiles := make(map[string]map[string]LockfileHandler, len(goFiles))
	for dir, filePathList := range files.GroupByDirectory(goFiles) {
		if _, ok := lockFiles[dir]; !ok {
			lockFiles[dir] = make(map[string]LockfileHandler)
		}
		lockFilePath := path.Join(outputDir, dir, lockFilename)
		lockFilePath = strings.ReplaceAll(lockFilePath, "internal", "internal_")
		groupLockFiles, err := readLockFile(lockFilePath)
		if err != nil {
			return nil, fmt.Errorf("reading .fake.lock.json file: %w", err)
		}
		for _, filePath := range filePathList {
			baseFilePath := path.Base(filePath)
			entry, ok := groupLockFiles[baseFilePath]
			if !ok {
				lockFiles[dir][baseFilePath] = &UnhashedLockFile{
					Filepath:     filePath,
					Dependencies: dependencies,
				}
				continue
			}
			importsHash, err := getImportsHash(filePath, dependencies)
			if err != nil {
				return nil, err
			}
			hash, err := hashFiles(filePath)
			if err != nil {
				return nil, fmt.Errorf("hashing file: %w", err)
			}
			if entry.Hash == hash && entry.Dependencies == importsHash {
				continue
			}
			lockFiles[dir][baseFilePath] = &HashedLockFile{
				Hash:         hash,
				Dependencies: importsHash,
				changed:      true,
			}
		}
	}
	return lockFiles, nil
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
	var imports []string
	for _, pkg := range pkgs {
		for imp := range pkg.Imports {
			imports = append(imports, imp)
		}
	}
	return imports, nil
}

// hashFiles returns the SHA256 hash of files
func hashFiles(files ...string) (string, error) {
	var hasher = sha256.New()
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
