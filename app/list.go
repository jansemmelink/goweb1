package app

import (
	"context"
	"fmt"
	"html/template"
	"io"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type list struct {
	Title      Caption     `json:"title"`
	GetItems   *Actions    `json:"get_items" doc:"Actions to execute to make items. It must have an item that sets \"Items\"."`
	Options    ListOptions `json:"options" doc:"Options to manipulate the display and behavior of the list"`
	Operations []menuItem  `json:"list_operations"`
}

func (list list) Validate(app App) error {
	if err := list.Title.Validate(false); err != nil {
		return errors.Wrapf(err, "invalid title")
	}
	if list.GetItems == nil {
		return errors.Errorf("missing get_items")
	}
	if err := list.GetItems.Validate(app); err != nil {
		return errors.Wrapf(err, "invalid get_items")
	}
	if err := list.Options.Validate(); err != nil {
		return errors.Wrapf(err, "invalid options")
	}
	for operIndex, oper := range list.Operations {
		if err := oper.Validate(); err != nil {
			return errors.Wrapf(err, "invalid operation[%d]", operIndex)
		}
	}
	return nil
} //list.Validate()

type ListOptions struct {
	Caption    Caption  `json:"item_caption"`
	ShowFilter bool     `json:"show_filter"`
	SortFields []string `json:"sort_fields"`
	Limit      int      `json:"limit"`
}

func (o ListOptions) Validate() error {
	if err := o.Caption.Validate(false); err != nil {
		return errors.Wrapf(err, "invalid/blank item_caption")
	}
	if o.Limit < 0 {
		return errors.Errorf("limit:%d is negative", o.Limit)
	}
	log.Debugf("Validated %T: %+v", o, o)
	return nil
}

func (list list) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	pageData := PageData{
		Links: map[string]fileItemNext{},
		Data:  nil,
	}

	lang := ctx.Value(CtxLang{}).(string)
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	title, err := list.Title.Render(lang, sessionData(session))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render title")
	}
	listTmplData := tmplDataForList{
		Title:      title,
		Items:      nil,
		Operations: []tmplDataForListOperation{},
	}

	//clear items and then call actions to generate fresh list of items
	delete(session.Values, "Items")
	if err := list.GetItems.Execute(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to get items")
	}
	log.Debugf("after action: %+v", sessionData(session))

	//items must be an array of structs or map[string]interface{}
	columnList, ok := session.Values["Items"].(ColumnList)
	if !ok {
		return nil, errors.Errorf("Items (%T) not ColumnList", session.Values["Items"])
	}

	log.Debugf("%d columns with %d items to render", len(columnList.Columns), len(columnList.Items))
	//add list items
	for itemIndex, item := range columnList.Items {
		//item caption is templated from item data
		caption, err := list.Options.Caption.Render(lang, item)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render item caption")
		}
		uuid := uuid.New().String()
		log.Debugf("  item[%d]: %+v -> caption:\"%s\" -> %s", itemIndex, item, caption, uuid)

		//next is the same for all item except it sets the item identifier
		//to be used by other steps in next or the next page to be displayed
		//todo:...
		// pageData.Links[uuid] = .Next

		listTmplData.Items = append(listTmplData.Items,
			tmplDataForListItem{
				Caption:  caption,
				NextUUID: uuid,
			})
	}

	//add list operations
	for _, oper := range list.Operations {
		caption, err := oper.Caption.Render(lang, sessionData(session))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render operation caption")
		}
		uuid := uuid.New().String()
		pageData.Links[uuid] = oper.Next
		listTmplData.Operations = append(listTmplData.Operations,
			tmplDataForListOperation{
				Caption:  caption,
				NextUUID: uuid,
			})
	}

	tmplData := TmplData{
		NavBar: TmplNavBar{
			Email: "a@b.c", //todo...
		},
		Body: listTmplData,
	}
	if err := listTmpl.ExecuteTemplate(buffer, "page", tmplData); err != nil {
		return nil, errors.Wrapf(err, "failed to exec list template")
	}
	return &pageData, nil
}

type tmplDataForList struct {
	Title      string
	Items      []tmplDataForListItem
	Operations []tmplDataForListOperation
}

type tmplDataForListItem struct {
	Caption  string
	NextUUID string
}

type tmplDataForListOperation struct {
	Caption  string
	NextUUID string
}

var listTmpl *template.Template

func init() {
	var err error
	listTmpl, err = LoadPageTemplates([]string{"list"})
	if err != nil {
		panic(fmt.Sprintf("failed to load list template: %+v", err))
	}
} //init()

// list action function that sets "Items" must return ColumnList
type ColumnList struct {
	Columns []string
	Items   []ColumnItem
}

type ColumnItem map[string]interface{}
