package hashing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type (
	FileLockData struct {
		Hash  string `json:"hash,omitempty"`
		GoSum string `json:"gosum,omitempty"`
	}

	Hashes map[string]FileLockData
)

// readLockFile reads and parses the json model from the fake.lock.json file
// parses file from mocks/{path}/fake.lock.json
func readLockFile(path string) (Hashes, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Hashes{}, nil
		}
		return nil, err
	}
	var model Hashes
	err = json.Unmarshal(data, &model)
	return model, err
}

/*
saveLockFiles function takes dir string
and the target directory (dir), as well as a hash map (hash).
It saves file at path output/{dir}/fake.lock.json
*/
func saveLockFiles(dir, output string, hash map[string]FileLockData) error {
	data, err := json.MarshalIndent(hash, "", "\t")
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(output, dir), os.ModePerm)
	if err != nil {
		return err
	}
	w, err := os.Create(filepath.Join(output, dir, lockFilename))
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(data)
	return err
}
