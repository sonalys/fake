package fake

import (
	"bytes"
	"fmt"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
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

func (g *Generator) ParseFile(input string) *ParsedFile {
	file, err := parser.ParseFile(g.FileSet, input, nil, parser.ParseComments)
	if err != nil {
		panic(fmt.Sprintf("Error parsing file: %v\n", err))
	}
	pkgPath, _ := GetPackagePath(g.FileSet, input)
	return &ParsedFile{
		Generator:   g,
		Ref:         file,
		PkgPath:     pkgPath,
		PkgName:     file.Name.Name,
		Imports:     ParseImports(file.Imports),
		UsedImports: make(map[string]struct{}),
	}
}

func (g *Generator) WriteFile(input, output string) bool {
	parsedFile := g.ParseFile(input)
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

func Run(inputs []string, output string, ignore []string) {
	gen := NewGenerator("mocks")
	var filenames []string
	for _, input := range inputs {
		files, err := ListGoFiles(input, append(ignore, output))
		if err != nil {
			log.Fatal().Msgf("error listing files: %s", err)
		}
		filenames = append(filenames, files...)
	}
	if len(filenames) == 0 {
		log.Info().Msgf("no files found, nothing to be done")
		return
	}
	log.Info().Msgf("scanning %d files", len(filenames))
	for _, filename := range filenames {
		pkg := path.Dir(filename)
		pkg = strings.ReplaceAll(pkg, "internal", "internal_")
		out := path.Join(output, pkg)
		gen.WriteFile(filename, out)
		log.Info().Msgf("digesting %s to %s", filename, out)
	}
}
