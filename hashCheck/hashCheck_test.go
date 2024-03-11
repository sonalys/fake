package hashCheck

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sliceToMap(slice []string) map[string]bool {
	result := make(map[string]bool)
	for _, item := range slice {
		result[item] = true
	}
	return result
}

func TestGroupByDirectory(t *testing.T) {
	t.Run("Should group files by directory", func(t *testing.T) {
		inputs := []string{
			"/home/user/documents/file1.txt",
			"/home/user/documents/file2.txt",
			"/home/user/images/image1.png",
			"/home/user/images/image2.png",
			"/home/user/images/image3.png",
		}

		expectedOutput := map[string][]string{
			"/home/user/documents": {
				"/home/user/documents/file1.txt",
				"/home/user/documents/file2.txt",
			},
			"/home/user/images": {
				"/home/user/images/image1.png",
				"/home/user/images/image2.png",
				"/home/user/images/image3.png",
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
			"/file1.txt",
			"/file2.txt",
		}

		expectedOutputs := map[string][]string{
			"/": {
				"/file1.txt",
				"/file2.txt",
			},
		}

		actualOutputs := groupByDirectory(inputs)

		if !reflect.DeepEqual(actualOutputs, expectedOutputs) {
			t.Errorf("Expected %v but got %v", expectedOutputs, actualOutputs)
		}
	})
}

func TestParseJsonModel(t *testing.T) {
	t.Run("Should return json model", func(t *testing.T) {
		input := "/json/valid"
		expectedOutput := Hashes{
			"repository/repository.go": {
				Hash:  "94d34c809fe95f47f6157330958e673cb19922c11091af2d3259e795a47914cc",
				GoSum: "87e39adc6f6fbe5f5b49ac3e7ce96b1193fe571760097e40f02768d43601b141",
			},
		}

		actualOutput, _ := parseJsonModel(input)

		if diff := cmp.Diff(actualOutput, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}

	})

	t.Run("Should return error because json file is not valid", func(t *testing.T) {
		input := "json/invalid"
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

func TestSaveHashToFile(t *testing.T) {
	t.Run("Should save JSON model to file in existing directory", func(t *testing.T) {
		root := "."
		dir := "output"
		hash := Hashes{
			"test.go": {
				Hash:  "testhash",
				GoSum: "testgosum",
			},
		}

		err := saveHashToFile(root, dir, hash)

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
		root := "."
		dir := "doesntexists"
		hash := Hashes{
			"test.go": {
				Hash:  "testhash",
				GoSum: "testgosum",
			},
		}

		err := saveHashToFile(root, dir, hash)

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

		os.RemoveAll(filepath.Join("./mocks", dir))

	})
}

func TestLoadPackageImports(t *testing.T) {
	t.Run("Should return path to zerolog/log go.sum fiel", func(t *testing.T) {
		input := "test_data/do_not_change/log.go"
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
		input := "test_data/do_not_change/no_imports.go"

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

func TestGetPackagePath(t *testing.T) {
	t.Run("Should return path to zerolog/log go.sum file", func(t *testing.T) {
		input := "github.com/rs/zerolog"
		expectedOutput := fmt.Sprintf(os.Getenv("GOPATH") + "/pkg/mod/github.com/rs/zerolog@v1.32.0/go.sum")

		output, err := getPackagePath(input)

		if err != nil {
			t.Errorf("Expected nil but got %v", err)
		}

		if diff := cmp.Diff(output, expectedOutput); diff != "" {
			t.Errorf("(-got +want)\n%s", diff)
		}
	})
}
