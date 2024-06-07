package imports

import (
	"go/ast"
	"strings"

	"github.com/sonalys/fake/internal/packages"
)

type (
	ImportEntry struct {
		packages.PackageInfo
		Alias string
	}
)

func FileListUsedImports(f *ast.File) (nameMap, pathMap map[string]ImportEntry) {
	importNamePathMap := make(map[string]*ImportEntry, len(f.Imports))
	for _, i := range f.Imports {
		trimmedPath := strings.Trim(i.Path.Value, "\"")
		info, ok := packages.Parse(trimmedPath)
		if !ok {
			continue
		}
		var importEntry = &ImportEntry{
			PackageInfo: *info,
			Alias:       "",
		}
		usedName := info.Name
		if i.Name != nil {
			usedName = i.Name.Name
			importEntry.Alias = usedName
		}
		importNamePathMap[usedName] = importEntry
	}
	// We want all imports used by interfaces.
	importChecker := getUsedInterfacePackages(f)
	nameMap = make(map[string]ImportEntry, len(importChecker))
	pathMap = make(map[string]ImportEntry, len(importChecker))
	for name := range importChecker {
		nameMap[name] = *importNamePathMap[name]
		pathMap[importNamePathMap[name].Path] = *importNamePathMap[name]
	}
	return nameMap, pathMap
}

// getUsedInterfacePackages returns all packages imported by interfaces.
func getUsedInterfacePackages(f *ast.File) map[string]struct{} {
	resp := make(map[string]struct{}, len(f.Imports))
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.InterfaceType); !ok {
				return true
			}
			ast.Inspect(x.Type, func(n ast.Node) bool {
				// We don't consider imports from interfaces, as they will also be traversed.
				sel, ok := n.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if ident, ok := sel.X.(*ast.Ident); ok {
					resp[ident.Name] = struct{}{}
				}
				return true
			})
		}
		return true
	})
	return resp
}
