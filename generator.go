package fake

import (
	"go/token"
)

// Generator is the controller for the whole module, caching files and holding metadata.
type Generator struct {
	FileSet         *token.FileSet
	MockPackageName string
}

// NewGenerator will create a new mock generator for the specified module.
func NewGenerator(n string) *Generator {
	return &Generator{
		FileSet:         token.NewFileSet(),
		MockPackageName: n,
	}
}
