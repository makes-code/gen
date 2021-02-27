package inspect

import (
	"fmt"
	"go/build"
	"path"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/camelcase"
)

const (
	rootDir = "."
)

type Data struct {
	Pkg     string
	Names   Names
	Fields  []Field
	Imports Imports
}

type Field struct {
	Names Names
	Type  FieldType
	Tags  map[string]string
}

func NewField(pkg, model, name, rawType, rawTags string) Field {
	return Field{
		Names: NewNames(name, NamesOptions{Prefix: model}),
		Type:  newFieldType(pkg, rawType),
		Tags:  newFieldTags(rawTags),
	}
}

func (f Field) Tag(tag string) string { return fmt.Sprintf("`%s:%q`", tag, f.Names.Field) }

type FieldType interface {
	fmt.Stringer
	Imports() []string
}

func newFieldType(pkg, typeRaw string) FieldType {
	name := strings.TrimSpace(typeRaw)

	if strings.HasPrefix(name, "[]") {
		return arrayFieldType{newFieldType(pkg, name[2:])}
	}

	if strings.HasPrefix(name, "map") {
		return mapFieldType{
			newFieldType(pkg, name[4:strings.Index(name, "]")]),
			newFieldType(pkg, name[strings.Index(name, "]")+1:]),
		}
	}

	pointer := strings.HasPrefix(name, "*")
	if pointer {
		name = name[1:]
	}

	var typePkg, typeName string

	if unicode.IsUpper(rune(name[0])) {
		typePkg = pkg
		typeName = name
	} else if i := strings.Index(name, "."); i > 0 {
		typePkg = name[0:i]
		typeName = name[i+1:]
	} else {
		typeName = name
	}

	return scalarFieldType{pointer, typePkg, typeName}
}

type scalarFieldType struct {
	pointer bool
	pkg     string
	name    string
}

func (t scalarFieldType) String() string {
	sb := strings.Builder{}
	if t.pointer {
		sb.WriteString("*")
	}
	if t.pkg != "" {
		sb.WriteString(t.pkg + ".")
	}
	sb.WriteString(t.name)
	return sb.String()
}

func (t scalarFieldType) Imports() []string {
	if t.pkg == "" {
		return nil
	}
	return []string{t.pkg}
}

type arrayFieldType struct {
	elemType interface{}
}

func (t arrayFieldType) String() string {
	return fmt.Sprintf("[]%s", typeString(t.elemType))
}

func (t arrayFieldType) Imports() []string {
	return typeImports(t.elemType)
}

type mapFieldType struct {
	keyType   interface{}
	valueType interface{}
}

func (t mapFieldType) String() string {
	return fmt.Sprintf("map[%s]%s", typeString(t.keyType), typeString(t.valueType))
}

func (t mapFieldType) Imports() []string {
	var imports []string
	imports = append(imports, typeImports(t.keyType)...)
	imports = append(imports, typeImports(t.valueType)...)
	return imports
}

func typeImports(tt interface{}) []string {
	if t, ok := tt.(FieldType); ok {
		return t.Imports()
	}
	return nil
}

func typeString(tt interface{}) string {
	if t, ok := tt.(fmt.Stringer); ok {
		return t.String()
	}
	return "interface{}"
}

func newFieldTags(tagsRaw string) map[string]string {
	if tagsRaw == "" {
		return nil
	}

	tags := map[string]string{}
	fmt.Println("parsing tag:", tagsRaw)
	for _, t := range strings.Split(strings.Trim(tagsRaw, "`"), " ") {
		fmt.Println("tag:", t)
		p := strings.Split(t, ":")
		tags[p[0]] = p[1]
	}

	return tags
}

type Imports struct {
	repo         string
	stmts        []string
	stmtsByAlias map[string]string
}

func (i Imports) New() Imports {
	return Imports{repo: i.repo, stmtsByAlias: i.stmtsByAlias}
}

func (i *Imports) Add(alias, stmt string) {
	i.stmtsByAlias[alias] = stmt
}

func (i Imports) Empty() bool {
	return len(i.stmts) == 0
}

func (i *Imports) Include(aliases ...string) {
	for _, alias := range aliases {
		if stmt, ok := i.stmtsByAlias[strings.TrimSpace(alias)]; ok {
			i.stmts = append(i.stmts, stmt)
		}
	}
}

func (i *Imports) Use(alias, stmt string) {
	i.Add(alias, stmt)
	i.Include(alias)
}

func (i Imports) Groups() [][]string {
	var stdlib, app, vendor []string

	for _, stmt := range i.stmts {
		switch {
		case isAppImport(stmt, i.repo):
			app = append(app, stmt)
		case isVendorImport(stmt):
			vendor = append(vendor, stmt)
		default:
			stdlib = append(stdlib, stmt)
		}
	}

	var out [][]string
	if len(stdlib) > 0 {
		out = append(out, stdlib)
	}
	if len(app) > 0 {
		out = append(out, app)
	}
	if len(vendor) > 0 {
		out = append(out, vendor)
	}
	return out
}

type Names struct {
	Public  string
	Private string
	Display string
	Short   string
	System  string
	Field   string
}

type NamesOptions struct {
	Prefix            string
	WhitelistOverride string
}

func NewNames(publicName string, opts NamesOptions) Names {
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "_"
	}

	var short string
	if publicName != "" {
		short = publicName[0:1]
	}

	system := lowerName(publicName, "_", "", "")

	field := opts.WhitelistOverride
	if field == "" {
		field = system
	}

	return Names{
		Public:  publicName,
		Private: privateName(publicName, prefix),
		Display: lowerName(publicName, " ", "", ""),
		Short:   strings.ToLower(short),
		System:  system,
		Field:   field,
	}
}

func isAppImport(stmt, repo string) bool {
	if repo == "" {
		return false
	}
	return strings.Contains(stmt, repo)
}

func isVendorImport(stmt string) bool {
	return strings.Contains(stmt, ".")
}

func lowerName(name string, delim, lhs, rhs string) string {
	parts := camelcase.Split(name)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return fmt.Sprintf("%s%s%s", lhs, strings.Join(parts, delim), rhs)
}

func pkgName(importPath, srcDir string) string {
	pkg, err := build.Import(importPath, srcDir, build.IgnoreVendor)
	if err == nil {
		return pkg.Name
	}
	return pkgNameFromImportPath(importPath)
}

func pkgNameFromImportPath(importPath string) string {
	base := path.Base(importPath)
	if strings.HasPrefix(base, "v") {
		if _, err := strconv.Atoi(base[1:]); err == nil {
			if dir := path.Dir(importPath); dir != "." {
				return path.Base(dir)
			}
		}
	}
	return base
}

func privateName(args ...string) string {
	if len(args) == 0 {
		return ""
	}

	name := args[0]
	parent := "_"
	if len(args) > 1 {
		parent = args[1]
	}

	parts := camelcase.Split(name)
	if len(parts) == 0 {
		return name
	}

	parts[0] = strings.ToLower(parts[0])

	name = strings.Join(parts, "")

	if isGoWord(name) {
		if parent == "" {
			parent = "_"
		}

		name = privateName(parent) + strings.Title(name)
	}

	return name
}

func isGoWord(word string) bool {
	switch word {
	case "break",
		"case",
		"chan",
		"const",
		"continue",
		"default",
		"defer",
		"else",
		"fallthrough",
		"for",
		"func",
		"go",
		"goto",
		"if",
		"import",
		"interface",
		"map",
		"package",
		"range",
		"return",
		"select",
		"struct",
		"switch",
		"type",
		"var":
		return true
	}
	return false
}
