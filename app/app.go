package app

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

type App interface {
	//RegisterFunc:
	//	appFunc must be a func taking args (context.Context, optional request)
	//			and respond with (optional response, error)
	RegisterFunc(name string, appFunc interface{}) error
	FuncByName(name string) (*AppFunc, bool)
	Load(filename string) error
	GetItem(id string) (AppItem, bool)
}

type AppFunc struct {
	reqType   reflect.Type
	resType   reflect.Type
	funcValue reflect.Value
}

func init() {
	//register types stored in session data, else session save will fail
	gob.Register(PageData{})
	gob.Register(map[string]interface{}{})
	gob.Register(ColumnList{})
	gob.Register(ColumnItem{})
	gob.Register(map[string]ColumnItem{})
}

func New() App {
	return &app{
		funcs: map[string]*AppFunc{},
		items: map[string]AppItem{},
	}
}

type app struct {
	funcs map[string]*AppFunc
	items map[string]AppItem
}

func (app *app) MustRegisterFunc(name string, appFunc interface{}) {
	if err := app.RegisterFunc(name, appFunc); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
} //app.MustRegisterFunc()

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
var errorType = reflect.TypeOf((*error)(nil)).Elem()

func (app *app) RegisterFunc(name string, appFunc interface{}) error {
	if _, ok := app.funcs[name]; ok {
		return errors.Errorf("register func %s() already registered", name)
	}
	if appFunc == nil {
		return errors.Errorf("register func %s() is nil", name)
	}
	funcType := reflect.TypeOf(appFunc)
	if funcType.NumIn() < 1 || funcType.In(0) != contextType {
		return errors.Errorf("%s() first arg %v is not %v", name, funcType.In(0), contextType)
	}
	if funcType.NumIn() > 2 {
		return errors.Errorf("%s() takes more args than only (ctx, req)", name)
	}
	if funcType.NumOut() < 1 || funcType.Out(funcType.NumOut()-1) != errorType {
		return errors.Errorf("%s() last result %v is not %v", name, funcType.Out(funcType.NumOut()-1), errorType)
	}
	if funcType.NumOut() > 2 {
		return errors.Errorf("%s() returns more results than only (res, error)", name)
	}

	info := &AppFunc{
		funcValue: reflect.ValueOf(appFunc),
	}
	if funcType.NumIn() == 2 {
		info.reqType = funcType.In(1)
	}
	if funcType.NumOut() == 2 {
		info.resType = funcType.Out(0)
	}
	app.funcs[name] = info
	log.Debugf("Registered func %s(%v) -> %v", name, info.reqType, info.resType)
	return nil
} //app.RegisterFunc()

func (app app) FuncByName(name string) (*AppFunc, bool) {
	fnc, ok := app.funcs[name]
	if ok {
		return fnc, true
	}
	return nil, false
} //app.FuncByName()

func (app *app) Load(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer f.Close()

	fileItems := map[string]item{}
	if err := json.NewDecoder(f).Decode(&fileItems); err != nil {
		return errors.Wrapf(err, "failed to read items from JSON file %s", filename)
	}
	for id, item := range fileItems {
		if !itemIdRegex.MatchString(id) {
			return errors.Errorf("missing/invalid item id \"%s\" (expect lower alnum with dashes, e.g. \"my-item1-loader\")", id)
		}
		if err := item.Validate(app); err != nil {
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
