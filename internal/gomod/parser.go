package gomod

import (
	"fmt"
	"io"
	"os"

	"github.com/sonalys/fake/internal/files"
	"golang.org/x/mod/modfile"
)

func Parse(dir string) (map[string]string, error) {
	goModPath, err := files.FindFile(dir, "go.mod")
	if err != nil {
		return nil, fmt.Errorf("could not find go.mod: %w", err)
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
