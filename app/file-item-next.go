package app

import (
	"context"

	"github.com/go-msvc/errors"
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
	for stepIndex, step := range next {
		if step.Set != nil {

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

type fileItemSet struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (set fileItemSet) Validate() error {
	if set.Name == "" {
		return errors.Errorf("missing name")
	}
	if set.Value == "" {
		return errors.Errorf("missing value")
	}
	return nil
}
