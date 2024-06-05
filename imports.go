package fake

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/packages"
)

type PackageInfo struct {
	Ref   *ast.ImportSpec
	Alias string
	Name  string
	Path  string
}

// ParsePackageInfo parses the specified package and returns its package name and import path.
func ParsePackageInfo(importPath string) (*PackageInfo, bool) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
	}
	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, false
	}
	if len(pkgs) == 0 {
		return nil, false
	}
	// Assuming the first package is the main package, you might need to adjust this logic.
	mainPkg := pkgs[0]
	return &PackageInfo{
		Name: mainPkg.Name,
		Path: mainPkg.PkgPath,
	}, true
}

func ParseImports(imports []*ast.ImportSpec) *map[string]*PackageInfo {
	importNamePathMap := make(map[string]*PackageInfo)
	for _, i := range imports {
		trimmedPath := strings.Trim(i.Path.Value, "\"")
		info, ok := ParsePackageInfo(trimmedPath)
		if !ok {
			continue
		}
		info.Ref = i
		name := info.Name
		if i.Name != nil {
			name = i.Name.Name
			info.Alias = name
		}
		importNamePathMap[name] = info
	}
	return &importNamePathMap
}
