package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	"github.com/go-msvc/errors"
)

type App interface {
	RegisterFunc(name string, appFunc AppFunc)
	Load(filename string) error
	GetItem(id string) (AppItem, bool)
}

type AppFunc func(ctx context.Context, args map[string]interface{}) error

func New() App {
	return &app{
		funcs: map[string]AppFunc{},
		items: map[string]AppItem{},
	}
}

type app struct {
	funcs map[string]AppFunc
	items map[string]AppItem
}

func (app *app) RegisterFunc(name string, appFunc AppFunc) {
	if _, ok := app.funcs[name]; ok {
		panic(fmt.Sprintf("register func %s() already registered", name))
	}
	if appFunc == nil {
		panic(fmt.Sprintf("register func %s() is nil", name))
	}
	app.funcs[name] = appFunc
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
			return errors.Errorf("missing/invalid item id \"%s\" (expect lower alnum with dashes, e.g. \"my-item1-loader\")", id)
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

type CtxSession struct{}
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
