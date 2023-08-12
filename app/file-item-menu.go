package app

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
)

type fileItemMenu struct {
	Title string             `json:"title"`
	Items []fileItemMenuItem `json:"items"`
}

func (menu fileItemMenu) Validate() error {
	if menu.Title == "" {
		return errors.Errorf("missing title")
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
	menuTmplData := tmplDataForMenu{
		Title: menu.Title, //todo: i18n and substitute...
		Items: []tmplDataForMenuItem{},
	}
	for _, item := range menu.Items {
		uuid := uuid.New().String()
		pageData.Links[uuid] = item.Next
		menuTmplData.Items = append(menuTmplData.Items,
			tmplDataForMenuItem{
				Caption:  item.Caption, //todo: i18n and substitute...
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
	Caption string       `json:"caption"`
	Next    fileItemNext `json:"next"`
}

func (item fileItemMenuItem) Validate() error {
	if item.Caption == "" {
		return errors.Errorf("missing caption")
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
