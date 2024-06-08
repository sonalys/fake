package caching

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

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/files"
	"github.com/sonalys/fake/internal/gosum"
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

func GetUncachedFiles(inputs, ignore []string, outputDir string) (map[string]LockfileHandler, error) {
	lockFilePath := path.Join(outputDir, lockFilename)
	lockFilePath = strings.ReplaceAll(lockFilePath, "internal", "internal_")
	groupLockFiles, err := readLockFile(lockFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading %s file: %w", lockFilename, err)
	}
	var dependencies map[string]string
	if len(groupLockFiles) > 0 {
		dependencies, err = gosum.Parse(inputs[0])
		if err != nil {
			return nil, fmt.Errorf("parsing go.sum file: %w", err)
		}
	}
	goFiles, err := files.ListGoFiles(inputs, append(ignore, outputDir))
	if err != nil {
		return nil, fmt.Errorf("listing *.go files: %w", err)
	}
	out := make(map[string]LockfileHandler, len(groupLockFiles))

	cachedHasher := getFileHasher(len(goFiles))
	// TODO: split into a function.
	for _, filePathList := range files.GroupByDirectory(goFiles) {
		for _, filePath := range filePathList {
			entry, ok := groupLockFiles[filePath]
			// If file is not in lock file hashes, then we delay hash calculation for after the mock generation.
			// this makes it faster by avoiding calculation of useless files.
			if !ok {
				out[filePath] = &UnhashedLockFile{
					Filepath:     filePath,
					Dependencies: dependencies,
				}
				continue
			}
			stat, _ := os.Stat(filePath)
			if !entry.ModifiedAt.IsZero() && !stat.ModTime().IsZero() && stat.ModTime().Equal(entry.ModifiedAt) {
				entry.exists = true
				out[filePath] = &entry
				continue
			}
			importsHash, err := getImportsHash(filePath, dependencies)
			if err != nil {
				return nil, err
			}
			hash, err := cachedHasher(filePath)
			if err != nil {
				return nil, fmt.Errorf("hashing file: %w", err)
			}
			if entry.Hash == hash && entry.Dependencies == importsHash {
				// Mark file as processed, to further delete unused entries.
				entry.exists = true
				out[filePath] = &entry
				continue
			}
			out[filePath] = &HashedLockFile{
				changed:      true,
				exists:       true,
				Hash:         hash,
				Dependencies: importsHash,
				ModifiedAt:   stat.ModTime(),
			}
		}
	}
	for filePath := range groupLockFiles {
		if _, ok := out[filePath]; !ok {
			// Remove empty files from our new lock file.
			rmFileName := files.GenerateOutputFileName(filePath, outputDir)
			os.Remove(rmFileName)
			log.Info().Msgf("removing legacy mock from %s", rmFileName)
		}
	}
	return out, nil
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

func getFileHasher(cacheSize int) func(...string) (string, error) {
	cache := make(map[string]string, cacheSize)
	return func(files ...string) (string, error) {
		var hasher = sha256.New()
		for _, file := range files {
			if hash, hit := cache[file]; hit {
				hasher.Write([]byte(hash))
				continue
			}
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
