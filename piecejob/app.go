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
	piecejobApp.RegisterFunc("getProfile", getProfile)
	piecejobApp.RegisterFunc("updProfile", updProfile)
	piecejobApp.RegisterFunc("getMySkills", getMySkills)
	piecejobApp.RegisterFunc("listOfSkills", listOfSkills)
	piecejobApp.RegisterFunc("listOfJobs", listOfJobs)
	piecejobApp.RegisterFunc("getJob", getJob)
	piecejobApp.RegisterFunc("updJob", updJob)
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

type Job struct {
	Id      string
	Date    string
	Type    string
	Details string
}

var jobs = map[string]Job{
	"1": {Id: "1", Date: "Mon", Type: "Clean", Details: "1X"},
	"2": {Id: "2", Date: "Tue", Type: "Wash", Details: "2X"},
	"3": {Id: "3", Date: "Wed", Type: "Weld", Details: "3X"},
	"4": {Id: "4", Date: "Thu", Type: "Build", Details: "4X"},
	"5": {Id: "5", Date: "Fri", Type: "Paint", Details: "5X"},
}

func getJob(ctx context.Context, jobId string) (Job, error) {
	// if !natIdRegex.MatchString(jobId) {
	// 	return Profile{}, errors.Errorf("invalid natId=\"%s\"", natId)
	// }
	p, ok := jobs[jobId]
	if !ok {
		return Job{}, errors.Errorf("job not found")
	}
	return p, nil
}

func updJob(ctx context.Context, j Job) error {
	// if err := j.Validate(); err != nil {
	// 	return errors.Wrapf(err, "invalid profile")
	// }
	//todo: need to get id in req
	log.Debugf("Saving job:%+v", j)
	jobs[j.Id] = j
	return nil
}

// list returning struct that can be templated into items
func listOfJobs(ctx context.Context) (app.ColumnList, error) {
	listOfJobs := []app.ColumnItem{}
	for id, j := range jobs {
		listOfJobs = append(listOfJobs, app.ColumnItem{
			"Id":   id,
			"Date": j.Date,
			"Type": j.Type,
		})
	}
	return app.ColumnList{
		Items: listOfJobs,
	}, nil
}
