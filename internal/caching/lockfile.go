package caching

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/files"
)

type (
	UnhashedLockFile struct {
		Filepath     string
		Dependencies map[string]string
	}

	HashedLockFile struct {
		ModifiedAt   time.Time `json:"modifiedAt,omitempty"`
		Hash         string    `json:"hash"`
		Dependencies string    `json:"dependencies,omitempty"`
		// Changed is used as an in-memory flag to say that a file lock changed.
		filepath string `json:"-"`
		changed  bool   `json:"-"`
		exists   bool   `json:"-"`
	}

	LockFilePackage map[string]HashedLockFile
)

type LockfileHandler interface {
	Changed() bool
	AbsolutePath() string
	Exists() bool
	Compute() *HashedLockFile
}

func (f *UnhashedLockFile) Changed() bool {
	return true
}

func (f *UnhashedLockFile) AbsolutePath() string {
	return f.Filepath
}

func (f *HashedLockFile) AbsolutePath() string {
	return f.filepath
}

func (f *UnhashedLockFile) Exists() bool {
	return true
}

func (f *UnhashedLockFile) Compute() *HashedLockFile {
	hash, err := hashFiles(f.Filepath)
	if err != nil {
		log.Error().Err(err).Msg("could not compute file hash")
	}
	dep, err := getImportsHash(f.Filepath, f.Dependencies)
	if err != nil {
		log.Error().Err(err).Msg("could not compute file imports hash")
	}
	return &HashedLockFile{
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

func (f *HashedLockFile) Compute() *HashedLockFile {
	return f
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
	var out = make(map[string]*HashedLockFile, len(hash))
	for file, entry := range hash {
		if entry.Exists() {
			out[file] = entry.Compute()
		}
	}
	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		return err
	}
	w, err := files.CreateFileAndFolders(filepath.Join(output, lockFilename))
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(data)
	return err
}
