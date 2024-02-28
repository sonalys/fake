package fake

import (
	"fmt"
	"go/ast"
	"io"
)

func (f *ParsedInterface) WriteMethodParams(implFile io.Writer, params []*ast.Field) {
	fmt.Fprintf(implFile, "(")
	f.PrintAstFields(implFile, params, true)
	fmt.Fprintf(implFile, ")")
}

func (f *ParsedInterface) WriteMethodResults(implFile io.Writer, results []*ast.Field) {
	if len(results) == 0 {
		return
	}
	fmt.Fprintf(implFile, " (")
	f.PrintAstFields(implFile, results, false)
	fmt.Fprintf(implFile, ") ")
}

func (f *ParsedInterface) PrintMethodHeader(file io.Writer, methodName string, field *ParsedField) {
	fmt.Fprint(file, methodName)
	funcType := field.Ref.Type.(*ast.FuncType)
	if fields := funcType.Params; fields != nil {
		field.Interface.WriteMethodParams(file, fields.List)
	}
	if fields := funcType.Results; fields != nil {
		field.Interface.WriteMethodResults(file, fields.List)
	}
}
