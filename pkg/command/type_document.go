package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/makes-code/gen/internal/cli"
	"github.com/makes-code/gen/internal/inspect"
	"github.com/makes-code/gen/internal/utils"

	mcli "github.com/mitchellh/cli"
)

type typeDocumentInputs struct {
	tag     string
	include utils.StringArray
	exclude utils.StringArray
	strict  bool
}

func TypeDocument() (mcli.Command, error) {
	var inputs typeDocumentInputs

	return &cli.CmdCodegen{
		CmdMeta: cli.CmdMeta{
			Name:     "document",
			Help:     "Generate a document model",
			Synopsis: "Generate a document model",
		},
		Flags: func(fs *flag.FlagSet) {
			fs.StringVar(&inputs.tag, "tag", "", "")
			fs.Var(&inputs.include, "i", "")
			fs.Var(&inputs.exclude, "x", "")
			fs.BoolVar(&inputs.strict, "strict", false, "")
		},
		FileName: func(systemName string) string {
			var suffix string
			if inputs.tag != "" {
				suffix = "_" + strings.ToLower(inputs.tag)
			}
			return fmt.Sprintf("%s_gen_document%s.go", systemName, suffix)
		},
		Runner: func(data inspect.Data) (string, interface{}, error) {
			whitelist := map[string]string{}
			for _, f := range inputs.include {
				name := f
				var nameOverride string

				if strings.Contains(f, "=") {
					parts := strings.Split(f, "=")
					name, nameOverride = parts[0], parts[1]
				}

				whitelist[name] = nameOverride
			}

			blacklist := map[string]struct{}{}
			for _, f := range inputs.exclude {
				blacklist[f] = struct{}{}
			}

			var fields []inspect.Field
			for _, field := range data.Fields {
				if _, ok := blacklist[field.Names.Public]; ok {
					continue
				}

				nameOverride, ok := whitelist[field.Names.Public]
				if !ok && inputs.strict {
					continue
				}

				if field.Names.Public == "ID" {
					field.Names.Field = "_id"
				}

				if nameOverride != "" {
					field.Names.Field = nameOverride
				}

				fields = append(fields, field)
			}

			imports := data.Imports.New()
			imports.Use("bson", `"go.mongodb.org/mongo-driver/bson"`)
			for _, field := range fields {
				imports.Include(field.Type.Imports()...)
			}

			return tmplDocument, tmplDataDocument{
				Data: inspect.Data{
					Pkg:     data.Pkg,
					Names:   data.Names,
					Fields:  fields,
					Imports: imports,
				},
				Tag: inputs.tag,
			}, nil
		},
	}, nil
}

type tmplDataDocument struct {
	inspect.Data
	Tag string
}

var tmplDocument = `
{{$ := .Names}}
// This file is auto-generated by makes-code ... do not edit

package {{.Pkg}}

{{if not .Imports.Empty}}
import ({{range .Imports.Groups}}
{{range .}}  {{.}}
{{end -}}
{{end}})
{{end}}

type {{$.Public}}Document{{.Tag}}s []*{{$.Public}}Document{{.Tag}}

type {{$.Public}}Document{{.Tag}} struct {
	{{$.Private}}Data
}

type {{$.Private}}Document{{.Tag}} struct {
{{range .Fields}} {{.Names.Public}} {{.Type}} {{.Tag "bson"}}
{{end -}}
}

func To{{$.Public}}Document{{.Tag}}({{$.Short}} {{$.Public}}) *{{$.Public}}Document{{.Tag}} {
	return &{{$.Public}}Document{{.Tag}}{{"{"}}{{$.Private}}Data{{"{"}}
{{range .Fields}}    {{.Names.Private}}: {{$.Short}}.{{.Names.Public}}(),
{{end -}}
	{{"}"}}{{"}"}}
}

func ({{$.Short}} {{$.Public}}Document{{.Tag}}) MarshalBSON() ([]byte, error) {
	return bson.Marshal({{$.Private}}Document{{.Tag}}{
{{range .Fields}}    {{.Names.Public}}: {{$.Short}}.{{.Names.Private}},
{{end -}}
	})
}

func ({{$.Short}} *{{$.Public}}Document{{.Tag}}) UnmarshalBSON(data []byte) error {
	var tmp {{$.Private}}Document{{.Tag}}
	if err := bson.Unmarshal(data, &tmp); err != nil {
		return err
	}

	{{$.Short}}.{{$.Private}}Data = {{$.Private}}Data{
{{range .Fields}}    {{.Names.Private}}: tmp.{{.Names.Public}},
{{end -}}
	}
	return nil
}

func To{{$.Public}}Document{{.Tag}}s({{$.Private}}s {{$.Public}}s) {{$.Public}}Document{{.Tag}}s {
  docs := make({{$.Public}}Document{{.Tag}}s, len({{$.Private}}s))
	for i, {{$.Private}} := range {{$.Private}}s {
		docs[i] = To{{$.Public}}Document{{.Tag}}({{$.Private}})
	}
	return docs
}

func (docs {{$.Public}}Document{{.Tag}}s) {{$.Public}}s() {{$.Public}}s {
	{{$.Private}}s := make({{$.Public}}s, len(docs))
	for i, doc := range docs {
		{{$.Private}}s[i] = doc
	}
	return {{$.Private}}s
}
`
