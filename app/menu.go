package app

import "github.com/go-msvc/logger"

var log = logger.New().WithLevel(logger.LevelDebug)

// type AppMenu interface {
// 	With(MenuOption) AppMenu
// }

// func Menu(options ...MenuOption) AppMenu {
// 	m := menu{
// 		lines: []menuLine{},
// 	}
// 	for _, o := range options {
// 		m = m.With(o).(menu)
// 	}
// 	return m
// }

// type menu struct {
// 	lines []menuLine
// }

// type MenuOption interface{}

// type menuLine struct {
// 	caption string
// 	next    App
// }

// func (m menu) With(o MenuOption) AppMenu {
// 	log.Debugf("with %T", o)
// 	return m
// }

// func MenuTitle(title string) MenuOption {
// 	return menuTitle{title: title}
// }

// type menuTitle struct {
// 	title string
// }

// func MenuItem(caption string, next App) MenuOption {
// 	return menuItem{caption: caption, next: next}
// }

// type menuItem struct {
// 	caption string
// 	next    App
// }
