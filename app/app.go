package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/go-msvc/errors"
)

type App interface {
	Load(filename string) error
	GetItem(id string) (AppItem, bool)
}

func New() App {
	return &app{
		items: map[string]AppItem{},
	}
}

type app struct {
	items map[string]AppItem
}

func (app *app) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer f.Close()

	fileItems := map[string]fileItem{}
	if err := json.NewDecoder(f).Decode(&fileItems); err != nil {
		return errors.Wrapf(err, "failed to read items from JSON file %s", filename)
	}
	for id, item := range fileItems {
		if !itemIdRegex.MatchString(id) {
			return errors.Errorf("invalid item id \"%s\" (expect lower alnum with dashes, e.g. \"my-item1-loader\")", id)
		}
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "invalid item \"%s\"", id)
		}
		//todo: validate etc...
		app.items[id] = item
	}
	return nil
}

func (app *app) GetItem(id string) (AppItem, bool) {
	item, ok := app.items[id]
	if !ok {
		return nil, false
	}
	return item, true
} //app.GetItem()

type AppItem interface {
	Render(ctx context.Context, buffer io.Writer) (
		pageData *PageData,
		err error)

	//Process is called on method POST
	Process(ctx context.Context, httpReq *http.Request) error
}

type CtxPageData struct{}

type PageData struct {
	Links map[string]fileItemNext //key is uuid for mapping URL ?next=<uuid> -> next steps
	Data  interface{}             //anything else the page needs in Process()
}

// func (p PageData) Value() {
// ...
// }

const itemIdPattern = `[a-z]([a-z0-9-]*[a-z0-9])*`

var itemIdRegex = regexp.MustCompile("^" + itemIdPattern + "$")
