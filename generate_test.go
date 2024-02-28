package fake

import "testing"

func Test_Generate(t *testing.T) {
	Run([]string{"testdata"}, "testdata/out", []string{"testdata/out"})
}
