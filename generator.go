package fake

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/hashCheck"
	"github.com/sonalys/fake/internal/files"
)

// Generator is the controller for mock generation, holding cache for the targeted module.
type Generator struct {
	FileSet *token.FileSet
	// Name of the generated mock package.
	MockPackageName string
}

// NewGenerator will create a new mock generator for the specified module.
func NewGenerator(n string) *Generator {
	return &Generator{
		FileSet:         token.NewFileSet(),
		MockPackageName: n,
	}
}

func (g *Generator) ParseFile(input string) (*ParsedFile, error) {
	file, err := parser.ParseFile(g.FileSet, input, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	pkgPath, _ := files.GetPackagePath(g.FileSet, input)
	return &ParsedFile{
		Generator:   g,
		Ref:         file,
		PkgPath:     pkgPath,
		PkgName:     file.Name.Name,
		Imports:     ParseImports(file.Imports),
		UsedImports: make(map[string]struct{}),
	}, nil
}

func generateOutputFile(input, output string) *os.File {
	filename, _ := strings.CutSuffix(path.Base(input), ".go")
	outputFile := path.Join(output, fmt.Sprintf("%s.gen.go", filename))
	outFile, err := files.CreateFileAndFolders(outputFile)
	if err != nil {
		log.Panic().Msgf("Error creating mock file: %v\n", err)
	}
	return outFile
}

func (g *Generator) WriteFile(input, output string) bool {
	parsedFile, err := g.ParseFile(input)
	if err != nil {
		log.Panic().Msgf("failed to parse file: %s", input)
	}
	body := bytes.NewBuffer(make([]byte, 0, 2048))
	header := bytes.NewBuffer(make([]byte, 0, 2048))

	interfaces := parsedFile.ListInterfaces()
	if len(interfaces) == 0 {
		return false
	}
	// Iterate through the declarations in the file
	for _, parsedInterface := range interfaces {
		for _, field := range parsedInterface.ListFields() {
			field.UpdateImports()
		}
		parsedInterface.WriteMock(body)
	}

	WriteHeader(header, g.MockPackageName)
	parsedFile.WriteImports(header)
	// Append body to the header.
	header.Write(body.Bytes())
	// Fetch file buffer.
	b := header.Bytes()
	// Run gofmt in the buffer.
	b = FormatCode(b)
	outputFile := generateOutputFile(input, output)
	outputFile.Write(b)
	outputFile.Close()
	return true
}

func WriteHeader(w io.Writer, packageName string) {
	fmt.Fprintf(w, "// Code generated by mockgen. DO NOT EDIT.\n\n")
	fmt.Fprintf(w, "package %s\n\n", packageName)
}

func FormatCode(in []byte) []byte {
	out, err := format.Source(in)
	if err != nil {
		log.Panic().Msgf("Error formatting file: %v\n", err)
	}
	return out
}

func Run(dirs []string, output string, ignore []string) {
	gen := NewGenerator("mocks")
	dirs, err := hashCheck.CompareFileHashes(dirs, append(ignore, output), output)
	if err != nil {
		log.Fatal().Err(err).Msg("Error comparing file hashes")
	}
	if len(dirs) == 0 {
		log.Info().Msgf("no files found, nothing to be done")
		return
	}
	log.Info().Msgf("scanning %d files", len(dirs))
	for _, dir := range dirs {
		pkg := path.Dir(dir)
		pkg = strings.ReplaceAll(pkg, "internal", "internal_")
		out := path.Join(output, pkg)
		gen.WriteFile(dir, out)
		log.Info().Msgf("digesting %s to %s", dir, out)
	}
}
