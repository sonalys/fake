package files

import (
	"fmt"
	"path"
	"strings"
)

func GenerateOutputFileName(input, output string) string {
	filename, _ := strings.CutSuffix(path.Base(input), ".go")
	return path.Join(output, fmt.Sprintf("%s.gen.go", filename))
}
