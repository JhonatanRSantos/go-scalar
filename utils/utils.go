package utils

import (
	"io/fs"
	"text/template"
)

func ParseTemplateFromFS(fsys fs.FS, patterns ...string) (*template.Template, error) {
	tmp, err := template.ParseFS(fsys, patterns...)
	if err != nil {
		return nil, err
	}
	return tmp, nil
}
