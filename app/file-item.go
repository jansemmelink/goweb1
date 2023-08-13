package app

import (
	"context"
	"io"
	"net/http"

	"github.com/go-msvc/errors"
)

type AppItem interface {
	OnEnterActions() *Actions
	Render(ctx context.Context, buffer io.Writer) (
		pageData *PageData,
		err error)

	//Process is called on method POST
	//return next item or error
	//next item is required, return own if need to stay put
	Process(ctx context.Context, httpReq *http.Request) (string, error)
}

type fileItem struct {
	//optional
	OnEnter *Actions `json:"on_enter_actions,omitempty" doc:"Optional list of actions to take when entering the item"`

	//union: one of the following is required
	Menu   *fileItemMenu
	Prompt *fileItemPrompt
}

func (i fileItem) Validate() error {
	if i.OnEnter != nil {
		if err := i.OnEnter.Validate(); err != nil {
			return errors.Wrapf(err, "invalid on_enter")
		}
	}

	count := 0
	if i.Menu != nil {
		if err := i.Menu.Validate(); err != nil {
			return errors.Wrapf(err, "invalid menu")
		}
		count++
	}
	if i.Prompt != nil {
		if err := i.Prompt.Validate(); err != nil {
			return errors.Wrapf(err, "invalid prompt")
		}
		count++
	}
	if count == 0 {
		return errors.Errorf("missing menu|prompt|...")
	}
	if count > 1 {
		return errors.Errorf("has %d instead of 1 of menu|prompt|...", count)
	}
	return nil
} //fileItem.Validate()

func (item fileItem) OnEnterActions() *Actions {
	//do not return nil, else OnEnterActions().Execute() will fail
	//rather return an empty string
	if item.OnEnter == nil {
		log.Debugf("OnEnter = nil")
		return &Actions{list: []Action{}}
	} else {
		log.Debugf("OnEnter.list=%d", len(item.OnEnter.list))
	}
	return item.OnEnter
}

func (item fileItem) Render(ctx context.Context, buffer io.Writer) (*PageData, error) {
	if item.Menu != nil {
		return item.Menu.Render(ctx, buffer)
	}
	if item.Prompt != nil {
		return item.Prompt.Render(ctx, buffer)
	}
	return nil, errors.Errorf("cannot render %+v", item)
}

func (item fileItem) Process(ctx context.Context, httpReq *http.Request) (string, error) {
	//menu does not process a http POST
	// if item.Menu != nil {
	// 	return item.Menu.Process(ctx, httpReq)
	// }
	if item.Prompt != nil {
		return item.Prompt.Process(ctx, httpReq)
	}
	return "", errors.Errorf("cannot process %+v", item)
}
