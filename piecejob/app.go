package piecejob

import (
	"context"
	"regexp"

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
	piecejobApp.RegisterFunc("getProfile", getProfile)
	piecejobApp.RegisterFunc("updProfile", updProfile)
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
		Items: skills,
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
		Items: jobs,
	}, nil
}

// profile keyed on national id
var profiles = map[string]Profile{}

type Profile struct {
	NatId string
	Name  string
	Dob   string `label:"Date of birth"`
	ID    string `label:"National ID"`
}

func (p Profile) Validate() error {
	if !natIdRegex.MatchString(p.NatId) {
		return errors.Errorf("invalid natId(%s)", p.NatId)
	}
	return nil
} //Profile.Validate()

const natIdPattern = `[0-9]{13}`

var natIdRegex = regexp.MustCompile("^" + natIdPattern + "$")

func getProfile(ctx context.Context, natId string) (Profile, error) {
	if !natIdRegex.MatchString(natId) {
		return Profile{}, errors.Errorf("invalid natId=\"%s\"", natId)
	}
	p, ok := profiles[natId]
	if !ok {
		return Profile{
			NatId: natId,
		}, nil
	}
	log.Debugf("Retrieved profile(%s):%+v", natId, p)
	return p, nil
}

func updProfile(ctx context.Context, p Profile) error {
	if err := p.Validate(); err != nil {
		return errors.Wrapf(err, "invalid profile")
	}
	//todo: need to get id in req
	log.Debugf("Saving profile:%+v", p)
	profiles[p.NatId] = p
	return nil
}
