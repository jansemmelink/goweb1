package main

import (
	"encoding/json"
	"os"

	"github.com/go-msvc/errors"
)

type App map[string]Item

func (app App) Validate() error {
	if len(app) < 1 {
		return errors.Errorf("empty app")
	}
	for name, item := range app {
		if name == "" {
			return errors.Errorf("empty item name")
		}
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "invalid item \"%s\"", name)
		}
	}
	return nil
} //app.Validate()

type Item struct {
	Menu   *Menu   `json:"menu"`
	Prompt *Prompt `json:"prompt"`
	Final  *Final  `json:"final"`
}

func (item Item) Validate() error {
	count := 0
	if item.Menu != nil {
		if err := item.Menu.Validate(); err != nil {
			return errors.Wrapf(err, "invalid menu")
		}
		count++
	}
	if item.Prompt != nil {
		if err := item.Prompt.Validate(); err != nil {
			return errors.Wrapf(err, "invalid prompt")
		}
		count++
	}
	if item.Final != nil {
		if err := item.Final.Validate(); err != nil {
			return errors.Wrapf(err, "invalid final")
		}
		count++
	}
	if count != 1 {
		return errors.Errorf("has %d menu|prompt|final instead of exactly 1", count)
	}
	return nil
}

type Menu struct {
	Title string     `json:"title"`
	Items []MenuItem `json:"items"`
}

func (menu Menu) Validate() error {
	if menu.Title == "" {
		return errors.Errorf("missing title")
	}
	if len(menu.Items) == 0 {
		return errors.Errorf("missing items")
	}
	for itemIndex, item := range menu.Items {
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "invalid item[%d]", itemIndex)
		}
	}
	return nil
} //Menu.Validate()

type MenuItem struct {
	Caption string `json:"caption"`
	Next    string `json:"next"`
}

func (menuItem MenuItem) Validate() error {
	if menuItem.Caption == "" {
		return errors.Errorf("missing caption")
	}
	if menuItem.Next == "" {
		return errors.Errorf("missing next")
	}
	return nil
}

type Prompt struct {
	Caption string `json:"caption"`
	Name    string `json:"name"`
	Next    string `json:"next"`
}

func (prompt Prompt) Validate() error {
	if prompt.Caption == "" {
		return errors.Errorf("missing caption")
	}
	if prompt.Name == "" {
		return errors.Errorf("missing name")
	}
	if prompt.Next == "" {
		return errors.Errorf("missing next")
	}
	return nil
}

type Final struct {
	Caption string `json:"caption"`
}

func (final Final) Validate() error {
	if final.Caption == "" {
		return errors.Errorf("missing caption")
	}
	return nil
}

func Load(filename string) (App, error) {
	f, err := os.Open(filename)
	if err != nil {
		return App{}, errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer f.Close()

	var app App
	if err := json.NewDecoder(f).Decode(&app); err != nil {
		return App{}, errors.Wrapf(err, "failed to read JSON from file %s", filename)
	}
	if err := app.Validate(); err != nil {
		return App{}, errors.Wrapf(err, "invalid app")
	}
	return app, nil
} //Load()
