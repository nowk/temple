package temple

import (
	"bytes"
	"path"
	"testing"

	"github.com/nowk/mockvar"
	"github.com/spf13/afero"
)

func TestParsesTemplatesFromDefinedDir(t *testing.T) {
	td := SetupFs(FsTree{
		"/html": FsTree{
			"index.html": `<% define "index" %>Hello World!<% end %>`,
		},
	}, &Fs)
	defer td()

	var (
		tmpl = NewTemplates("/html")

		err = tmpl.Parse()
	)
	if err != nil {
		t.Fatal(err)
	}

	w := bytes.NewBuffer(nil)

	err = tmpl.Render(w, "index", nil)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	var (
		exp = "Hello World!"
		got = w.String()
	)
	if exp != got {
		t.Errorf("expected %s, got %s", exp, got)
	}
}

func TestChangeDelims(t *testing.T) {
	td := SetupFs(FsTree{
		"/html": FsTree{
			"index.html": `{{ define "index" }}Hello World!{{ end }}`,
		},
	}, &Fs)
	defer td()

	var (
		tmpl = NewTemplates("/html")

		err = tmpl.Parse("{{", "}}")
	)
	if err != nil {
		t.Fatal(err)
	}

	w := bytes.NewBuffer(nil)

	err = tmpl.Render(w, "index", nil)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	var (
		exp = "Hello World!"
		got = w.String()
	)
	if exp != got {
		t.Errorf("expected %s, got %s", exp, got)
	}
}

// test helpers

type FsTree map[string]interface{}

func SetupFs(tree FsTree, fs interface{}) func() {
	mockFs := &afero.MemMapFs{}

	re := mockvar.Mock(fs, mockFs)
	makeTree("/", tree, mockFs)

	return re
}

func makeTree(root string, tree FsTree, fs afero.Fs) {
	for k, v := range tree {
		node := path.Join(root, k)

		switch content := v.(type) {
		case string:
			f, err := fs.Create(node)
			if err != nil {
				panic(err)
			}
			if content != "" {
				f.WriteString(content)
			}
			f.Close()
		case bool, nil, FsTree:
			fs.Mkdir(node, 0777)

			// if fstree, continue down the branch to create fs
			if br, ok := v.(FsTree); ok {
				makeTree(node, br, fs)
			}
		}
	}
}
