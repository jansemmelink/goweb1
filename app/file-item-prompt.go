package app

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"regexp"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

type fileItemPrompt struct {
	Caption Caption            `json:"caption"`
	Name    ConfiguredTemplate `json:"name" doc:"Template to construct name where value will be stored. Result must be CamelCase."`
	Next    fileItemNext       `json:"next"`
	//todo: validation rules/function with args
}

func (prompt *fileItemPrompt) Validate() error {
	if err := prompt.Caption.Validate(); err != nil {
		return errors.Wrapf(err, "invalid caption")
	}
	if err := prompt.Name.Validate(); err != nil {
		return errors.Errorf("invalid name")
	}
	if err := prompt.Next.Validate(); err != nil {
		return errors.Wrapf(err, "invalid next")
	}
	return nil
}

func (prompt fileItemPrompt) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	caption, err := prompt.Caption.Render(session)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render caption")
	}
	promptTmplData := tmplDataForPrompt{
		Caption: caption,
	}
	tmplData := TmplData{
		NavBar: TmplNavBar{
			Email: "a@b.c", //todo... phone/user name?
		},
		Body: promptTmplData,
	}
	if err := promptTmpl.ExecuteTemplate(buffer, "page", tmplData); err != nil {
		return nil, errors.Wrapf(err, "failed to exec prompt template")
	}
	return nil, nil
} //fileItemPrompt.Render()

func (prompt fileItemPrompt) Process(ctx context.Context, httpReq *http.Request) (string, error) {
	httpReq.ParseForm()
	submittedValueList, ok := httpReq.Form["SubmittedValue"] //"SubmittedValue" is used in prompt.tmpl...
	if !ok {
		return "", errors.Errorf("form did not post expected value")
	}
	if len(submittedValueList) != 1 {
		return "", errors.Errorf("form post len(SubmittedValue)=%d", len(submittedValueList))
	}
	renderedName := prompt.Name.Rendered(ctx)
	if !fieldNameRegex.MatchString(renderedName) {
		return "", errors.Errorf("prompt invalid name(%s).rendered->\"%s\"", prompt.Name.UnparsedTemplate, renderedName)
	}

	log.Debugf("Set %s=\"%s\"", renderedName, submittedValueList[0])
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	session.Values[renderedName] = submittedValueList[0]

	//process next steps to return nextId or error
	return prompt.Next.Execute(ctx)
} //fileItemPrompt.Process()

type tmplDataForPrompt struct {
	Caption string
	Name    string
}

const fieldNamePattern = `[A-Z][a-zA-Z0-9]*` //CamelCase

var fieldNameRegex = regexp.MustCompile("^" + fieldNamePattern + "$")

var promptTmpl *template.Template

func init() {
	var err error
	promptTmpl, err = LoadPageTemplates([]string{"prompt"})
	if err != nil {
		panic(fmt.Sprintf("failed to load prompt template: %+v", err))
	}
} //init()

func sessionData(s *sessions.Session) map[string]interface{} {
	data := map[string]interface{}{}
	for n, v := range s.Values {
		name := fmt.Sprintf("%v", n)
		if fieldNameRegex.MatchString(name) { //this is expensive...
			data[name] = v
		}
	}
	return data
}
