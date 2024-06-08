package imports

import (
	"go/ast"
	"strings"

	"github.com/sonalys/fake/internal/packages"
)

type (
	ImportEntry struct {
		*packages.PackageInfo
		Alias string
	}
)

func CachedImportInformation(dir string) func(f *ast.File) (nameMap, pathMap map[string]*ImportEntry) {
	cache := make(map[string]*packages.PackageInfo, 100)
	return func(f *ast.File) (nameMap map[string]*ImportEntry, pathMap map[string]*ImportEntry) {
		nameMap = make(map[string]*ImportEntry, len(f.Imports))
		pathMap = make(map[string]*ImportEntry, len(f.Imports))

		for _, i := range f.Imports {
			trimmedPath := strings.Trim(i.Path.Value, "\"")
			var info *packages.PackageInfo
			if cachedInfo, ok := cache[trimmedPath]; ok {
				info = cachedInfo
			} else {
				info, ok = packages.Parse(dir, trimmedPath)
				if !ok {
					continue
				}
			}
			var importEntry = &ImportEntry{
				PackageInfo: info,
				Alias:       "",
			}
			usedName := info.Name
			if i.Name != nil && i.Name.Name != "" && info.Name != i.Name.Name {
				usedName = i.Name.Name
				importEntry.Alias = usedName
			}
			nameMap[usedName] = importEntry
			pathMap[info.Path] = importEntry
		}
		return nameMap, pathMap
	}
}
