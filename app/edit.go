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
	Title      Caption  `json:"title"`
	GetActions *Actions `json:"get_actions" doc:"Actions to execute to get the item. It must have an item that sets \"Item\"."`
	UpdActions *Actions `json:"upd_actions" doc:"Actions to execute to update the item."`
	// Options    ListOptions `json:"options" doc:"Options to manipulate the display and behavior of the list"`
	// Operations []menuItem  `json:"operations"`
}

func (edit edit) Validate(app App) error {
	if err := edit.Title.Validate(false); err != nil {
		return errors.Wrapf(err, "invalid title")
	}
	if edit.GetActions == nil {
		return errors.Errorf("missing get_actions")
	}
	if err := edit.GetActions.Validate(app); err != nil {
		return errors.Wrapf(err, "invalid get_actions")
	}
	if edit.UpdActions == nil {
		return errors.Errorf("missing upd_actions")
	}
	if err := edit.UpdActions.Validate(app); err != nil {
		return errors.Wrapf(err, "invalid upd_actions")
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
	// session := ctx.Value(CtxSession{}).(*sessions.Session)
	// renderedName := prompt.Name.Rendered(sessionData(session))
	// if !fieldNameRegex.MatchString(renderedName) {
	// 	return "", errors.Errorf("prompt invalid name(%s).rendered->\"%s\"", prompt.Name.UnparsedTemplate, renderedName)
	// }

	// log.Debugf("Set %s=\"%s\"", renderedName, submittedValueList[0])
	// session.Values[renderedName] = submittedValueList[0]

	//process next steps to return nextId or error
	return "", errors.Errorf("NYI") //edit.Next.Execute(ctx)
	//need next for saved and next for cancel==back...
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
