package gosum

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sonalys/fake/internal/files"
	"github.com/sonalys/fake/internal/gomod"
)

// readGoSum reads and parses the go.sum file into a map
// import : []string{all related hashes}
func readGoSum(path string, goMod map[string]string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
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

func Parse(dir string) (map[string]string, error) {
	goMod, err := gomod.Parse(dir)
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
