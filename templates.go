package temple

import (
	"fmt"
	"html/template"
	"io"
	logger "log"
	"os"
	"path/filepath"
	"sync"

	"github.com/russross/blackfriday"
)

// default delim left and right
// due to angular {{}} collisions opted to change default delimters
var (
	DelimL = "<%"
	DelimR = "%>"
)

// DevsMode is a templateCfgFunc that sets Reload to true
func DevsMode(t *Templates) {
	t.Reload = true
}

type Templates struct {
	Reload bool
	Dir    string
	T      *template.Template // parsed templates
	Fns    template.FuncMap   // default template funcs

	// mu provides a locking mechanism to sync parsing in order avoid redefined
	// errors
	mu *sync.Mutex
}

// NewTemplates provides a new Templates
// Overiding Fns at this level will not reinclude the DefaultFns, those will
// require redefinition
func NewTemplates(dir string, opts ...func(*Templates)) *Templates {
	t := &Templates{
		Reload: false,
		Dir:    dir,
		Fns:    DefaultFns,

		mu: &sync.Mutex{},
	}
	for _, v := range opts {
		v(t)
	}
	return t
}

// Parse walks a directory tree and parses .html and .tmpl files
func (t *Templates) Parse() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.T = template.New("-").Funcs(t.Fns)
	t.T.Delims(DelimL, DelimR)

	log(INFO, "templates: %s", t.Dir)

	return filepath.Walk(t.Dir, func(path string, _ os.FileInfo,
		err error) error {

		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		if ext == ".html" || ext == ".tmpl" {
			_, err := t.T.ParseFiles(path)
			if err != nil {
				log(ERROR, "%s: %s", path, err)

				return err
			}

			log(INFO, "parsed: %s", path)
		} else {
			// skip, not something we want to parse
		}

		return nil
	})
}

// Render renders the defined template to w. Data can be passed in as d as well
// as additional template funcs
func (t *Templates) Render(w io.Writer, name string, d interface{},
	fns ...template.FuncMap) error {

	if t.Reload {
		log(INFO, "reload...")

		err := t.Parse()
		if err != nil {
			log(ERROR, "reload: %s", err)

			return err
		}
	}

	err := t.T.Funcs(TemplateFuncs(t.Fns, fns...)).ExecuteTemplate(w, name, d)
	if err != nil {
		log(ERROR, "%s: %s", name, err)

		return err
	}

	return nil
}

// DefaultFns are the default template funcs provided with this package
var DefaultFns = template.FuncMap{
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
