package templ

const codeTpl = `
{{/* 定义全局变量  */}}
{{$type := .Name}} 
{{$indexs := .IndexList}}
{{$indexMap := .Indexs}}
{{$pkg := .PbPkg}}

/*
* 本代码由cfgtool工具生成，请勿手动修改
*/

package {{.Pkg}}

import (
	"poker_server/common/config"
	"poker_server/common/pb"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
)

var obj = atomic.Value{}

type {{$type}}Data struct {
{{- range $index := $indexs -}}
    {{- if eq $index.Type.ValueOf 2 -}}         {{/*ValueOfList*/}}
        _{{$index.Name}} []*{{$pkg}}.{{$type}}
    {{- else if eq $index.Type.ValueOf 3 -}}    {{/*ValueOfMap*/}}
        {{- if or (eq $index.Type.Name "int32") (eq $index.Type.Name "uint32") (eq $index.Type.Name "int64") (eq $index.Type.Name "uint64") }}
        _Max{{$index.Name}} {{$index.Type.Name}}
        {{- end}}
        _{{$index.Name}} map[{{$index.Type.Name}}]*{{$pkg}}.{{$type}}
    {{- else if eq $index.Type.ValueOf 4 -}}    {{/*ValueOfGroup*/}}
        _{{$index.Name}} map[{{$index.Type.Name}}][]*{{$pkg}}.{{$type}}
    {{- end -}}
{{- end -}}
}

// 注册函数
func init() {
    config.Register("{{$type}}", parse)
}

func parse(buf string) error {
    data := &{{$pkg}}.{{$type}}Ary{}
    if err := proto.UnmarshalText(buf, data); err != nil {
        return err
    }

{{if or (index $indexMap 3) (index $indexMap 4)}}
{{range $index := $indexs -}}
    {{- if eq $index.Type.ValueOf 3 -}}    {{/*ValueOfMap*/}}
        {{- if or (eq $index.Type.Name "int32") (eq $index.Type.Name "uint32") (eq $index.Type.Name "int64") (eq $index.Type.Name "uint64") }}
        var _Max{{$index.Name}} {{$index.Type.Name}}
        {{- end}}
        _{{$index.Name}} := make(map[{{$index.Type.Name}}]*{{$pkg}}.{{$type}})
    {{- else if eq $index.Type.ValueOf 4 -}}    {{/*ValueOfGroup*/}}
        _{{$index.Name}} := make(map[{{$index.Type.Name}}][]*{{$pkg}}.{{$type}})
    {{- end -}}  
{{- end}}
    for _, item :=range data.Ary {
{{- range $index := $indexs -}} 
    {{$key := $index.Value "item" ","}}
    {{- if eq $index.Type.ValueOf 3 -}}    {{/*ValueOfMap*/}}
        {{- if or (eq $index.Type.TypeOf 1) (eq $index.Type.TypeOf 2) -}} {{/*TypeOfBase*/}}
            {{- if or (eq $index.Type.Name "int32") (eq $index.Type.Name "uint32") (eq $index.Type.Name "int64") (eq $index.Type.Name "uint64") }}
                if _Max{{$index.Name}} < item.{{$index.Name}} {
                    _Max{{$index.Name}} = item.{{$index.Name}}
                }
            {{- end}}
            _{{$index.Name}}[{{$key}}] = item
        {{- else if eq $index.Type.TypeOf 3 -}} {{/*TypeOfStruct*/}}
            key{{$index.Name}} := {{$index.Type.Name}}{ {{$key}} }
            _{{$index.Name}}[key{{$index.Name}}] = item
        {{- end -}}
    {{- else if eq $index.Type.ValueOf 4 -}}    {{/*ValueOfGroup*/}}
        {{- if or (eq $index.Type.TypeOf 1) (eq $index.Type.TypeOf 2) -}} {{/*TypeOfBase*/}}
            _{{$index.Name}}[{{$key}}] = append(_{{$index.Name}}[{{$key}}], item)
        {{- else if eq $index.Type.TypeOf 3 -}} {{/*TypeOfStruct*/}}
            key{{$index.Name}} := {{$index.Type.Name}}{ {{$key}} }
            _{{$index.Name}}[key{{$index.Name}}] = append(_{{$index.Name}}[key{{$index.Name}}], item)
        {{- end -}}
    {{- end -}}  
{{- end -}}
    }
{{end}}
    obj.Store(&{{$type}}Data{
{{- range $index := $indexs}} 
    {{- if and (eq $index.Type.TypeOf 1) (eq $index.Type.ValueOf 3) -}} {{/*TypeOfBase*/}}
        {{- if or (eq $index.Type.Name "int32") (eq $index.Type.Name "uint32") (eq $index.Type.Name "int64") (eq $index.Type.Name "uint64") }}
            _Max{{$index.Name}}: _Max{{$index.Name}},
        {{- end -}}
    {{end}}
    {{- if or (eq $index.Type.ValueOf 3) (eq $index.Type.ValueOf 4)}}
        _{{$index.Name}}: _{{$index.Name}},
    {{- else}}
        _{{$index.Name}}: data.Ary,
    {{- end -}}  
{{- end}}
    })
    return nil
}

{{$index := index (index $indexMap 2) 0}}
{{if $index -}}
func SGet() *{{$pkg}}.{{$type}} {
    obj, ok := obj.Load().(*{{$type}}Data)
    if !ok {
        return nil
    }
    return obj._{{$index.Name}}[len(obj._{{$index.Name}})-1]
}

func LGet() (rets []*{{$pkg}}.{{$type}}) {
    obj, ok := obj.Load().(*{{$type}}Data)
    if !ok {
        return
    }
    rets = make([]*{{$pkg}}.{{$type}}, len(obj._{{$index.Name}}))
    copy(rets, obj._{{$index.Name}})
    return
}

func Walk(f func(*{{$pkg}}.{{$type}})bool) {
    obj, ok := obj.Load().(*{{$type}}Data)
    if !ok {
        return
    }
    for _, item := range obj._{{$index.Name}} {
        if !f(item) {
            return
        }
    }
}
{{- end}}

{{- range $index := $indexs -}} 
    {{$arg := $index.Arg ","}}
    {{$key := $index.Value "" ","}}
    {{- if eq $index.Type.ValueOf 3 -}}    {{/*ValueOfMap*/}}
{{- if or (eq $index.Type.Name "int32") (eq $index.Type.Name "uint32") (eq $index.Type.Name "int64") (eq $index.Type.Name "uint64") }}
func MGet{{$index.Name}}Key(val {{$index.Type.Name}}) {{$index.Type.Name}} {
    if obj, ok := obj.Load().(*{{$type}}Data); ok && val > obj._Max{{$index.Name}} {
        return obj._Max{{$index.Name}}
    }
    return val
}
{{- end}}

func MGet{{$index.Name}}({{$arg}}) *{{$pkg}}.{{$type}} {
    obj, ok := obj.Load().(*{{$type}}Data)
    if !ok {
        return nil
    }
    {{- if or (eq $index.Type.TypeOf 1) (eq $index.Type.TypeOf 2) -}}                       {{/*TypeOfBase*/}}
        if val, ok := obj._{{$index.Name}}[{{$key}}]; ok {
    {{- else if eq $index.Type.TypeOf 3 -}}                                                 {{/*TypeOfStruct*/}}
        if val, ok := obj._{{$index.Name}}[{{$index.Type.Name}}{ {{$key}} }]; ok {
    {{- end}}
        return val
    }
    return nil
}
    {{- else if eq $index.Type.ValueOf 4 -}}    {{/*ValueOfGroup*/}}
func GGet{{$index.Name}}({{$arg}}) (rets []*{{$pkg}}.{{$type}}) {
    obj, ok := obj.Load().(*{{$type}}Data)
    if !ok {
        return
    } 
    {{- if or (eq $index.Type.TypeOf 1) (eq $index.Type.TypeOf 2) -}} {{/*TypeOfBase*/}}
        if vals, ok := obj._{{$index.Name}}[{{$key}}]; ok {
    {{- else if eq $index.Type.TypeOf 3 -}} {{/*TypeOfStruct*/}}
        if vals, ok := obj._{{$index.Name}}[{{$index.Type.Name}}{ {{$key}} }]; ok {
    {{- end -}}
        rets = make([]*{{$pkg}}.{{$type}}, len(vals))
        copy(rets, vals)
        return
    }
    return
}
    {{- end -}}  
{{- end -}}

`
