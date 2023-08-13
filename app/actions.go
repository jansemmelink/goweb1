package app

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-msvc/errors"
	"github.com/gorilla/sessions"
)

// actions is a list of actions to execute in sequence
type Actions struct {
	list []Action
}

func (actions *Actions) Validate(app App) error {
	for actionIndex, action := range actions.list {
		if err := action.Validate(app); err != nil {
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
	actionList := []map[string]interface{}{}
	if err := json.Unmarshal(value, &actionList); err != nil {
		return errors.Wrapf(err, "invalid action list (each must be a JSON object)")
	}
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
		// log.Debugf("  action[%d]: %s = (%T)%+v", actionIndex, outName, action, action)

		//action is either a set value of a function call
		//to be a function call, it must be an obj with {"<func name>":<req>}
		if actionObj, ok := action.(map[string]interface{}); ok {
			if len(actionObj) == 1 {
				var funcName string
				var funcReq interface{}
				for funcName, funcReq = range actionObj {
					//do nothing
				}
				//function name resolution is done in the app - not accessible here
				//if name ends with "()" then it is a func name
				if len(funcName) > 2 && strings.HasSuffix(funcName, "()") {
					log.Debugf("  action[%d]: %s = func %s(%+v)", actionIndex, outName, funcName, funcReq)
					actions.list = append(actions.list, &actionFunc{
						set:  outName,
						name: funcName[0 : len(funcName)-2], //trim "()"
						fnc:  nil,                           //resolved at runtime for now...
						req:  funcReq,
					})
					continue
				}
			}
		}
		log.Debugf("  action[%d] %s = %+v", actionIndex, outName, action)

		//not a func call, add as a simple assignment
		actions.list = append(actions.list, &actionSet{
			set:   outName,
			value: action,
		})

	} //for each action in the list
	return nil
}

type Action interface {
	Validate(app App) error
	Execute(ctx context.Context) error //todo...add req...
}

type actionFunc struct {
	set  string //must also be a template???
	name string
	fnc  *AppFunc
	req  interface{} //value template - should be a generic type to render from session data with recursive values...
}

func (f *actionFunc) Validate(app App) error {
	if f.set != "" && !fieldNameRegex.MatchString(f.set) { //may be empty when not storing anything, e.g. func has no result value
		return errors.Errorf("invalid field name \"%s\"", f.set)
	}
	if f.name == "" {
		return errors.Errorf("missing name")
	}
	var ok bool
	f.fnc, ok = app.FuncByName(f.name)
	if !ok {
		return errors.Errorf("unknown func %s", f.name)
	}
	return nil
}

func (f actionFunc) Execute(ctx context.Context) error {
	args := []reflect.Value{
		reflect.ValueOf(ctx),
	}
	if f.fnc.reqType != nil {
		//todo: execute req templates into a value... for now just static value as configured
		jsonReq, _ := json.Marshal(f.req)
		reqValuePtr := reflect.New(f.fnc.reqType)
		if err := json.Unmarshal(jsonReq, reqValuePtr.Interface()); err != nil {
			return errors.Wrapf(err, "failed to parse %s() req into %v", f.name, f.fnc.reqType)
		}
		log.Debugf("req: (%T)%+v", reqValuePtr.Elem().Interface(), reqValuePtr.Elem().Interface())
		args = append(args, reqValuePtr.Elem())
	}
	results := f.fnc.funcValue.Call(args)
	errValue := results[len(results)-1] //i.e. last result from the func
	log.Debugf("err valid: %v", errValue.IsValid())
	log.Debugf("err nil: %v", errValue.IsNil())
	if !errValue.IsNil() {
		return errors.Wrapf(errValue.Interface().(error), "action func %s() failed", f.name)
	}

	if f.set != "" {
		if len(results) != 2 {
			return errors.Errorf("func %s() did not return a value for %s", f.name, f.set)
		}
		session := ctx.Value(CtxSession{}).(*sessions.Session)
		session.Values[f.set] = results[0].Interface()
		log.Debugf("action %s(): %s = (%T)%+v", f.name, f.set, results[0].Interface(), results[0].Interface())
	}
	return nil
}

type actionSet struct {
	set   string      //must also be a template???
	value interface{} //value template - should be a generic type to render from session data with recursive values...
}

func (f *actionSet) Validate(app App) error {
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
