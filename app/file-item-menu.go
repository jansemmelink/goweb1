package app

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type fileItemMenu struct {
	Title Caption            `json:"title"`
	Items []fileItemMenuItem `json:"items"`
}

func (menu fileItemMenu) Validate() error {
	if err := menu.Title.Validate(); err != nil {
		return errors.Wrapf(err, "invalid title")
	}
	if len(menu.Items) == 0 {
		return errors.Errorf("missing items")
	}
	for itemIndex, item := range menu.Items {
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "invalid item[%d]", itemIndex)
		}
	}
	return nil
}

//todo: make menu with sub menus that can expand and collapse with headings
//the rendering template can display it any way required...

func (menu fileItemMenu) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	//for each menu item, generate a uuid stored in the session
	//which are used in the URL and avoids a user to manipulate
	//the app by changing URLs
	pageData := PageData{
		Links: map[string]fileItemNext{},
		Data:  nil,
	}
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	title, err := menu.Title.Render(session)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render title")
	}
	menuTmplData := tmplDataForMenu{
		Title: title, //todo: i18n and substitute...
		Items: []tmplDataForMenuItem{},
	}
	for _, item := range menu.Items {
		caption, err := item.Caption.Render(session)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render item caption")
		}
		uuid := uuid.New().String()
		pageData.Links[uuid] = item.Next
		menuTmplData.Items = append(menuTmplData.Items,
			tmplDataForMenuItem{
				Caption:  caption, //todo: i18n and substitute...
				NextUUID: uuid,
			})
	}

	tmplData := TmplData{
		NavBar: TmplNavBar{
			Email: "a@b.c", //todo...
		},
		Body: menuTmplData,
	}
	if err := menuTmpl.ExecuteTemplate(buffer, "page", tmplData); err != nil {
		return nil, errors.Wrapf(err, "failed to exec menu template")
	}
	return &pageData, nil
} //menu.Render()

func (menu fileItemMenu) Process(ctx context.Context, httpReq *http.Request) error {
	return errors.Errorf("Menu does not handle POST")
}

type fileItemMenuItem struct {
	Caption Caption      `json:"caption"`
	Next    fileItemNext `json:"next"`
}

func (item fileItemMenuItem) Validate() error {
	if err := item.Caption.Validate(); err != nil {
		return errors.Wrapf(err, "invalid caption")
	}
	if err := item.Next.Validate(); err != nil {
		return errors.Wrapf(err, "invalid next")
	}
	return nil
} //fileItemMenuItem.Validate()

type tmplDataForMenu struct {
	Title string
	Items []tmplDataForMenuItem
}

type tmplDataForMenuItem struct {
	Caption  string //displayed to user
	NextUUID string //uuid value used in sessionDataForMenu.Items[<uuid>]
}

// generic
type TmplData struct {
	NavBar TmplNavBar
	Body   interface{} //depends on the page
}
type TmplNavBar struct {
	//Items...
	//User  *TmplUser
	Email string
}

var menuTmpl *template.Template

func init() {
	var err error
	menuTmpl, err = LoadPageTemplates([]string{"menu"})
	if err != nil {
		panic(fmt.Sprintf("failed to load menu template: %+v", err))
	}
} //init()
