package hashing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type (
	UnhashedLockFile struct {
		Filepath     string
		Dependencies map[string]string
	}

	HashedLockFile struct {
		Hash         string `json:"hash"`
		Dependencies string `json:"dependencies,omitempty"`
		// Changed is used as an in-memory flag to say that a file lock changed.
		changed bool `json:"-"`
		exists  bool `json:"-"`
	}

	LockFilePackage map[string]HashedLockFile
)

type LockfileHandler interface {
	Changed() bool
	Exists() bool
	Compute() HashedLockFile
}

func (f *UnhashedLockFile) Changed() bool {
	return true
}

func (f *UnhashedLockFile) Exists() bool {
	return true
}

func (f *UnhashedLockFile) Compute() HashedLockFile {
	hash, _ := hashFiles(f.Filepath)
	dep, _ := getImportsHash(f.Filepath, f.Dependencies)
	return HashedLockFile{
		Hash:         hash,
		Dependencies: dep,
	}
}

func (f *HashedLockFile) Changed() bool {
	return f.changed
}

func (f *HashedLockFile) Exists() bool {
	return f.exists
}

func (f *HashedLockFile) Compute() HashedLockFile {
	return *f
}

// readLockFile reads and parses the json model from the fake.lock.json file
// parses file from mocks/{path}/fake.lock.json
func readLockFile(path string) (LockFilePackage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var model LockFilePackage
	err = json.Unmarshal(data, &model)
	return model, err
}

/*
WriteLockFile function takes dir string
and the target directory (dir), as well as a hash map (hash).
It saves file at path output/{dir}/fake.lock.json
*/
func WriteLockFile(output string, hash map[string]LockfileHandler) error {
	var out = make(map[string]HashedLockFile, len(hash))
	for file, entry := range hash {
		if entry.Exists() {
			out[file] = entry.Compute()
		}
	}
	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		return err
	}
	w, err := os.Create(filepath.Join(output, lockFilename))
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(data)
	return err
}
