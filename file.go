package fake

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"slices"

	"github.com/sonalys/fake/internal/imports"
)

type ParsedFile struct {
	Generator       *Generator
	Size            int
	Ref             *ast.File
	PkgPath         string
	PkgName         string
	Imports         map[string]*imports.ImportEntry
	OriginalImports map[string]*imports.ImportEntry
	ImportsPathMap  map[string]*imports.ImportEntry
	UsedImports     map[string]struct{}
}

func (f *ParsedFile) ListInterfaces(names ...string) []*ParsedInterface {
	var resp []*ParsedInterface
	// Iterate through the declarations in the file
	for _, decl := range f.Ref.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || len(names) > 0 && !slices.Contains(names, typeSpec.Name.Name) {
				continue
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			cur := &ParsedInterface{
				ParsedFile: f,
				Type:       typeSpec,
				Ref:        interfaceType,
				Name:       typeSpec.Name.Name,
			}
			cur.GenericsTypes, cur.GenericsNames = cur.getGenericsInfo()
			resp = append(resp, cur)
		}
	}
	return resp
}

func (f *ParsedFile) writeImports(w io.Writer) {
	// Write import statements
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, "\t\"fmt\"\n")
	fmt.Fprintf(w, "\t\"testing\"\n")
	fmt.Fprintf(w, "\tmockSetup \"github.com/sonalys/fake/boilerplate\"\n")
	for name := range f.UsedImports {
		info := f.Imports[name]
		fmt.Fprintf(w, "\t")
		if info.Alias != "" {
			fmt.Fprintf(w, "%s ", info.Alias)
		}
		fmt.Fprintf(w, "\"%s\"\n", info.Path)
	}
	fmt.Fprintf(w, ")\n\n")
}
