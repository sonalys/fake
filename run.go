package fake

import (
	"fmt"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/caching"
	"github.com/sonalys/fake/internal/files"
)

type GenerateInterfaceConfig struct {
	PackageName   string
	Inputs        []string
	InterfaceName string
	OutputFolder  string
}

func GenerateInterface(c GenerateInterfaceConfig) {
	fileHashes, err := caching.GetUncachedFiles(c.Inputs, nil, "")
	if err != nil {
		log.Fatal().Err(err).Msg("error comparing file hashes")
	}
	log.Info().Msgf("scanning %d files", len(fileHashes))
	gen := NewGenerator(c.PackageName)
	for curFilePath := range fileHashes {
		b := gen.GenerateFile(curFilePath, c.InterfaceName)
		if b == nil {
			continue
		}
		log.Info().Msgf("generating mock for %s:%s", curFilePath, c.InterfaceName)
		oldFilename := strings.TrimRight(path.Base(curFilePath), path.Ext(curFilePath))
		filename := fmt.Sprintf("%s.%s.gen.go", oldFilename, c.InterfaceName)
		outputFilename := path.Join(c.OutputFolder, filename)
		outputFile, err := files.CreateFileAndFolders(outputFilename)
		if err != nil {
			log.Fatal().Err(err).Msgf("error opening file %s", outputFilename)
		}
		outputFile.Write(b)
		outputFile.Close()
	}
}

func Run(inputs []string, output string, ignore []string, interfaces ...string) {
	gen := NewGenerator("mocks")
	fileHashes, err := caching.GetUncachedFiles(inputs, append(ignore, output), output)
	if err != nil {
		log.Fatal().Err(err).Msg("error comparing file hashes")
	}
	var counter int
	for curFilePath, lockFile := range fileHashes {
		if !lockFile.Changed() {
			continue
		}
		if b := gen.GenerateFile(curFilePath); len(b) > 0 {
			log.Info().Msgf("generating mock for %s", curFilePath)
			counter++
			outputFile := openOutputFile(curFilePath, output)
			outputFile.Write(b)
			outputFile.Close()
		} else {
			// Remove empty files from our new lock file.
			delete(fileHashes, curFilePath)
		}
	}
	if len(fileHashes) == 0 || counter == 0 {
		log.Info().Msgf("nothing to be done")
		return
	}
	if err := caching.WriteLockFile(output, fileHashes); err != nil {
		log.Error().Err(err).Msg("error saving lock file")
	}
}
