package app

import (
	"context"
	"encoding/json"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

// actions is a list of actions to execute in sequence
type Actions struct {
	list []Action
}

func (actions Actions) Validate() error {
	for actionIndex, action := range actions.list {
		if err := action.Validate(); err != nil {
			return errors.Wrapf(err, "invalid action[%d]", actionIndex)
		}
	}
	return nil
}

func (actions Actions) Execute(ctx context.Context) (err error) {
	log.Debugf("Executing %d actions ...", len(actions.list))
	for actionIndex, action := range actions.list {
		log.Debugf("Executing action[%d]:(%T) ...", actionIndex, action)
		if err := action.Execute(ctx); err != nil {
			return errors.Wrapf(err, "action[%d] failed", actionIndex)
		}
	}
	return nil
} //Actions.Execute()

func (actions *Actions) UnmarshalJSON(value []byte) error {
	actions.list = []Action{}

	log.Debugf("decoding JSON: %s", string(value))
	actionList := []map[string]interface{}{}
	if err := json.Unmarshal(value, &actionList); err != nil {
		return errors.Wrapf(err, "invalid action list (each must be a JSON object)")
	}
	log.Debugf("%d items:", len(actionList))
	for actionIndex, objNameAndAction := range actionList {
		//object has one item
		if len(objNameAndAction) != 1 {
			return errors.Errorf("action[%d] has %d keys instead of 1 which must be the output field name", actionIndex, len(objNameAndAction))
		}
		var outName string
		var action interface{}
		for outName, action = range objNameAndAction {
			//do nothing
		}
		log.Debugf("  action[%d]: %s = (%T)%+v", actionIndex, outName, action, action)

		//action is either a set value of a function call
		//to be a function call, it must be an obj with {"<func name>":<req>}
		if actionObj, ok := action.(map[string]interface{}); ok && len(actionObj) == 1 {
			var funcName string
			var funcReq interface{}
			for funcName, funcReq = range actionObj {
				//do nothing
			}
			if fnc, ok := registeredFuncByName[funcName]; ok {
				//this is a function call
				actions.list = append(actions.list, actionFunc{
					set:  outName,
					name: funcName,
					fnc:  fnc,
					req:  funcReq,
				})
				continue
			}
		}

		//not a func call, add as a simple assignment
		actions.list = append(actions.list, actionSet{
			set:   outName,
			value: action,
		})

	} //for each action in the list
	return nil
}

type Action interface {
	Validate() error
	Execute(ctx context.Context) error //todo...add req...
}

type actionFunc struct {
	set  string //must also be a template???
	name string
	fnc  AppFunc
	req  interface{} //value template - should be a generic type to render from session data with recursive values...
}

func (f actionFunc) Validate() error {
	if !fieldNameRegex.MatchString(f.set) {
		return errors.Errorf("invalid field name \"%s\"", f.set)
	}
	return errors.Errorf("NYI") //not sure if required, not called yet
}

func (f actionFunc) Execute(ctx context.Context) error {
	return errors.Errorf("NYI")
}

type actionSet struct {
	set   string      //must also be a template???
	value interface{} //value template - should be a generic type to render from session data with recursive values...
}

func (f actionSet) Validate() error {
	if !fieldNameRegex.MatchString(f.set) {
		return errors.Errorf("invalid field name \"%s\"", f.set)
	}
	return nil
}

func (f actionSet) Execute(ctx context.Context) error {
	session := ctx.Value(CtxSession{}).(*sessions.Session)
	session.Values[f.set] = f.value
	log.Debugf("action set \"%s\" = (%T)%+v", f.set, f.value, f.value)
	return nil
}

var (
	registeredFuncByName = map[string]AppFunc{}
)

// type ActionFuncOptions struct {
// 	Args interface{} `json:"args" doc:"Function input argument"`
// 	Set  string      `json:"set" doc:"Store function result in session data using this name"` //todo: tmpl to make the name
// }

// type ActionStep interface {
// 	Execute(ctx context.Context) error
// }

type actionStepFunc struct {
}
