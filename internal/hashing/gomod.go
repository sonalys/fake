package hashing

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sonalys/fake/internal/files"
	"golang.org/x/mod/modfile"
)

// readGoSum reads and parses the go.sum file into a map
// import : []string{all related hashes}
func readGoSum(path string, goMod map[string]string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	dependencies := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}
		if _, ok := goMod[parts[0]]; ok {
			dependencies[parts[0]] = parts[2]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return dependencies, nil
}

func readGoMod(dir string) (map[string]string, error) {
	goModPath, err := files.FindFile(dir, "go.mod")
	if err != nil {
		return nil, fmt.Errorf("could not find go.sum: %w", err)
	}
	f, err := os.Open(goModPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	parsedGoMod, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return nil, err
	}
	dependencies := make(map[string]string, len(parsedGoMod.Require))
	for _, require := range parsedGoMod.Require {
		dependencies[require.Mod.Path] = require.Mod.Version
	}
	return dependencies, nil
}

func parseGoSum(dir string) (map[string]string, error) {
	goMod, err := readGoMod(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read go.mod: %w", err)
	}
	goSumPath, err := files.FindFile(dir, "go.sum")
	if err != nil {
		return nil, fmt.Errorf("could not find go.sum: %w", err)
	}
	dependenciesParsed, err := readGoSum(goSumPath, goMod)
	if err != nil {
		return nil, fmt.Errorf("could not read go.sum: %w", err)
	}
	return dependenciesParsed, nil
}
