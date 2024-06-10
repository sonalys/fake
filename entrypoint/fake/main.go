package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	mockgen "github.com/sonalys/fake"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
	})
}

type StrSlice []string

func (s *StrSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *StrSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var input, ignore StrSlice
	var interfaceName, pkgName *string
	flag.Var(&input, "input", "Folder to scan for .go files recursively")
	output := flag.String("output", "mocks", "Folder to output the generated mocks")
	flag.Var(&ignore, "ignore", "Specify which folders should be ignored")
	interfaceName = flag.String("interface", "", "If you want to generate a single interface on the same folder, specify using this flag")
	pkgName = flag.String("mockPackage", "", "Usable with -interface only. Provide if you want a different package from the interface being generated")
	flag.Parse()
	if len(input) == 0 {
		// Defaults to $CWD
		input = []string{"."}
	}
	for i := range input {
		absInput, err := filepath.Abs(input[i])
		if err == nil {
			input[i] = absInput
		}
	}
	if *interfaceName != "" {
		if *output != "mocks" {
			log.Error().Msgf("-output %s cannot be used when -interface is set", *output)
			return
		}
		mockgen.GenerateInterface(mockgen.GenerateInterfaceConfig{
			PackageName:   *pkgName,
			Inputs:        input,
			InterfaceName: *interfaceName,
			OutputFolder:  path.Dir(input[0]),
		})
		return
	}
	mockgen.Run(input, *output, ignore)
}
