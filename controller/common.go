package controller

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"reflect"
	_ "upendo/controller/resources"
	"upendo/pages"
	"upendo/settings"
)

var (
	templates map[string]*pages.Page = make(map[string]*pages.Page)
)

func HandlePageTemplate(controller interface{}, template string) {
	if settings.ReloadTemplates {
		pages.LoadTemplates(settings.TemplatesDir)
	}

	buff := reflect.ValueOf(controller).MapIndex(reflect.ValueOf("writer")).Interface().(*bytes.Buffer)

	if pages.TemplatesRoot == nil {
		panic("TemplatesRoot is nil! Probably no templates found in directory: " + settings.StartDir + settings.TemplatesDir)
	}

	err := pages.TemplatesRoot.ExecuteTemplate(buff, template, controller)
	if err != nil {
		panic(err)
	}
}

func ReadBodyAsString(params map[string]interface{}) string {
	r := params["request"].(*http.Request)
	bodyBuff, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	return string(bodyBuff)
}

func PanicIfNeeded(err error) {
	if err != nil {
		panic(err)
	}
}
