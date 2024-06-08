package fake

import (
	"go/parser"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/files"
)

func (g *Generator) ParseFile(input string) (*ParsedFile, error) {
	file, err := parser.ParseFile(g.FileSet, input, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	packagePath, err := files.GetPackagePath(g.goModFilename, g.goMod, input)
	if err != nil {
		log.Error().Err(err).Msg("failed to get package info")
		return nil, err
	}
	imports, importsPathMap := g.cachedPackageInfo(file)
	return &ParsedFile{
		Generator:      g,
		Ref:            file,
		Size:           int(file.End()),
		PkgPath:        packagePath,
		PkgName:        file.Name.Name,
		Imports:        imports,
		ImportsPathMap: importsPathMap,
		UsedImports:    make(map[string]struct{}),
	}, nil
}
