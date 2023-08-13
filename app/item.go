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
		nextItemId string, //only for redirect
		pageData *PageData, //only when ready to display
		err error)

	//Process is called on method POST
	//return next item or error
	//next item is required, return own if need to stay put
	Process(ctx context.Context, httpReq *http.Request) (string, error)
}

type item struct {
	//optional
	OnEnter *Actions `json:"on_enter_actions,omitempty" doc:"Optional list of actions to take when entering the item"`

	//union: one of the following is required
	Menu   *menu         `json:"menu"`
	Prompt *prompt       `json:"prompt"`
	List   *list         `json:"list"`
	Edit   *edit         `json:"edit"`
	Next   *fileItemNext `json:"next" doc:"A series of actions to get to next"`
}

func (i item) Validate(app App) error {
	if i.OnEnter != nil {
		if err := i.OnEnter.Validate(app); err != nil {
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
	if i.List != nil {
		if err := i.List.Validate(app); err != nil {
			return errors.Wrapf(err, "invalid list")
		}
		count++
	}
	if i.Edit != nil {
		if err := i.Edit.Validate(app); err != nil {
			return errors.Wrapf(err, "invalid edit")
		}
		count++
	}
	if i.Next != nil {
		if err := i.Next.Validate(); err != nil {
			return errors.Wrapf(err, "invalid next")
		}
		count++
	}
	if count == 0 {
		return errors.Errorf("missing menu|prompt|list|edit|...")
	}
	if count > 1 {
		return errors.Errorf("has %d instead of 1 of menu|prompt|list|...", count)
	}
	return nil
} //item.Validate()

func (item item) OnEnterActions() *Actions {
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

func (item item) Render(ctx context.Context, buffer io.Writer) (string, *PageData, error) {
	if item.Menu != nil {
		if pageData, err := item.Menu.Render(ctx, buffer); err != nil {
			return "", nil, err
		} else {
			return "", pageData, nil
		}
	}
	if item.Prompt != nil {
		if pageData, err := item.Prompt.Render(ctx, buffer); err != nil {
			return "", nil, err
		} else {
			return "", pageData, nil
		}
	}
	if item.List != nil {
		if pageData, err := item.List.Render(ctx, buffer); err != nil {
			return "", nil, err
		} else {
			return "", pageData, nil
		}
	}
	if item.Edit != nil {
		if pageData, err := item.Edit.Render(ctx, buffer); err != nil {
			return "", nil, err
		} else {
			return "", pageData, nil
		}
	}

	if item.Next != nil {
		//next sets the next item which should be rendered
		nextItemId, err := item.Next.Execute(ctx)
		if err != nil {
			return "", nil, errors.Wrapf(err, "failed to execute action to determine next item")
		}
		return nextItemId, nil, nil //redirect
	}
	return "", nil, errors.Errorf("cannot render %+v", item)
}

func (item item) Process(ctx context.Context, httpReq *http.Request) (string, error) {
	//menu does not process a http POST
	// if item.Menu != nil {
	// 	return item.Menu.Process(ctx, httpReq)
	// }

	//todo: maybe list will post search filter?
	// if item.List != nil {
	// 	return item.List.Process(ctx, httpReq)
	// }
	if item.Prompt != nil {
		return item.Prompt.Process(ctx, httpReq)
	}
	if item.Edit != nil {
		return item.Edit.Process(ctx, httpReq)
	}
	return "", errors.Errorf("cannot process %+v", item)
}
