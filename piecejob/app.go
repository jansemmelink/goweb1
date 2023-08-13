package piecejob

import (
	"context"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
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

type GetMySkillsReq struct {
	UpperCase bool
	Prefix    string
}

func getMySkills(ctx context.Context, req GetMySkillsReq) ([]string, error) {
	log.Debugf("req: (%T)%+v", req, req)
	mySkillsList := []string{"Cleaner", "Painter"}
	return mySkillsList, nil
}
