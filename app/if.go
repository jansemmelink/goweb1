package app

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/expression"
)

type fileItemIf struct {
	Expr string `json:"expr"`
	expr expression.IExpression
	Then fileItemNext `json:"then"`
	Else fileItemNext `json:"else,omitempty"`
}

func (i *fileItemIf) Validate() error {
	var err error
	if i.expr, err = expression.NewExpression(i.Expr); err != nil {
		return errors.Wrapf(err, "invalid expression(%s)", i.Expr)
	}
	if err := i.Then.Validate(); err != nil {
		return errors.Wrapf(err, "invalid then")
	}
	if err := i.Else.Validate(); err != nil {
		return errors.Wrapf(err, "invalid else")
	}
	return nil
} //fileItemIf.Validate()
