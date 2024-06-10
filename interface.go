package fake

import (
	"fmt"
	"go/ast"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/fake/internal/imports"
	pkgs "github.com/sonalys/fake/internal/packages"
)

type ParsedInterface struct {
	ParsedFile    *ParsedFile
	Type          *ast.TypeSpec
	Ref           *ast.InterfaceType
	Name          string
	GenericsTypes []string
	GenericsNames []string
	// TranslateGenericNames translates generic type names from any imported interfaces.
	// Example:
	//	type A[T any] interface{ B[T] }
	//	type B[J any] interface{ Method() J }
	// it should have method Method() T when implementing A mock.
	TranslateGenericNames []string

	fieldsCache []*ParsedField
}

func (i *ParsedInterface) ListFields() []*ParsedField {
	if i.fieldsCache != nil {
		return i.fieldsCache
	}
	i.fieldsCache = i.ParsedFile.Generator.listInterfaceFields(i, i.ParsedFile.Imports)
	return i.fieldsCache
}

// ListInterfaceFields receives an interface to translate fields into fields.
// It cannot be a ParsedInterface method because we need to translate imports from the original file,
// some interfaces are originated from external packages.
func (g *Generator) listInterfaceFields(i *ParsedInterface, imports map[string]*imports.ImportEntry) []*ParsedField {
	if i == nil || i.Ref.Methods == nil {
		return nil
	}
	if i.fieldsCache != nil {
		return i.fieldsCache
	}
	var resp []*ParsedField
	for _, field := range i.Ref.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:
			for _, name := range field.Names {
				resp = append(resp, &ParsedField{
					Interface: i,
					Ref:       field,
					Name:      name.Name,
				})
			}
		case *ast.SelectorExpr:
			// Interface from another package.
			resp = append(resp, g.listInterfaceFields(g.parseInterface(t, i.ParsedFile), imports)...)
		case *ast.IndexExpr:
			// Interface from same package.
			// Child interface has only 1 generic parameter.
			ident := t.X.(*ast.Ident)
			newInterface := i.ParsedFile.FindInterfaceByName(ident.Name)
			index, ok := t.Index.(*ast.Ident)
			if !ok {
				log.Fatal().Msgf("unexpected type of *ast.FuncType: expected *ast.Ident, got %T", t.Index)
			}
			newInterface.TranslateGenericNames = []string{index.Name}
			newInterface.GenericsNames[0] = i.printAstExpr(index)
			// IndexExpr means 1 parameter, so newInterface only has 1 type and 1 name.
			newInterface.Type.TypeParams.List[0].Names[0].Name = index.Name
			resp = append(resp, g.listInterfaceFields(newInterface, imports)...)
		case *ast.IndexListExpr:
			// Interface from same package.
			// Child interface has many generic parameter.
			ident := t.X.(*ast.Ident)
			newInterface := i.ParsedFile.FindInterfaceByName(ident.Name)
			fields := g.listInterfaceFields(newInterface, imports)
			var names []string
			for _, index := range t.Indices {
				names = append(names, i.printAstExpr(index))
			}
			newInterface.TranslateGenericNames = names
			resp = append(resp, fields...)
		}
	}

	// Deduplicate resp entries with the same name
	nameMap := make(map[string]bool)
	deduplicatedResp := []*ParsedField{}
	for _, field := range resp {
		if !nameMap[field.Name] {
			deduplicatedResp = append(deduplicatedResp, field)
			nameMap[field.Name] = true
		}
	}
	resp = deduplicatedResp
	i.fieldsCache = resp
	return resp
}

func (f *ParsedInterface) getTypeGenerics(t *ast.TypeSpec) ([]string, []string) {
	var genericsTypes []string
	var genericsNames []string
	if t.TypeParams != nil {
		for _, t := range t.TypeParams.List {
			for _, name := range t.Names {
				genericsTypes = append(genericsTypes, f.printAstExpr(t.Type))
				genericsNames = append(genericsNames, name.Name)
			}
		}
	}
	return genericsTypes, genericsNames
}

func (f *ParsedFile) FindInterfaceByName(name string) *ParsedInterface {
	for _, i := range f.ListInterfaces() {
		if i.Name == name {
			return i
		}
	}
	return nil
}

func (g *Generator) parseInterface(ident *ast.SelectorExpr, f *ParsedFile) *ParsedInterface {
	interfaceName := ident.Sel.Name
	// Packages can have different names than their path, Example: ctx "context" would return ctx.
	pkgName := ident.X.(*ast.Ident).Name
	pkgInfo, ok := f.Imports[pkgName]
	if !ok {
		return nil
	}
	pkg, ok := pkgs.Parse(g.goModFilename, pkgInfo.Path)
	if !ok {
		return nil
	}
	var i *ParsedInterface
	var externalFile *ParsedFile
	var err error
	for _, filename := range pkg.Files {
		externalFile, err = g.ParseFile(filename)
		if err != nil {
			log.Fatal().Msgf("failed to parse file: %s", filename)
		}
		i = externalFile.FindInterfaceByName(interfaceName)
		if i != nil {
			break
		}
	}
	oldImportList := externalFile.Imports
	externalFile.OriginalImports = oldImportList
	// If different imports collide with same name, we alias the new imports being used.
	for name, info := range oldImportList {
		// If the pkg path is already there, then entry was already added.
		if _, ok := f.ImportsPathMap[info.Path]; ok {
			continue
		}
		currentEntry, collides := f.Imports[name]
		if !collides {
			f.Imports[name] = info
			f.ImportsPathMap[info.Path] = info
			continue
		}
		if collides && info.Path != currentEntry.Path {
			alias := fmt.Sprintf("%s1", name)
			info.Alias = alias
			f.Imports[alias] = info
			f.ImportsPathMap[info.Path] = info
		}
	}
	externalFile.UsedImports = f.UsedImports
	externalFile.Imports = f.Imports
	externalFile.ImportsPathMap = f.ImportsPathMap
	return i
}

func (i *ParsedInterface) getGenericsInfo() ([]string, []string) {
	return i.getTypeGenerics(i.Type)
}

func (i *ParsedInterface) writeGenericsHeader() string {
	if len(i.GenericsTypes) == 0 {
		return ""
	}
	var merge []string
	for idx := range i.GenericsTypes {
		merge = append(merge, fmt.Sprintf("%s %s", i.GenericsNames[idx], i.GenericsTypes[idx]))
	}
	return fmt.Sprintf("[%s]", strings.Join(merge, ", "))
}

func (i *ParsedInterface) writeGenericsNameHeader() string {
	if len(i.GenericsTypes) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(i.GenericsNames, ", "))
}

func (i *ParsedInterface) writeStruct(w io.Writer) {
	// Write struct definition implementing the interface
	fmt.Fprintf(w, "type %s%s struct {\n", i.Name, i.writeGenericsHeader())
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "\tsetup%s mockSetup.Mock[", field.Name)
		i.PrintMethodHeader(w, "func", field)
		fmt.Fprintf(w, "]\n")
	}
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) writeInitializer(w io.Writer) {
	genericsNameHeader := i.writeGenericsNameHeader()
	fmt.Fprintf(w, "func New%s%s(t *testing.T) *%s%s {\n", i.Name, i.writeGenericsHeader(), i.Name, genericsNameHeader)
	fmt.Fprintf(w, "\treturn &%s%s{\n", i.Name, genericsNameHeader)
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "\t\tsetup%s: mockSetup.NewMock[", field.Name)
		i.PrintMethodHeader(w, "func", field)
		fmt.Fprintf(w, "](t),\n")
	}
	fmt.Fprintf(w, "\t}\n")
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) writeAssertExpectations(w io.Writer) {
	genericsTypeHeader := i.writeGenericsNameHeader()
	fmt.Fprintf(w, "func (s *%s%s) AssertExpectations(t *testing.T) bool {\n", i.Name, genericsTypeHeader)
	fmt.Fprintf(w, "\treturn ")
	for _, field := range i.ListFields() {
		fmt.Fprintf(w, "s.setup%s.AssertExpectations(t) &&\n\t\t", field.Name)
	}
	fmt.Fprintf(w, "true\n")
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) writeOnMethod(w io.Writer, methodName string, f *ParsedField) {
	fmt.Fprintf(w, "func (s *%s%s) On%s(funcs ...", i.Name, i.writeGenericsNameHeader(), methodName)
	i.PrintMethodHeader(w, "func", f)
	fmt.Fprintf(w, ") mockSetup.Config {\n")
	fmt.Fprintf(w, "\treturn s.setup%s.Append(funcs...)\n", methodName)
	fmt.Fprintf(w, "}\n\n")
}

func (i *ParsedInterface) writeMethod(w io.Writer, methodName string, f *ParsedField) {
	fmt.Fprintf(w, "func (s *%s%s) ", i.Name, i.writeGenericsNameHeader())
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
		argNames = append(argNames, getFieldName(i))
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

func (i *ParsedInterface) writeStructMethods(file io.Writer) {
	// Implement each method in the interface with dummy bodies.
	for _, field := range i.ParsedFile.Generator.listInterfaceFields(i, i.ParsedFile.Imports) {
		methodName := field.Name
		i.writeOnMethod(file, methodName, field)
		i.writeMethod(file, methodName, field)
	}
}

func (i *ParsedInterface) write(w io.Writer) {
	i.writeStruct(w)
	i.writeInitializer(w)
	i.writeAssertExpectations(w)
	i.writeStructMethods(w)
}
