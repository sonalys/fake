package fake

import (
	"fmt"
	"go/ast"
	"strings"
)

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
		_, collision := (*file.UsedImports)[file.PkgName]
		var alias string
		if collision {
			// If collision never happened, then rename interface package reference.
			// If it already happened, then re-utilize the same name.
			// Appending 1 to the end should be enough to avoid any collision at all.
			if (*file.Imports)[file.PkgName].Path != file.PkgPath {
				file.PkgName = fmt.Sprintf("%s1", file.PkgName)
				alias = file.PkgName
			} else {
				return fmt.Sprintf("%s.%s", file.PkgName, fieldType.Name)
			}
		}
		// If we have an object, that means we need to translate the type from mock package to current package.
		(*file.Imports)[file.PkgName] = &PackageInfo{
			Name:  file.PkgName,
			Path:  file.PkgPath,
			Alias: alias,
		}
		(*file.UsedImports)[file.PkgName] = struct{}{}
		return fmt.Sprintf("%s.%s", file.PkgName, fieldType.Name)
	case *ast.SelectorExpr:
		// Type from another package
		(*file.UsedImports)[fmt.Sprint(fieldType.X)] = struct{}{}
		return fmt.Sprintf("%s.%s", fieldType.X, fieldType.Sel)
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
		methods := gen.ListInterfaceFields(&ParsedInterface{
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
