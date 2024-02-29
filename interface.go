package fake

import (
	"fmt"
	"go/ast"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"golang.org/x/tools/go/packages"
)

type ParsedInterface struct {
	ParsedFile             *ParsedFile
	Type                   *ast.TypeSpec
	Ref                    *ast.InterfaceType
	Name                   string
	GenericsHeader         string
	GenericsNamelessHeader string
	GenericsName           []string
}

func (i *ParsedInterface) ListFields() []*ParsedField {
	return i.ParsedFile.Generator.ListInterfaceFields(i, i.ParsedFile.Imports)
}

func (g *Generator) ListInterfaceFields(i *ParsedInterface, imports map[string]*PackageInfo) []*ParsedField {
	if i.Ref.Methods == nil {
		return nil
	}
	var resp []*ParsedField
	for _, method := range i.Ref.Methods.List {
		switch t := method.Type.(type) {
		case *ast.FuncType:
			resp = append(resp, &ParsedField{
				Interface: i,
				Ref:       method,
				Name:      method.Names[0].Name,
			})
		case *ast.SelectorExpr:
			resp = append(resp, g.ListInterfaceFields(g.ParseInterface(t, i.ParsedFile.UsedImports, i.ParsedFile.Imports), imports)...)
		}
	}
	return resp
}

func (f *ParsedInterface) getTypeGenerics(t *ast.TypeSpec) (string, []string) {
	var genericsHeader string
	var genericsNames []string
	if t.TypeParams != nil {
		types := []string{}
		for _, t := range t.TypeParams.List {
			types = append(types, fmt.Sprintf("%s %s", t.Names[0].Name, f.printAstExpr(t.Type)))
			genericsNames = append(genericsNames, t.Names[0].Name)
		}
		genericsHeader = fmt.Sprintf("[%s]", strings.Join(types, ","))
	}
	return genericsHeader, genericsNames
}

func findInterfaceByName(file *ParsedFile, name string) (*ParsedInterface, map[string]*PackageInfo) {
	for _, i := range file.ListInterfaces() {
		if i.Name == name {
			return i, file.Imports
		}
	}
	return nil, nil
}

func (g *Generator) ParseInterface(ident *ast.SelectorExpr, usedImports map[string]struct{}, imports map[string]*PackageInfo) *ParsedInterface {
	// Packages can have different names than their path, Example: ctx "context" would return ctx.
	pkgName := ident.X.(*ast.Ident).Name
	// Example: "Context" from context.Context.
	pkgType := ident.Sel.Name
	pkgPath := imports[pkgName].ImportPath
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		log.Fatal().Msgf("could not load file '%s': %s", pkgPath, err)
	}
	if len(pkgs) == 0 {
		log.Fatal().Msgf("no package found in file: %s", pkgPath)
	}
	pkg := pkgs[0]
	for _, file := range pkg.GoFiles {
		parsed := g.ParseFile(file)
		i, importsInfo := findInterfaceByName(parsed, pkgType)
		if i == nil {
			continue
		}
		for name, info := range importsInfo {
			imports[name] = info
		}
		// Link the usedImports for the 2 files since we will be printing their ast.Type.
		parsed.UsedImports = usedImports
		return i
	}
	return nil
}

func (i *ParsedInterface) GetGenericsInfo() (string, []string) {
	return i.getTypeGenerics(i.Type)
}

func (i *ParsedInterface) WriteStruct(w io.Writer) {
	// Write struct definition implementing the interface
	fmt.Fprintf(w, "type %s%s struct {\n", i.Name, i.GenericsHeader)
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "\tsetup%s mockSetup.Mock[", field.Name)
		i.PrintMethodHeader(w, "func", field)
		fmt.Fprintf(w, "]\n")
	}
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) WriteInitializer(w io.Writer) {
	fmt.Fprintf(w, "func New%s%s(t *testing.T) *%s%s {\n", i.Name, i.GenericsHeader, i.Name, i.GenericsNamelessHeader)
	fmt.Fprintf(w, "\treturn &%s%s{\n", i.Name, i.GenericsNamelessHeader)
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "\t\tsetup%s: mockSetup.NewMock[", field.Name)
		i.PrintMethodHeader(w, "func", field)
		fmt.Fprintf(w, "](t),\n")
	}
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) WriteAssertExpectations(w io.Writer) {
	fmt.Fprintf(w, "func (s *%s%s) AssertExpectations(t *testing.T) bool {\n", i.Name, i.GenericsNamelessHeader)
	fmt.Fprintf(w, "\treturn ")
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "s.setup%s.AssertExpectations(t) &&\n\t\t", field.Name)
	}
	fmt.Fprintf(w, "true\n")
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) WriteOnMethod(w io.Writer, methodName string, f *ParsedField) {
	fmt.Fprintf(w, "func (s *%s%s) On%s(funcs ...", i.Name, i.GenericsNamelessHeader, methodName)
	i.PrintMethodHeader(w, "func", f)
	fmt.Fprintf(w, ") mockSetup.Config {\n")
	fmt.Fprintf(w, "\treturn s.setup%s.Append(funcs...)\n", methodName)
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) WriteMethod(w io.Writer, methodName string, f *ParsedField) {
	fmt.Fprintf(w, "func (s *%s%s) ", i.Name, i.GenericsNamelessHeader)
	i.PrintMethodHeader(w, methodName, f)
	fmt.Fprintf(w, "{\n")
	var callingNames []string
	funcType := f.Ref.Type.(*ast.FuncType)
	for i := range funcType.Params.List {
		callingNames = append(callingNames, getFieldCallingName(i, funcType.Params.List[i]))
	}
	var argNames []string
	var argFlag []string
	for i := range funcType.Params.List {
		argNames = append(argNames, getFieldName(i, funcType.Params.List[i]))
		nameLen := len(funcType.Params.List[i].Names)
		if nameLen == 0 {
			nameLen = 1
		}
		for j := 0; j < nameLen; j++ {
			argFlag = append(argFlag, "%v")
		}
	}
	fmt.Fprintf(w, "\tf, ok := s.setup%s.Call()\n", methodName)
	fmt.Fprintf(w, "\tif !ok {\n")
	fmt.Fprintf(
		w, "\t\tpanic(fmt.Sprintf(\"unexpected call %s(%s)\", %v))\n",
		methodName, strings.Join(argFlag, ","), strings.Join(argNames, ","),
	)
	fmt.Fprintf(w, "\t}\n")
	if funcType.Results.NumFields() > 0 {
		fmt.Fprintf(w, "\treturn (*f)(%s)\n", strings.Join(callingNames, ","))
	} else {
		fmt.Fprintf(w, "\t(*f)(%s)\n", strings.Join(callingNames, ","))
	}
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) WriteStructMethods(file io.Writer) {
	// Implement each method in the interface with dummy bodies.
	for _, field := range i.ParsedFile.Generator.ListInterfaceFields(i, i.ParsedFile.Imports) {
		methodName := field.Ref.Names[0].Name
		i.WriteOnMethod(file, methodName, field)
		i.WriteMethod(file, methodName, field)
	}
}

func (i *ParsedInterface) WriteMock(w io.Writer) {
	log.Info().Msgf("generating mock for %s", i.Name)
	i.WriteStruct(w)
	i.WriteInitializer(w)
	i.WriteAssertExpectations(w)
	i.WriteStructMethods(w)
}
