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
	log.Info().Msgf("scanning %d files for interface %s", len(fileHashes), c.InterfaceName)
	gen, err := NewGenerator(c.PackageName, c.Inputs[0])
	if err != nil {
		log.Fatal().Err(err).Msg("error creating mock generator")
	}
	for relPath, hash := range fileHashes {
		b := gen.GenerateFile(hash.AbsolutePath(), c.InterfaceName)
		if b == nil {
			continue
		}
		log.Info().Msgf("generating mock for %s:%s", relPath, c.InterfaceName)
		oldFilename := strings.TrimRight(path.Base(relPath), path.Ext(relPath))
		filename := fmt.Sprintf("%s.%s.gen.go", oldFilename, c.InterfaceName)
		outputFilename := path.Join(c.OutputFolder, filename)
		outputFile, err := files.CreateFileAndFolders(outputFilename)
		if err != nil {
			log.Fatal().Err(err).Msgf("error opening file %s", outputFilename)
		}
		outputFile.Write(b)
		outputFile.Close()
	}
	if err := caching.WriteLockFile(path.Dir(gen.goModFilename), fileHashes); err != nil {
		log.Error().Err(err).Msg("error saving lock file")
	}
}

func Run(inputs []string, output string, ignore []string, interfaces ...string) {
	gen, err := NewGenerator("mocks", inputs[0])
	if err != nil {
		log.Fatal().Err(err).Msg("error creating mock generator")
	}
	fileHashes, err := caching.GetUncachedFiles(inputs, append(ignore, output), output)
	if err != nil {
		log.Fatal().Err(err).Msg("error comparing file hashes")
	}
	var counter int
	for relPath, lockFile := range fileHashes {
		if !lockFile.Changed() {
			continue
		}
		if b := gen.GenerateFile(lockFile.AbsolutePath()); len(b) > 0 {
			log.Info().Msgf("generating mock for %s", relPath)
			counter++
			outputFile := openOutputFile(relPath, output)
			outputFile.Write(b)
			outputFile.Close()
		} else {
			// Remove empty files from our new lock file.
			// delete(fileHashes, relPath)
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
