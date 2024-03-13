package fake

import (
	"os"
	"path/filepath"
	"strings"
)

// ListGoFiles lists all Go files under a directory.
func ListGoFiles(input, ignore []string) ([]string, error) {
	var goFiles []string
	for _, filename := range input {
		err := filepath.Walk(filename, func(filename string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			for _, entry := range ignore {
				if _, ok := strings.CutPrefix(filename, entry); ok {
					return nil
				}
			}
			// Check if the file has a ".go" extension
			if strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go") {
				goFiles = append(goFiles, filename)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return goFiles, nil
}
