package fake

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
)

type ParsedFile struct {
	Generator   *Generator
	Ref         *ast.File
	PkgPath     string
	PkgName     string
	Imports     map[string]*PackageInfo
	UsedImports map[string]struct{}
}

func (f *ParsedFile) ListInterfaces() []*ParsedInterface {
	var resp []*ParsedInterface
	// Iterate through the declarations in the file
	for _, decl := range f.Ref.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			continue
		}
		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
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
			cur.GenericsTypes, cur.GenericsNames = cur.GetGenericsInfo()
			resp = append(resp, cur)
		}
	}
	return resp
}

func (f *ParsedFile) WriteImports(w io.Writer) {
	// Write import statements
	fmt.Fprintf(w, "import (\n")
	fmt.Fprintf(w, "\t\"fmt\"\n")
	fmt.Fprintf(w, "\t\"testing\"\n")
	fmt.Fprintf(w, "\tmockSetup \"github.com/sonalys/fake/boilerplate\"\n")
	for name := range f.UsedImports {
		info, ok := f.Imports[name]
		if !ok {
			continue
		}
		if info.Name == name {
			fmt.Fprintf(w, "\t\"%s\"\n", info.ImportPath)
			continue
		}
		fmt.Fprintf(w, "\t%s \"%s\"\n", name, info.ImportPath)
	}
	fmt.Fprintf(w, ")\n\n")
}
