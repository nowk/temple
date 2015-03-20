# temple

<!-- [![Build Status](https://travis-ci.org/nowk/urit.svg?branch=master)](https://travis-ci.org/nowk/urit) -->
[![GoDoc](https://godoc.org/gopkg.in/nowk/temple.v0?status.svg)](http://godoc.org/gopkg.in/nowk/temple.v0)

A simple template render struct

## Install

    go get gopkg.in/nowk/temple.v0


## Usage

    t := temple.NewTemplates("app/views")
    err := t.Parse()
    if err != nil {
      // handle err
    }

    // in your handler
    h := func(w http.ResponseWriter, req *http.Request) {
      t.Render(w, "template-name", nil)
    }

To force reload on each request you can send in the `DevsMode` configuration.

    t := temple.NewTemplates("app/views", temple.DevsMode)


## License

MIT
