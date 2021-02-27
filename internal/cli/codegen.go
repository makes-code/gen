package cli

import (
	"bytes"
	"flag"
	"go/format"
	"html/template"
	"log"
	"os"
	"strings"

	"github.com/makes-code/gen/internal/inspect"
)

const (
	quoteEscaped = "&#34;"
	quoteRaw     = `"`
)

type CmdMeta struct {
	Name     string
	Help     string
	Synopsis string
}

type CmdCodegen struct {
	CmdMeta
	Flags    func(fs *flag.FlagSet)
	Runner   func(data inspect.Data) (string, interface{}, error)
	FileName func(systemName string) string
}

func NewCmdCodegen() *CmdCodegen {
	var cmd CmdCodegen
	return &cmd
}

func (cmd *CmdCodegen) Help() string     { return cmd.CmdMeta.Help }
func (cmd *CmdCodegen) Synopsis() string { return cmd.CmdMeta.Synopsis }

func (cmd *CmdCodegen) Run(args []string) int {
	var name, repo string

	fs := flag.NewFlagSet(cmd.Name, flag.ContinueOnError)
	fs.StringVar(&name, "name", "", "")
	fs.StringVar(&repo, "repo", "", "")

	if cmd.Flags != nil {
		cmd.Flags(fs)
	}

	if err := fs.Parse(args); err != nil {
		log.Println(err)
		return 1
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Print(err)
		return 1
	}

	names := inspect.NewNames(name, inspect.NamesOptions{})

	pkgName, files, filesErr := inspect.DirectoryGoFiles(wd)
	if filesErr != nil {
		log.Print(filesErr)
		return 1
	}

	parsed, parsedErr := files.FindAndParse(func(f string) bool {
		return strings.HasSuffix(f, names.System+".go")
	})
	if parsedErr != nil {
		log.Print(parsedErr)
		return 1
	}

	fieldsOpts := inspect.TypeFieldsOptions{Pkg: pkgName, Target: name}
	fields, fieldsErr := inspect.TypeFields(parsed, fieldsOpts)
	if fieldsErr != nil {
		log.Print(fieldsErr)
		return 1
	}

	imports, importsErr := inspect.FileImports(repo, parsed)
	if importsErr != nil {
		log.Print(importsErr)
		return 1
	}

	for _, field := range fields {
		imports.Include(field.Type.Imports()...)
	}

	tmpl, tmplData, tmplErr := cmd.Runner(inspect.Data{
		Pkg:     pkgName,
		Names:   names,
		Fields:  fields,
		Imports: imports,
	})
	if tmplErr != nil {
		log.Print(tmplErr)
		return 1
	}

	src, srcErr := generateCode(cmd.Name, tmpl, tmplData)
	if srcErr != nil {
		log.Print(srcErr)
		return 1
	}

	if err := writeFile(cmd.FileName(names.System), src); err != nil {
		log.Print(err)
		return 1
	}

	return 0
}

func generateCode(name, tmpl string, tmplData interface{}) ([]byte, error) {
	src := new(bytes.Buffer)
	if err := template.Must(
		template.New(name).Parse(tmpl),
	).Execute(src, tmplData); err != nil {
		return nil, err
	}

	return format.Source([]byte(
		strings.ReplaceAll(src.String(), quoteEscaped, quoteRaw),
	))
}

func writeFile(path string, data []byte) error {
	file, createErr := os.Create(path)
	if createErr != nil {
		return createErr
	}

	_, writeErr := file.Write(data)
	return writeErr
}
