package inspect

import (
	"bytes"
	"go/ast"
	"go/printer"
	"strings"
)

func FileImports(repo string, file *ParsedFile) (Imports, error) {
	imports := Imports{repo: repo, stmtsByAlias: map[string]string{}}

	var err error
	ast.Inspect(file.File, func(node ast.Node) bool {
		if err != nil {
			return false
		}

		i, ok := node.(*ast.ImportSpec)
		if !ok {
			return true
		}

		var out bytes.Buffer
		if err = printer.Fprint(&out, file.tokens, i); err != nil {
			return false
		}

		if path := strings.Trim(i.Path.Value, `"`); path != "C" {
			var pkg string
			if i.Name != nil {
				pkg = i.Name.Name
			} else {
				pkg = pkgName(path, rootDir)
			}

			imports.Add(pkg, out.String())
		}

		return false
	})

	return imports, err
}
