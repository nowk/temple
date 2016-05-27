package temple

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

var Fs afero.Fs = &afero.OsFs{}

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
	Reload  bool
	Dir     string
	T       *template.Template // parsed templates
	FuncMap template.FuncMap   // default template funcs

	// mu provides a locking mechanism to sync parsing in order avoid redefined
	// errors
	mu *sync.Mutex
}

// NewTemplates provides a new Templates
// Overiding FuncMap at this level will not reinclude the DefaultFuncMap, those
// will require redefinition
func NewTemplates(dir string, opts ...func(*Templates)) *Templates {
	t := &Templates{
		Reload:  false,
		Dir:     dir,
		FuncMap: DefaultFuncMap,

		mu: &sync.Mutex{},
	}
	for _, v := range opts {
		v(t)
	}
	return t
}

// Parse walks a directory tree and parses .html and .tmpl files
func (t *Templates) Parse(delims ...string) error {
	n := len(delims)
	if n > 0 && n != 2 {
		return fmt.Errorf("there must be exactly 2 delims")
	}
	if n == 0 {
		delims = []string{
			DelimL,
			DelimR,
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.T = template.New("-").Funcs(t.FuncMap)
	t.T.Delims(delims[0], delims[1])

	log(INFO, "templates: %s", t.Dir)

	return afero.Walk(t.Dir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)
		if ext == ".html" || ext == ".tmpl" {
			_, err := parseFiles(t.T, path)
			if err != nil {
				log(ERROR, "%s: %s", path, err)

				return err
			}

			log(INFO, "parsed: %s", path)
		} else {
			// skip, not something we want to parse
		}

		return nil
	}, Fs)
}

// Render renders the defined template to w. Data can be passed in as d as well
// as additional template funcs
func (t *Templates) Render(
	w io.Writer, name string, d interface{}, fns ...template.FuncMap) error {

	if t.Reload {
		log(INFO, "reload...")

		err := t.Parse()
		if err != nil {
			log(ERROR, "reload: %s", err)

			return err
		}
	}

	err := t.T.Funcs(TemplateFuncs(t.FuncMap, fns...)).ExecuteTemplate(w, name, d)
	if err != nil {
		log(ERROR, "%s: %s", name, err)

		return err
	}

	return nil
}

// parseFiles is the helper for the method and function. If the argument
// template is nil, it is created from the first file.
// NOTE this is copied from https://golang.org/src/text/template/helper.go?s=1608:1677#L34
// to utlilize the afero.Fs
func parseFiles(
	t *template.Template, filenames ...string) (*template.Template, error) {

	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	for _, filename := range filenames {
		// use afero to do the look up here and manually parse
		f, err := Fs.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var (
			name = filepath.Base(filename)

			tmpl *template.Template
		)
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		if _, err = tmpl.Parse(string(b)); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (t *Templates) Lookup(name string) (*template.Template, bool) {
	tmpl := t.T.Lookup(name)
	if tmpl == nil {
		return nil, false
	}

	return tmpl, true
}
