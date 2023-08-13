package app

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"

	"github.com/go-msvc/errors"
)

// key is language code (any string, "" for default) and value is text
// must have at least key = "" for default value
type Caption map[string]ConfiguredTemplate

func (c *Caption) Validate(allowBlank bool) error {
	hasDefault := false
	for lang, ct := range *c {
		if lang == "" {
			hasDefault = true
		}
		if !allowBlank && ct.UnparsedTemplate == "" {
			return errors.Errorf("blank template not allowed")
		}
	}
	if !hasDefault {
		return errors.Errorf("missing default entry")
	}
	//note: blank text is allowed
	return nil
}

type ConfiguredTemplate struct {
	tmpl             *template.Template
	UnparsedTemplate string
}

func (lc *ConfiguredTemplate) UnmarshalJSON(data []byte) error {
	//expect string which is a template
	lc.UnparsedTemplate = strings.Trim(string(data), "\"")
	var err error
	lc.tmpl, err = template.New("caption").Parse(lc.UnparsedTemplate)
	if err != nil {
		return errors.Wrapf(err, "invalid template(%s)", string(data))
	}
	return nil
}

func (ct ConfiguredTemplate) Validate() error {
	if ct.tmpl == nil {
		return errors.Errorf("missing template")
	}
	return nil
}

func (ct ConfiguredTemplate) Rendered(data interface{}) string {
	buf := bytes.NewBuffer(nil)
	if ct.tmpl == nil {
		var err error
		ct.tmpl, err = template.New("tmpl").Parse(ct.UnparsedTemplate)
		if err != nil {
			panic(fmt.Sprintf("failed template.parse(%s): %+v", ct.UnparsedTemplate, err))
		}
	}
	if err := ct.tmpl.Execute(buf, data); err != nil {
		log.Errorf("failed to render: %+v", err) //need context...
		return ""
	}
	//log.Debugf("tmpl(%s) with (%T)%+v -> \"%s\"", ct.UnparsedTemplate, data, data, string(buf.Bytes()))
	return string(buf.Bytes())
}

type CtxLang struct{}

func (c Caption) Render(lang string, data interface{}) (string, error) {
	ct, ok := c[lang]
	if !ok && lang != "" {
		ct, ok = c[""]
		if !ok {
			return "", errors.Errorf("missing caption for lang:\"%s\" and no default", lang)
		}
		lang = ""
	}

	//todo: this must be done once and stored!
	return ct.Rendered(data), nil
	// buffer := bytes.NewBuffer(nil)
	// if ct.tmpl == nil {
	// 	var err error
	// 	ct.tmpl, err = template.New("tmpl").Parse(ct.UnparsedTemplate)
	// 	if err != nil {
	// 		panic(fmt.Sprintf("tmpl is nil: failed to parse: %+v", err))
	// 	}
	// }
	// if err := ct.tmpl.Execute(buffer, data); err != nil {
	// 	return "", errors.Wrapf(err, "failed to execute template")
	// }
	// log.Debugf("Render(%s: %s) with (%T)%+v -> \"%s\"",
	// 	lang, ct.UnparsedTemplate, data, data, buffer.String())
	// return buffer.String(), nil
} //Caption.Render()
