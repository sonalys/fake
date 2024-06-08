package fake

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/sonalys/fake/internal/imports"
	"github.com/sonalys/fake/internal/packages"
)

func (file *ParsedFile) importConflictResolution(importUsedName string, importPath string) string {
	info, ok := file.OriginalImports[importUsedName]
	// If the original import is found, use either name or alias.
	if ok {
		pkgInfo := file.ImportsPathMap[importPath]
		if pkgInfo.Alias != "" {
			file.UsedImports[pkgInfo.Alias] = struct{}{}
			return pkgInfo.Alias
		}
		file.UsedImports[pkgInfo.Name] = struct{}{}
		return pkgInfo.Name
	}
	info = &imports.ImportEntry{
		PackageInfo: packages.PackageInfo{
			Path: file.PkgPath,
			Name: file.PkgName,
		},
	}
	file.Imports[file.PkgName] = info
	file.ImportsPathMap[file.PkgPath] = info
	file.UsedImports[file.PkgName] = struct{}{}
	return file.PkgName
}

func (f *ParsedInterface) printAstExpr(expr ast.Expr) string {
	file := f.ParsedFile
	gen := file.Generator
	// Extract package and type name
	switch fieldType := expr.(type) {
	case *ast.Ident:
		// If the type name starts with a lowercase letter, it's an internal type.
		if strings.ToLower(fieldType.Name[:1]) == fieldType.Name[:1] {
			return fieldType.Name
		}
		// If it's a generic type, we don't need to print package name with it.
		for idx, name := range f.GenericsNames {
			if name == fieldType.Name {
				if len(f.TranslateGenericNames) > 0 {
					return f.TranslateGenericNames[idx]
				}
				return fieldType.Name
			}
		}
		return fmt.Sprintf("%s.%s", file.importConflictResolution(file.PkgName, file.PkgPath), fieldType.Name)
	case *ast.SelectorExpr:
		// Type from another package.
		pkgName := fmt.Sprint(fieldType.X)
		if file.OriginalImports != nil {
			pkgInfo, ok := file.OriginalImports[pkgName]
			newPkgInfo := file.ImportsPathMap[pkgInfo.Path]
			var pkgAlias = newPkgInfo.Name
			if newPkgInfo.Alias != "" {
				pkgAlias = newPkgInfo.Alias
			}
			if ok {
				file.UsedImports[pkgAlias] = struct{}{}
				return fmt.Sprintf("%s.%s", pkgAlias, fieldType.Sel)
			}
		}
		file.UsedImports[pkgName] = struct{}{}
		return fmt.Sprintf("%s.%s", pkgName, fieldType.Sel)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", f.printAstExpr(fieldType.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", f.printAstExpr(fieldType.Elt))
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", f.printAstExpr(fieldType.Elt))
	case *ast.ChanType:
		switch fieldType.Dir {
		case ast.RECV:
			return fmt.Sprintf("<-chan %s", f.printAstExpr(fieldType.Value))
		case ast.SEND:
			return fmt.Sprintf("chan<-%s", f.printAstExpr(fieldType.Value))
		case ast.SEND | ast.RECV:
			return fmt.Sprintf("chan %s", f.printAstExpr(fieldType.Value))
		}
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", f.printAstExpr(fieldType.Key), f.printAstExpr(fieldType.Value))
	case *ast.FuncType:
		b := &strings.Builder{}
		f.PrintMethodHeader(b, "func", &ParsedField{
			Interface: f,
			Ref: &ast.Field{
				Type: expr,
			},
			Name: "func",
		})
		return b.String()
	case *ast.InterfaceType:
		if fieldType.Methods.NumFields() == 0 {
			return "interface{}"
		}
		methods := gen.listInterfaceFields(&ParsedInterface{
			Ref: fieldType,
		}, file.Imports)
		b := &strings.Builder{}
		for _, method := range methods {
			for _, methodName := range method.Ref.Names {
				b.WriteString("\t\t")
				f.PrintMethodHeader(b, methodName.Name, &ParsedField{
					Interface: f,
					Ref: &ast.Field{
						Type: method.Ref.Type.(*ast.FuncType),
					},
					Name: methodName.Name,
				})
			}
		}
		return fmt.Sprintf("interface{\n%s\n}", b.String())
	}
	return ""
}

func getAstTypeName(expr ast.Expr) string {
	switch fieldType := expr.(type) {
	case *ast.Ident:
		return fieldType.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s", fieldType.X)
	case *ast.StarExpr:
		return getAstTypeName(fieldType.X)
	case *ast.ArrayType:
		return getAstTypeName(fieldType.Elt)
	case *ast.Ellipsis:
		return getAstTypeName(fieldType.Elt)
	case *ast.ChanType:
		return getAstTypeName(fieldType.Value)
	}
	return ""
}
