package inspect

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

type TypeFieldsOptions struct {
	Target string
	Pkg    string
}

func TypeFields(file *ParsedFile, opts TypeFieldsOptions) (fields []Field, err error) {
	ast.Inspect(file.File, func(node ast.Node) bool {
		if err != nil {
			return false
		}

		t, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Name.Name != opts.Target {
			return false
		}

		fields, err = collectTypeFields(file, opts, t)
		return false
	})
	return
}

func collectTypeFields(file *ParsedFile, opts TypeFieldsOptions, t *ast.TypeSpec) ([]Field, error) {
	var fields []field
	var err error

	if i, ok := t.Type.(*ast.InterfaceType); ok {
		fields, err = collectInterfaceInfo(file, i)
	}

	if s, ok := t.Type.(*ast.StructType); ok {
		fields, err = collectStructInfo(file, s)
	}

	if err != nil {
		return nil, err
	}

	out := make([]Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, NewField(opts.Pkg, opts.Target, f.name, f.typeRaw, f.tagsRaw))
	}

	return out, nil
}

type field struct {
	name    string
	typeRaw string
	tagsRaw string
}

func collectInterfaceInfo(file *ParsedFile, i *ast.InterfaceType) ([]field, error) {
	fields := make([]field, 0, len(i.Methods.List))

	for _, m := range i.Methods.List {
		fieldName := m.Names[0].Name

		if fieldName == "Builder" {
			continue
		}

		fieldType, err := parseType(file.tokens, m.Type)
		if err != nil {
			continue
		}
		// collectFieldInfo(ws, fieldName, fieldType, "")
		fields = append(fields, field{name: fieldName, typeRaw: fieldType})
	}
	return fields, nil
}

func collectStructInfo(file *ParsedFile, s *ast.StructType) ([]field, error) {
	fields := make([]field, 0, len(s.Fields.List))

	for _, f := range s.Fields.List {
		fieldName := f.Names[0].Name

		if fieldName == "Builder" {
			continue
		}

		fieldType, err := parseType(file.tokens, f.Type)
		if err != nil {
			continue
		}

		var fieldTag string
		if f.Tag != nil {
			fieldTag = f.Tag.Value
		}

		fields = append(fields, field{fieldName, fieldType, fieldTag})
	}
	return fields, nil
}

// func collectFieldInfo(pkgName, modelName, fieldName, fieldType, fieldTag string) Field {
// 	f := NewField(pkgName, modelName, fieldName, fieldType, fieldTag)
// 	ws.Fields = append(ws.Fields, f)
// 	for _, i := range f.Type.Imports() {
// 		ws.Imports.Include(i)
// 	}
// }

func parseType(fs *token.FileSet, node ast.Node) (string, error) {
	out := new(bytes.Buffer)
	if err := printer.Fprint(out, fs, node); err != nil {
		return "", err
	}

	outStr := out.String()
	return outStr[strings.Index(outStr, ")")+1:], nil
}
