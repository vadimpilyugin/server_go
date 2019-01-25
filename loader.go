package main

import (
	"debug_print_go"
	"io/ioutil"
	"path"
)

var Resources map[string]string

func init() {
	Resources = make(map[string]string)
	dirs := []string{
		"static/icons/",
		"static/js/",
		"static/css/",
		"static/html/",
	}
	for _, basePath := range dirs {
		fillFromDir(basePath)
	}
}

func fillFromDir(basePath string) {
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		printer.Fatal(err)
	}

	for _, f := range files {
		Resources["/__"+f.Name()] = path.Join(basePath, f.Name())
	}
}
