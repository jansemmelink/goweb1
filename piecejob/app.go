package piecejob

import (
	"context"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
	"github.com/gorilla/sessions"
	"github.com/jansemmelink/goweb1/app"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func App() (app.App, error) {
	piecejobApp := app.New()

	//todo: install modules
	piecejobApp.RegisterFunc("getMySkills", getMySkills)
	//...
	//piecejobApp.Register("some-id", myFunc)
	//piecejobApp.Register("other-id", myType{})

	//todo: load app from file
	if err := piecejobApp.Load("../app.json"); err != nil {
		return nil, errors.Wrapf(err, "failed to load app.json")
	}
	return piecejobApp, nil
}

func getMySkills(ctx context.Context, args map[string]interface{}) error {
	mySkillsList := []string{"Cleaner", "Painter"}
	session := ctx.Value(app.CtxSession{}).(sessions.Session)
	session.Values["MySkills"] = mySkillsList
	return nil
}
