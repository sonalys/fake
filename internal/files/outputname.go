package files

import (
	"fmt"
	"path"
	"strings"
)

func GenerateOutputFileName(input, output string) string {
	filename, _ := strings.CutSuffix(path.Base(input), ".go")
	subTree := strings.ReplaceAll(path.Dir(input), "internal", "internal_")
	return path.Join(output, subTree, fmt.Sprintf("%s.gen.go", filename))
}
