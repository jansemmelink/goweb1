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
	piecejobApp.RegisterFunc("listOfSkills", listOfSkills)
	piecejobApp.RegisterFunc("listOfJobs", listOfJobs)
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

// list returning simple list of strings
func listOfSkills(ctx context.Context, req GetMySkillsReq) (app.ColumnList, error) {
	log.Debugf("req: (%T)%+v", req, req)
	skills := []app.ColumnItem{
		{"Skill": "Cleaner"},
		{"Skill": "Painter"},
	}
	return app.ColumnList{
		Columns: []string{"Skill"},
		Items:   skills,
	}, nil
}

type job struct {
	Date string
	Type string
}

// list returning struct that can be templated into items
func listOfJobs(ctx context.Context) (app.ColumnList, error) {
	jobs := []app.ColumnItem{
		{"Date": "Mon", "Type": "Clean"},
		{"Date": "Tue", "Type": "Wash"},
		{"Date": "Wed", "Type": "Weld"},
		{"Date": "Thu", "Type": "Build"},
		{"Date": "Fri", "Type": "Paint"},
	}
	return app.ColumnList{
		Columns: []string{"Date", "Type"},
		Items:   jobs,
	}, nil
}
