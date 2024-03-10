package fake

import (
	"go/token"
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

func (g *Generator) WriteFile(input, output string) bool {
	parsedFile := NewParsedFile(g, input)
	interfaces := parsedFile.ListInterfaces()
	if len(interfaces) == 0 {
		return false
	}
	if err := parsedFile.Write(output); err != nil {
		log.Fatal().Err(err).Msg("failed to generate file")
	}
	return true
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
