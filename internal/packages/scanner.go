package packages

import "golang.org/x/tools/go/packages"

type PackageInfo struct {
	Name  string
	Path  string
	Files []string
}

// Parse parses the specified package and returns its package name and import path.
func Parse(importPath string) (*PackageInfo, bool) {
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
		Name:  mainPkg.Name,
		Path:  mainPkg.PkgPath,
		Files: mainPkg.GoFiles,
	}, true
}
