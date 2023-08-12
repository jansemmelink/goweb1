package app

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

// key is language code (any string, "" for default) and value is text
// must have at least key = "" for default value
type Caption map[string]ConfiguredTemplate

func (c *Caption) Validate() error {
	hasDefault := false
	for lang := range *c {
		if lang == "" {
			hasDefault = true
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

func (ct ConfiguredTemplate) Rendered(ctx context.Context) string {
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	buf := bytes.NewBuffer(nil)
	if ct.tmpl == nil {
		var err error
		ct.tmpl, err = template.New("tmpl").Parse(ct.UnparsedTemplate)
		if err != nil {
			panic(fmt.Sprintf("tmpl is nil: failed to parse: %+v", err))
		}
	}
	if err := ct.tmpl.Execute(buf, sessionData(session)); err != nil {
		log.Errorf("failed to render: %+v", err) //need context...
		return ""
	}
	return string(buf.Bytes())
}

type CtxLang struct{}

func (c Caption) Render(s *sessions.Session) (string, error) {
	lang, ok := s.Values["lang"].(string)
	if !ok || len(lang) != 2 {
		lang = ""
	}
	ct, ok := c[lang]
	if !ok && lang != "" {
		ct, ok = c[""]
		if !ok {
			return "", errors.Errorf("missing caption for lang:\"%s\"", lang)
		}
	}

	//todo: this must be done once and stored!
	buffer := bytes.NewBuffer(nil)
	if ct.tmpl == nil {
		var err error
		ct.tmpl, err = template.New("tmpl").Parse(ct.UnparsedTemplate)
		if err != nil {
			panic(fmt.Sprintf("tmpl is nil: failed to parse: %+v", err))
		}
	}
	if err := ct.tmpl.Execute(buffer, s.Values); err != nil {
		return "", errors.Wrapf(err, "failed to execute template")
	}
	return buffer.String(), nil
} //Caption.Render()
