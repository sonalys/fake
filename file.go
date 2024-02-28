package fake

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"
)

type ParsedFile struct {
	Generator   *Generator
	Ref         *ast.File
	PkgPath     string
	PkgName     string
	Imports     map[string]*PackageInfo
	UsedImports map[string]struct{}
}

// CreateFileAndFolders creates a file and the necessary folders if they don't exist.
func CreateFileAndFolders(filePath string) (*os.File, error) {
	// Get the directory path from the file path
	dir := filepath.Dir(filePath)
	// Create directories if they don't exist
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, err
	}
	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
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
			cur.GenericsHeader, cur.GenericsName = cur.GetGenericsInfo()
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
	fmt.Fprintf(w, "\t_ \"github.com/sonalys/fake/boilerplate\"\n")
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
