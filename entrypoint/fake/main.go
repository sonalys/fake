package main

import (
	"flag"
	"fmt"
	"os"

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
	flag.Var(&input, "input", "Folder to scan for .go files recursively")
	output := flag.String("output", "mocks", "Folder to output the generated mocks")
	flag.Var(&ignore, "ignore", "Specify which folders should be ignored")
	flag.Parse()
	if len(input) == 0 {
		// Defaults to $CWD
		input = []string{"."}
	}
	mockgen.Run(input, *output, ignore)
}
