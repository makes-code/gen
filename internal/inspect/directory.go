package inspect

import (
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
)

type ParsedFile struct {
	*ast.File
	tokens *token.FileSet
}

type Files []string

func (fs Files) FindAndParse(fn func(string) bool) (*ParsedFile, error) {
	pf := ParsedFile{tokens: token.NewFileSet()}

	var file string
	for _, f := range fs {
		if fn(f) {
			file = f
			break
		}
	}

	if file == "" {
		return nil, errors.New("failed to find file")
	}

	src, srcErr := ioutil.ReadFile(file)
	if srcErr != nil {
		return nil, srcErr
	}

	parsed, parsedErr := parser.ParseFile(pf.tokens, file, src, parser.ParseComments)
	if parsedErr != nil {
		return nil, parsedErr
	}

	pf.File = parsed
	return &pf, nil
}

func DirectoryGoFiles(path string) (string, Files, error) {
	pkg, pkgErr := build.ImportDir(path, 0)
	if pkgErr != nil {
		return "", nil, pkgErr
	}

	files := make([]string, len(pkg.GoFiles))
	copy(files, pkg.GoFiles)

	if path != "." {
		for i, f := range files {
			files[i] = filepath.Join(path, f)
		}
	}
	return pkg.Name, files, nil
}
