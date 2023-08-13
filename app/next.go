package app

import (
	"context"

	"github.com/go-msvc/data"
	"github.com/go-msvc/errors"
	"github.com/go-msvc/expression"
	"github.com/gorilla/sessions"
)

type fileItemNext []fileItemNextStep

func (next fileItemNext) Validate() error {
	if len(next) == 0 {
		return errors.Errorf("missing")
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
			log.Debugf("next SET: %+v", step.Set)
			name := step.Set.Name.Rendered(sessionData(session))
			if !fieldNameRegex.MatchString(name) {
				return "", errors.Wrapf(err, "step[%d].name=\"%s\" is invalid fieldname (expecting CamelCase)", stepIndex, name)
			}
			value, ok := session.Values[step.Set.ValueStr]
			if ok {
				session.Values[name] = value
				log.Debugf("DIRECT SET(%s)=\"%s\"", name, value)
			} else {
				//array dereference...
				value, err = data.Get(sessionData(session), step.Set.ValueStr)
				if err != nil {
					value = step.Set.ValueStr
					//panic(fmt.Sprintf("cannot set from \"%s\": %+v", step.Set.ValueStr, err))
				}
			}
			// value := ...step.Set.Value.Rendered(sessionData(session))
			log.Debugf("SET(%s)=(%T)\"%v\"", name, value, value)
			session.Values[name] = value
			continue
		} //if SET
		if step.If != nil {
			log.Debugf("next If: %+v", step.If)
			condValue, err := step.If.expr.Eval(sessionForExpression(session))
			if err != nil {
				return "", errors.Wrapf(err, "failed to eval the expression")
			}
			log.Debugf("expr -> (%T)%+v", condValue, condValue)
			b, ok := condValue.(bool)
			if !ok {
				return "", errors.Errorf("if.expr(%s) -> (%T) != bool", step.If.Expr, condValue)
			}
			if b {
				log.Debugf("expr is TRUE (%T)%+v", condValue, condValue)
				if next, err := step.If.Then.Execute(ctx); err != nil {
					return "", errors.Wrapf(err, "failed to execute then")
				} else {
					if next != "" {
						log.Debugf("then next ITEM: %+v", next)
						return next, nil
					}
					//next not yet defined, continue with more actions
				}
			} else {
				log.Debugf("expr is FALSE (%T)%+v", condValue, condValue)
				if next, err := step.If.Else.Execute(ctx); err != nil {
					return "", errors.Wrapf(err, "failed to execute else")
				} else {
					if next != "" {
						log.Debugf("else next ITEM: %+v", next)
						return string(next), nil
					}
					//next not yet defined, continue with more actions
				}
			}
			continue
		} //if IF
		if step.Item != nil {
			log.Debugf("next ITEM: %+v", step.Set)
			return string(*step.Item), nil
		}
		return "", errors.Errorf("unhandled next step[%d] %T", stepIndex, step)
	}
	return "", nil
}

type fileItemNextStep struct {
	Item *fileItemNextItem `json:"item,omitempty" doc:"Value is next item id"`
	Set  *fileItemSet      `json:"set,omitempty"`
	If   *fileItemIf       `json:"if,omitemptu" doc:"Conditional step"`
}

type fileItemNextItem string

func (next fileItemNextStep) Validate() error {
	count := 0
	if next.Item != nil {
		if *next.Item == "" {
			return errors.Errorf("missing next item")
		}
		count++
	}
	if next.Set != nil {
		if err := next.Set.Validate(); err != nil {
			return errors.Wrapf(err, "invalid set")
		}
		count++
	}
	if next.If != nil {
		if err := next.If.Validate(); err != nil {
			return errors.Wrapf(err, "invalid if")
		}
		count++
	}
	if count == 0 {
		return errors.Errorf("missing item|set|if")
	}
	if count > 1 {
		return errors.Errorf("%d instead of 1 of id|set", count)
	}
	return nil
}

func sessionForExpression(s *sessions.Session) expression.IContext {
	return x{s: s}
}

type x struct {
	s *sessions.Session
}

func (x x) Get(name string) interface{} {
	value, ok := x.s.Values[name]
	if !ok {
		value = ""
	}
	log.Debugf("GETTING %s -> (%T)%+v", name, value, value)
	return value
}

func (x x) Set(name string, value interface{}) {
	log.Debugf("SETTING %s = (%T)%+v", name, value, value)
	x.s.Values[name] = value
}
