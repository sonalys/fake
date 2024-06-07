package fake

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

type ParsedField struct {
	Interface *ParsedInterface
	Ref       *ast.Field
	Name      string
	Type      string
}

func getFieldName(i int, field *ast.Field) string {
	if len(field.Names) == 0 {
		return fmt.Sprintf("a%d", i)
	}
	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	return strings.Join(names, ", ")
}

func getFieldCallingName(i int, field *ast.Field) string {
	if len(field.Names) == 0 {
		return fmt.Sprintf("a%d", i)
	}
	switch field.Type.(type) {
	case *ast.Ellipsis:
		return fmt.Sprintf("%s...", getFieldName(i, field))
	}
	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	return strings.Join(names, ",")
}

func (f *ParsedInterface) PrintAstField(i int, field *ast.Field, printName bool) string {
	typeName := f.printAstExpr(field.Type)
	if printName {
		return fmt.Sprintf("%s %s", getFieldName(i, field), typeName)
	}
	return typeName
}

func (f *ParsedInterface) PrintAstFields(implFile io.Writer, fields []*ast.Field, printName bool) {
	var buffer []string
	for i, field := range fields {
		buffer = append(buffer, f.PrintAstField(i, field, printName))
	}
	fmt.Fprint(implFile, strings.Join(buffer, ", "))
}
