package fake

import (
	"go/parser"

	"github.com/sonalys/fake/internal/files"
	"github.com/sonalys/fake/internal/imports"
)

func (g *Generator) ParseFile(input string) (*ParsedFile, error) {
	file, err := parser.ParseFile(g.FileSet, input, nil, parser.Mode(0))
	if err != nil {
		return nil, err
	}
	packagePath, _ := files.GetPackagePath(input)

	imports, importsPathMap := imports.FileListUsedImports(file)
	return &ParsedFile{
		Generator:      g,
		Ref:            file,
		Size:           int(file.End()),
		PkgPath:        packagePath,
		PkgName:        file.Name.Name,
		Imports:        imports,
		ImportsPathMap: importsPathMap,
	}, nil
}
