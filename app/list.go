package app

import (
	"context"
	"encoding/json"
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
	Operations []menuItem  `json:"operations"`
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
	Columns    []ListColumn `json:"columns"`
	ItemSet    string       `json:"item_set" doc:"When select, store item column values in this name"`
	ItemNext   fileItemNext `json:"item_next"`
	ShowFilter bool         `json:"show_filter"`
	SortFields []string     `json:"sort_fields"`
	Limit      int          `json:"limit"`
}

func (o ListOptions) Validate() error {
	if len(o.Columns) < 1 {
		return errors.Errorf("missing columns")
	}
	for colIndex, col := range o.Columns {
		if err := col.Validate(); err != nil {
			return errors.Wrapf(err, "invalid column[%d]", colIndex)
		}
	}
	if o.Limit < 0 {
		return errors.Errorf("limit:%d is negative", o.Limit)
	}
	log.Debugf("Validated %T: %+v", o, o)
	return nil
}

type ListColumn struct {
	Header Caption `json:"header" doc:"Template to construct the column header to display above the column, based on session data. May be blank."`
	Value  Caption `json:"value" doc:"Template to construct the column value for this item, based on item data. Must be specified."`
}

func (lc ListColumn) Validate() error {
	if err := lc.Header.Validate(false); err != nil {
		return errors.Wrapf(err, "missing/invalid header")
	}
	if err := lc.Value.Validate(false); err != nil {
		return errors.Wrapf(err, "missing/invalid value")
	}
	return nil
}

func (list list) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	lang := ctx.Value(CtxLang{}).(string)
	session := ctx.Value(CtxSession{}).(*sessions.Session)

	//clear items and then call actions to generate fresh list of items
	delete(session.Values, "Items")
	if err := list.GetItems.Execute(ctx); err != nil {
		return nil, errors.Wrapf(err, "failed to get items")
	}
	//items must be an array of structs or map[string]interface{}
	columnList, ok := session.Values["Items"].(ColumnList)
	if !ok {
		return nil, errors.Errorf("Items (%T) not ColumnList", session.Values["Items"])
	}

	//start prepare the template data so we can add info
	//about columns, items and operations below
	pageData := PageData{
		Links: map[string]fileItemNext{},
		Data:  nil,
	}
	title, err := list.Title.Render(lang, sessionData(session))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render title")
	}
	listTmplData := tmplDataForList{
		Title:      title,
		Items:      nil,
		Operations: []tmplDataForListOperation{},
		Columns:    []tmplDataForListColumn{},
	}

	//describe columns to be displayed
	for colIndex, col := range list.Options.Columns {
		header, err := col.Header.Render(lang, sessionData(session))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render column[%d] header", colIndex)
		}
		listTmplData.Columns = append(listTmplData.Columns, tmplDataForListColumn{
			Header: header,
		})
	} //for each column

	log.Debugf("%d items to render", len(columnList.Items))
	//add list items
	sessionItems := map[string]ColumnItem{}
	for itemIndex, item := range columnList.Items {
		uuid := uuid.New().String()
		itemData := tmplDataForListItem{
			ColumnValues: []string{},
			NextUUID:     uuid,
		}

		//render each column value
		for colIndex, col := range list.Options.Columns {
			caption, err := col.Value.Render(lang, item)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to render item[%d] caption for col[%d]", itemIndex, colIndex)
			}
			itemData.ColumnValues = append(itemData.ColumnValues, caption)
		}
		sessionItems[uuid] = item
		log.Debugf("  item[%d]: %+v -> %+v -> %s", itemIndex, item, itemData.ColumnValues, uuid)

		//next is the same for all item except it sets the selected item value as well
		pageData.Links[uuid] = append(fileItemNext{
			fileItemNextStep{Set: &fileItemSet{
				Name: ConfiguredTemplate{UnparsedTemplate: list.Options.ItemSet},
				//Value: ConfiguredTemplate{UnparsedTemplate: fmt.Sprintf("{{index .Items \"%s\"}}", uuid)}}},
				ValueStr: "Items[" + uuid + "]",
			}}}, list.Options.ItemNext...)

		listTmplData.Items = append(listTmplData.Items, itemData)
	}
	session.Values["Items"] = sessionItems

	//add list operations
	for _, oper := range list.Operations {
		caption, err := oper.Caption.Render(lang, sessionData(session))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render operation caption")
		}
		uuid := uuid.New().String()
		pageData.Links[uuid] = oper.Next
		operTmpl := tmplDataForListOperation{
			Caption:  caption,
			NextUUID: uuid,
		}
		listTmplData.Operations = append(listTmplData.Operations, operTmpl)
		log.Debugf("Added operation: %+v", operTmpl)
	}

	{
		log.Debugf("listTmplData: %+v", listTmplData)
		j, _ := json.Marshal(listTmplData)
		log.Debugf("listTmplData: %s", string(j))
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
	Columns    []tmplDataForListColumn
	Items      []tmplDataForListItem
	Operations []tmplDataForListOperation
}

type tmplDataForListColumn struct {
	Header string
}

type tmplDataForListItem struct {
	ColumnValues []string
	NextUUID     string
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
	Items []ColumnItem
}

type ColumnItem map[string]interface{}
