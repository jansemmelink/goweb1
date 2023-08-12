package app

import (
	"html/template"

	"github.com/go-msvc/errors"
)

func LoadPageTemplates(templateNames []string) (*template.Template, error) {
	templateNames = append(templateNames, "page")
	return LoadTemplates(templateNames)
}

func LoadTemplates(templateNames []string) (*template.Template, error) {
	templateFileNames := []string{}
	for _, n := range templateNames {
		templateFileNames = append(templateFileNames, "./templates/"+n+".tmpl")
	}
	t, err := template.ParseFiles(templateFileNames...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load template files: %v", templateFileNames)
	}
	log.Debugf("loaded %v", templateFileNames)
	return t, nil
}
