package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/labstack/echo"
)

func createView(c echo.Context, name string, path string) error {
	t := template.New(name)
	fmt.Printf("ページ'%s'へのアクセスがありました", name)

	t, err := t.ParseFiles(path)
	if err != nil {
		log.Fatalln(err)
	}

	var buff bytes.Buffer
	if err = t.Execute(&buff, ViewEnv{
		EnvAuthBaseUri: envAuthBaseUri,
	}); err != nil {
		log.Fatalln(err)
	}
	return c.HTML(200, buff.String())
}

func createLogin(c echo.Context) error {
	return createView(c, "login", "views/1-login.html")
}

func createOnetime(c echo.Context) error {
	return createView(c, "onetime", "views/2-onetime.html")
}

func createTest(c echo.Context) error {
	return createView(c, "test", "views/3-test.html")
}
