package temple

import (
	"fmt"
	"html/template"
	logger "log"

	"github.com/russross/blackfriday"
)

// DefaultFuncMap are the default template funcs provided with this package
var DefaultFuncMap = template.FuncMap{
	"markd":    MarkdFunc,
	"noescape": NoescapeFunc,
}

// MarkdFunc provides a markdown pipable func
func MarkdFunc(args ...interface{}) template.HTML {
	b := blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...)))

	return template.HTML(b)
}

// NoescapeFunc provides a function to output non-escpaed HTML
func NoescapeFunc(v interface{}) template.HTML {
	str, ok := v.(string)
	if !ok {
		return template.HTML("")
	}

	return template.HTML(str)
}

// TemplateFuncs merges two (or more) sets of template funcs
func TemplateFuncs(a template.FuncMap, b ...template.FuncMap) template.FuncMap {
	for _, v := range b {
		for k, fn := range v {
			a[k] = fn // override existing maps
		}
	}

	return a
}

type LogLevel int

const (
	NONE LogLevel = iota
	INFO
	WARN
	ERROR
	DEBUG
)

var LOG_LEVEL LogLevel = NONE

func init() {
	logger.SetPrefix("[temple] ")
}

// log writes at level or below
func log(ll LogLevel, f string, v ...interface{}) {
	if LOG_LEVEL >= ll {
		logger.Printf(f, v...)
	}
}

func SetLogLevel(ll LogLevel) {
	LOG_LEVEL = ll
}
