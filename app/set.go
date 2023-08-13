package app

import "github.com/go-msvc/errors"

// todo: action and next must both use the same way to set fields...
//
//	only diff is that next also sets the next item id in the end, could even be from
//	as specified field name, like "Next"
type fileItemSet struct {
	Name ConfiguredTemplate `json:"name"`
	//Value ConfiguredTemplate `json:"value"`
	ValueStr string `json:"value"`
}

func (set fileItemSet) Validate() error {
	if err := set.Name.Validate(); err != nil {
		return errors.Wrapf(err, "invalid name")
	}
	// if err := set.Value.Validate(); err != nil {
	// 	return errors.Wrapf(err, "invalid value")
	// }
	return nil
}
