package templ

import "text/template"

var (
	StringTpl *template.Template
	HashTpl   *template.Template
)

func init() {
	funcs := template.FuncMap{}
	StringTpl = template.Must(template.New("StringTpl").Funcs(funcs).Parse(stringTpl))
	HashTpl = template.Must(template.New("HashTpl").Funcs(funcs).Parse(hashTpl))
}
