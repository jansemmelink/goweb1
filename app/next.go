package app

import (
	"context"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

type fileItemNext []fileItemNextStep

func (next fileItemNext) Validate() error {
	if len(next) == 0 {
		return errors.Errorf("missing next")
	}
	for stepIndex, step := range next {
		if err := step.Validate(); err != nil {
			return errors.Wrapf(err, "invalid next step[%d]", stepIndex)
		}
		if step.Item != nil && stepIndex != len(next)-1 {
			return errors.Errorf("step[%d] is next, only allowed as last step", stepIndex)
		}
	}
	return nil
}

func (next fileItemNext) Execute(ctx context.Context) (nextItemId string, err error) {
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	for stepIndex, step := range next {
		if step.Set != nil {
			name := step.Set.Name.Rendered(sessionData(session))
			if !fieldNameRegex.MatchString(name) {
				return "", errors.Wrapf(err, "step[%d].name.render(%s) invalid", stepIndex, name)
			}
			value := step.Set.Value.Rendered(sessionData(session))
			log.Debugf("SET(%s)=\"%s\"", name, value)
			session.Values[name] = value
			continue
		}
		if step.Item != nil {
			return string(*step.Item), nil
		}
		return "", errors.Errorf("unhandled next step[%d] %T", stepIndex, step)
	}
	return "", nil
}

type fileItemNextStep struct {
	Item *fileItemNextItem `json:"item,omitempty" doc:"Value is next item id"`
	Set  *fileItemSet      `json:"set,omitempty"`
}

type fileItemNextItem string

func (next fileItemNextStep) Validate() error {
	count := 0
	if next.Item != nil {
		if *next.Item == "" {
			return errors.Errorf("missing next item id")
		}
		count++
	}
	if next.Set != nil {
		if err := next.Set.Validate(); err != nil {
			return errors.Wrapf(err, "invalid set")
		}
		count++
	}
	if count == 0 {
		return errors.Errorf("missing id|set")
	}
	if count > 1 {
		return errors.Errorf("%d instead of 1 of id|set", count)
	}
	return nil
}
