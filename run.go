package fake

import (
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/caching"
)

func Run(dirs []string, output string, ignore []string) {
	gen := NewGenerator("mocks")
	fileHashes, err := caching.GetUncachedFiles(dirs, append(ignore, output), output)
	if err != nil {
		log.Fatal().Err(err).Msg("error comparing file hashes")
	}
	var counter int
	for curFilePath, lockFile := range fileHashes {
		outDir := path.Join(output, path.Dir(curFilePath))
		outDir = strings.ReplaceAll(outDir, "internal", "internal_")
		if !lockFile.Changed() {
			continue
		}
		if gen.WriteFile(curFilePath, outDir) {
			counter++
		} else {
			// Remove empty files from our new lock file.
			delete(fileHashes, curFilePath)
		}
	}
	// if err := caching.WriteLockFile(output, fileHashes); err != nil {
	// 	log.Error().Err(err).Msg("error saving lock file")
	// }
	if counter == 0 {
		log.Info().Msgf("nothing to be done")
		return
	}
}
