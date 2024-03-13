package hashCheck

import (
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"testing"
)

func sliceToMap(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}

func TestCompareFileHashes(t *testing.T) {
	t.Run("Should return files that are not matching hashes", func(t *testing.T) {
		inputDirs := []string{filepath.FromSlash("test_data/TestCompareFileHashes")}
		ignore := []string{filepath.FromSlash("test_data/TestCompareFileHashes/change_file/")}

		expectedOutput := []string{
			filepath.FromSlash("test_data/TestCompareFileHashes/stub.go"),
		}

		output, err := CompareFileHashes(inputDirs, ignore)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})

	t.Run("Should return empty slice when all files are matching hashes", func(t *testing.T) {
		inputDirs := []string{filepath.FromSlash("test_data/TestCompareFileHashes")}
		ignore := []string{filepath.FromSlash("test_data/TestCompareFileHashes/change_file/")}

		expectedOutput := make([]string, 0)

		output, err := CompareFileHashes(inputDirs, ignore)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		os.Remove(filepath.Join("mocks/test_data/TestCompareFileHashes/fake.lock.json"))

		if diff := cmp.Diff(sliceToMap(output), sliceToMap(expectedOutput)); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})

	t.Run("Should return whatever.go (if changed this file)", func(t *testing.T) {
		inputDirs := []string{filepath.FromSlash("test_data/TestCompareFileHashes/change_file")}
		ignore := []string{filepath.FromSlash("test_data/TestCompareFileHashes/stub.go")}

		expectedOutput := []string{filepath.FromSlash("test_data/TestCompareFileHashes/change_file/whatever.go")}

		output, err := CompareFileHashes(inputDirs, ignore)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(sliceToMap(output), sliceToMap(expectedOutput)); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})

}

func TestGroupByDirectory(t *testing.T) {
	t.Run("Should group files by directory", func(t *testing.T) {
		inputs := []string{
			filepath.FromSlash("/home/user/documents/file1.txt"),
			filepath.FromSlash("/home/user/documents/file2.txt"),
			filepath.FromSlash("/home/user/images/image1.png"),
			filepath.FromSlash("/home/user/images/image2.png"),
			filepath.FromSlash("/home/user/images/image3.png"),
		}

		expectedOutput := map[string][]string{
			filepath.FromSlash("/home/user/documents"): {
				filepath.FromSlash("/home/user/documents/file1.txt"),
				filepath.FromSlash("/home/user/documents/file2.txt"),
			},
			filepath.FromSlash("/home/user/images"): {
				filepath.FromSlash("/home/user/images/image1.png"),
				filepath.FromSlash("/home/user/images/image2.png"),
				filepath.FromSlash("/home/user/images/image3.png"),
			},
		}

		output := groupByDirectory(inputs)

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})

	t.Run("Should return empty map when no inputs", func(t *testing.T) {
		var inputs []string

		expectedOutput := map[string][]string{}

		output := groupByDirectory(inputs)

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})

	t.Run("Should handle files in root directory", func(t *testing.T) {
		inputs := []string{
			filepath.FromSlash("/home/user/documents/file1.txt"),
			filepath.FromSlash("/home/user/documents/file2.txt"),
			filepath.FromSlash("/home/user/images/image1.png"),
			filepath.FromSlash("/home/user/images/image2.png"),
			filepath.FromSlash("/home/user/images/image3.png"),
		}

		expectedOutput := map[string][]string{
			filepath.FromSlash("/home/user/documents"): {
				filepath.FromSlash("/home/user/documents/file1.txt"),
				filepath.FromSlash("/home/user/documents/file2.txt"),
			},
			filepath.FromSlash("/home/user/images"): {
				filepath.FromSlash("/home/user/images/image1.png"),
				filepath.FromSlash("/home/user/images/image2.png"),
				filepath.FromSlash("/home/user/images/image3.png"),
			},
		}

		actualOutputs := groupByDirectory(inputs)

		if diff := cmp.Diff(actualOutputs, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}

func TestParseJsonModel(t *testing.T) {
	t.Run("Should return json model", func(t *testing.T) {
		input := "TestParseJsonModel/valid"
		expectedOutput := Hashes{
			"repository/repository.go": {
				Hash:  "94d34c809fe95f47f6157330958e673cb19922c11091af2d3259e795a47914cc",
				GoSum: "87e39adc6f6fbe5f5b49ac3e7ce96b1193fe571760097e40f02768d43601b141",
			},
		}

		actualOutput, err := parseJsonModel(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(actualOutput, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})

	t.Run("Should return error because json file is not valid", func(t *testing.T) {
		input := "TestParseJsonModel/invalid"
		_, err := parseJsonModel(input)

		if err == nil {
			t.Errorf("Expected error but got nil")
		}
	})

	t.Run("Should return empty struct when file is not found", func(t *testing.T) {
		input := "notfound"
		expectedOutput := Hashes{}

		output, err := parseJsonModel(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})
}

func TestParseGoSumFile(t *testing.T) {
	t.Run("Should return go.sum file content", func(t *testing.T) {
		input := "test_data/TestParseGoSumFile/go.sum"

		expectedOutput := map[string][]string{"github.com/bsm/ginkgo/v2": {"h1:Ny8MWAHyOepLGlLKYmXG4IEkioBysk6GpaRTLC8zwWs="},
			"github.com/bsm/gomega":         {"h1:yeMWxP2pV2fG3FgAODIY8EiRE3dy0aeFYt4l7wh6yKA="},
			"github.com/bytedance/sonic":    {"h1:ED5hyg4y6t3/9Ku1R6dU/4KyJ48DZ4jPhfY1O2AihPM=", "h1:ElCzW+ufi8qKqNW0FY314xriJhyJhuoJ3gFZdAHF7NM=", "h1:GQebETVBxYB7JGWJtLBi07OVzWwt+8dWA00gEVW2ZFE=", "h1:iZcSUejdk5aukTND/Eu/ivjQuEL0Cu9/rf50Hi0u/g4="},
			"github.com/cespare/xxhash/v2":  {"h1:DC2CZ1Ep5Y4k3ZQ899DldepgrayRUGE6BBZ/cd9Cj44=", "h1:VGX0DQ3Q6kWi7AoAeZDth3/j3BFtOZR5XLFGgcrjCOs="},
			"github.com/chenzhuoyu/base64x": {"h1:DH46F32mSOjUmXrMHnKwZdA8wcEefY7UVqBKYGjpdQY=", "h1:b583jCggY9gE99b6G5LEC39OIiVsWj+R97kbl5odCEk="},
		}

		output, err := parseGoSumFile(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}

func TestLoadPackageImports(t *testing.T) {
	t.Run("Should return path to zerolog/log go.sum file", func(t *testing.T) {
		input := "test_data/TestLoadPackageImports/log.go"
		expectedOutput := []string{"fmt", "github.com/rs/zerolog/log", "time"}

		output, err := loadPackageImports(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		// cmp.Diff does not work with slices well, since it compares the order of the elements too

		if diff := cmp.Diff(sliceToMap(output), sliceToMap(expectedOutput)); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})

	t.Run("Should return error when pass non existing file", func(t *testing.T) {
		input := "whatever.go"

		_, err := loadPackageImports(input)

		if err == nil {
			t.Errorf("Expected error but got nil")
		}

	})

	t.Run("Should return nil when pass .go file with no imports", func(t *testing.T) {
		input := "test_data/TestLoadPackageImports/no_imports.go"

		var expectedOutput []string

		output, err := loadPackageImports(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)

		}

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})
}

func TestSaveHashToFile(t *testing.T) {
	t.Run("Should save JSON model to file in existing directory", func(t *testing.T) {
		dir := "TestSaveHashToFile"
		hash := Hashes{
			"test.go": {
				Hash:  "testhash",
				GoSum: "testgosum",
			},
		}

		err := saveHashToFile(dir, hash)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		output, err := parseJsonModel(dir)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, hash); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})

	t.Run("Should save JSON model to file in non existing directory", func(t *testing.T) {
		dir := "TestSaveHashToFile/nonexistingdir"
		hash := Hashes{
			"test.go": {
				Hash:  "testhash",
				GoSum: "testgosum",
			},
		}

		err := saveHashToFile(dir, hash)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		output, err := parseJsonModel(dir)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, hash); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

		os.RemoveAll(filepath.Join(dir, "mocks", dir))

	})
}

func TestHashFiles(t *testing.T) {
	t.Run("Should return hash of file", func(t *testing.T) {
		input := "test_data/TestHashFiles/log.go"
		expectedOutput := "d523b4893bb6fe9173057f4b56c029d029309d340825049131117bd609b65f5e"

		output, err := hashFiles(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if output != expectedOutput {
			t.Errorf("Expected %v but got %v", expectedOutput, output)
		}
	})
}
