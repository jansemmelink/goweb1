package app

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"reflect"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

type edit struct {
	Title       Caption  `json:"title"`
	GetActions  *Actions `json:"get_actions" doc:"Actions to execute to get the item. It must have an item that sets \"Item\"."`
	UpdFuncName string   `json:"upd_func" doc:"Func to save item"`
	updFunc     *AppFunc
	SavedNext   fileItemNext `json:"saved_next"`
}

func (edit *edit) Validate(app App) error {
	if err := edit.Title.Validate(false); err != nil {
		return errors.Wrapf(err, "invalid title")
	}
	if edit.GetActions == nil {
		return errors.Errorf("missing get_actions")
	}
	if err := edit.GetActions.Validate(app); err != nil {
		return errors.Wrapf(err, "invalid get_actions")
	}
	var ok bool
	if edit.updFunc, ok = app.FuncByName(edit.UpdFuncName); !ok {
		return errors.Errorf("missing/unknown upd_func:\"%s\"", edit.UpdFuncName)
	}
	if err := edit.SavedNext.Validate(); err != nil {
		return errors.Wrapf(err, "invalid saved_next")
	}
	return nil
} //edit.Validate()

func (edit edit) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	lang := ctx.Value(CtxLang{}).(string)
	session := ctx.Value(CtxSession{}).(*sessions.Session)

	//clear item and then call actions to fetch item
	delete(session.Values, "Item")
	if err := edit.GetActions.Execute(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to get item")
	}
	//items must be struct
	item, ok := session.Values["Item"]
	if !ok {
		return nil, errors.Errorf("Item not defined get get_actions")
	}

	//todo: need way to register custom types else they cannot be stored in profile
	//this might be expensive... not sure
	gob.Register(item)

	log.Debugf("Editor for %T", item)
	structType := reflect.TypeOf(item)
	if structType.Kind() != reflect.Struct {
		return nil, errors.Errorf("edit.get_actions returned %T which is not a struct", item)
	}
	structValue := reflect.ValueOf(item)

	//start prepare the template data so we can add info
	//about fields
	pageData := PageData{
		Links: map[string]fileItemNext{},
		Data:  nil,
	}
	title, err := edit.Title.Render(lang, sessionData(session))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render title")
	}
	editTmplData := tmplDataForEdit{
		Title:  title,
		Fields: []tmplDataForEditField{},
	}
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		label := f.Tag.Get("label")
		if label == "" {
			label = f.Name
		}
		fieldData := tmplDataForEditField{
			Label: label,
			Name:  f.Name,
			Value: fmt.Sprintf("%v", structValue.Field(i).Interface()),
		}
		editTmplData.Fields = append(editTmplData.Fields, fieldData)
	}

	// //add list items
	// sessionItems := map[string]ColumnItem{}
	// for itemIndex, item := range columnList.Items {

	// 	//render each column value
	// 	for colIndex, col := range list.Options.Columns {
	// 		caption, err := col.Value.Render(lang, item)
	// 		if err != nil {
	// 			return nil, errors.Wrapf(err, "failed to render item[%d] caption for col[%d]", itemIndex, colIndex)
	// 		}
	// 		itemData.ColumnValues = append(itemData.ColumnValues, caption)
	// 	}
	// 	sessionItems[uuid] = item
	// 	log.Debugf("  item[%d]: %+v -> %+v -> %s", itemIndex, item, itemData.ColumnValues, uuid)

	// 	//next is the same for all item except it sets the selected item value as well
	// 	pageData.Links[uuid] = append(fileItemNext{
	// 		fileItemNextStep{Set: &fileItemSet{
	// 			Name: ConfiguredTemplate{UnparsedTemplate: list.Options.ItemSet},
	// 			//Value: ConfiguredTemplate{UnparsedTemplate: fmt.Sprintf("{{index .Items \"%s\"}}", uuid)}}},
	// 			ValueStr: "Items[" + uuid + "]",
	// 		}}}, list.Options.ItemNext...)

	// 	listTmplData.Items = append(listTmplData.Items, itemData)
	// }
	//session.Values["Items"] =

	// //add list operations
	// for _, oper := range list.Operations {
	// 	caption, err := oper.Caption.Render(lang, sessionData(session))
	// 	if err != nil {
	// 		return nil, errors.Wrapf(err, "failed to render operation caption")
	// 	}
	// 	uuid := uuid.New().String()
	// 	pageData.Links[uuid] = oper.Next
	// 	operTmpl := tmplDataForListOperation{
	// 		Caption:  caption,
	// 		NextUUID: uuid,
	// 	}
	// 	listTmplData.Operations = append(listTmplData.Operations, operTmpl)
	// 	log.Debugf("Added operation: %+v", operTmpl)
	// }

	{
		log.Debugf("editTmplData: %+v", editTmplData)
		j, _ := json.Marshal(editTmplData)
		log.Debugf("editTmplData: %s", string(j))
	}

	tmplData := TmplData{
		NavBar: TmplNavBar{
			Email: "a@b.c", //todo...
		},
		Body: editTmplData,
	}
	if err := editTmpl.ExecuteTemplate(buffer, "page", tmplData); err != nil {
		return nil, errors.Wrapf(err, "failed to exec edit template")
	}
	return &pageData, nil
} //edit.Render()

func (edit edit) Process(ctx context.Context, httpReq *http.Request) (string, error) {
	httpReq.ParseForm()
	log.Debugf("form data: %+v", httpReq.Form)
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	item := session.Values["Item"] //consider making this uuid so that re-submit of old form has no effect

	//apply the form values to struct fields
	structType := reflect.TypeOf(item)
	if structType.Kind() != reflect.Struct {
		return "", errors.Errorf("edit session.item %T is not a struct", item)
	}

	//make a new copy of item which we can edit
	newValuePtr := reflect.New(structType)
	for i := 0; i < structType.NumField(); i++ {
		f := structType.Field(i)
		v := httpReq.Form.Get(f.Name)
		if n, err := fmt.Sscanf(v, "%v", newValuePtr.Elem().Field(i).Addr().Interface()); err != nil || n != 1 {
			return "", errors.Wrapf(err, "failed to parse \"%s\" into %T", v, newValuePtr.Elem().Field(i).Interface())
		}
		x := newValuePtr.Elem().Field(i).Interface()
		log.Debugf("%s: \"%s\" -> (%T)%+v", f.Name, v, x, x)
	}
	item = newValuePtr.Elem().Interface()
	log.Debugf("Edited Item: (%T)%+v", item, item)

	//call update function
	results := edit.updFunc.funcValue.Call([]reflect.Value{
		reflect.ValueOf(ctx),
		reflect.ValueOf(item), //updated item
	})
	errValue := results[len(results)-1]
	if !errValue.IsNil() {
		return "", errors.Wrapf(errValue.Interface().(error), "failed to update item")
	}
	session.Values["Item"] = item
	nextItemId, err := edit.SavedNext.Execute(ctx)
	if err != nil {
		return "", errors.Errorf("failed to get next")
	}
	return nextItemId, nil
} //edit.Process()

type tmplDataForEdit struct {
	Title  string
	Fields []tmplDataForEditField
}

type tmplDataForEditField struct {
	Label string //displayed to user
	Name  string //name of value in struct
	Value string //value to put in form
}

var editTmpl *template.Template

func init() {
	var err error
	editTmpl, err = LoadPageTemplates([]string{"edit"})
	if err != nil {
		panic(fmt.Sprintf("failed to load edit template: %+v", err))
	}
} //init()
