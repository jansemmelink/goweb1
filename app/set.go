package app

import "github.com/go-msvc/errors"

type fileItemSet struct {
	Name  ConfiguredTemplate `json:"name"`
	Value ConfiguredTemplate `json:"value"`
}

func (set fileItemSet) Validate() error {
	if err := set.Name.Validate(); err != nil {
		return errors.Wrapf(err, "invalid name")
	}
	if err := set.Value.Validate(); err != nil {
		return errors.Wrapf(err, "invalid value")
	}
	return nil
}
