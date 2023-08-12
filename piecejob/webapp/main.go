package main

import (
	"fmt"

	"github.com/jansemmelink/goweb1/piecejob"
	"github.com/jansemmelink/goweb1/web"
)

func main() {
	app, err := piecejob.App()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
	web.New(app).Run()
}
