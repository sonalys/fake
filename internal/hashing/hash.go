package hashing

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/sonalys/fake/internal/files"
	"golang.org/x/tools/go/packages"
)

const (
	lockFilename = "fake.lock.json"
)

func GetUpdatedFiles(inputDirs, ignore []string, outputDir string) ([]string, error) {
	res := make([]string, 0, len(inputDirs))
	dependencies, err := parseGoSum(inputDirs[0])
	if err != nil {
		return nil, fmt.Errorf("parsing go.sum file: %w", err)
	}
	goFiles, err := files.ListGoFiles(inputDirs, ignore)
	if err != nil {
		return nil, fmt.Errorf("getFiles: %w", err)
	}
	groups := files.GroupByDirectory(goFiles)
	for group, data := range groups {
		lockFile, err := readLockFile(path.Join(outputDir, group, lockFilename))
		if err != nil {
			return nil, fmt.Errorf("parseJsonModel: %w", err)
		}
		for _, file := range data {
			imports, err := loadPackageImports(file)
			importHash := make([]string, 0, len(imports))
			if err != nil {
				return nil, fmt.Errorf("loadPackageImports: %w", err)
			}
			for _, importName := range imports {
				if importPath, ok := dependencies[importName]; ok {
					importHash = append(importHash, importPath)
				}
			}
			goSum, err := hashFiles(importHash...)
			if err != nil {
				return nil, fmt.Errorf("hashFiles: %w", err)
			}
			hash, err := hashFiles(file)
			if err != nil {
				return nil, fmt.Errorf("hashFiles: %w", err)
			}
			if lockFile[file].Hash == hash && lockFile[file].GoSum == goSum {
				continue
			}
			res = append(res, file)
			lockFile[file] = FileLockData{
				Hash:  hash,
				GoSum: goSum,
			}
		}
		if err = saveLockFiles(group, outputDir, lockFile); err != nil {
			return nil, fmt.Errorf("saveHashToFile: %w", err)
		}
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
		return nil, errors.New("loaded packages contain errors")
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
