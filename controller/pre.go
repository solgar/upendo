package controller

import (
	"net/http"
	"reflect"
	"strings"
	"upendo/router"
	"upendo/session"
)

var (
	smanager *session.Manager = session.GetManager()
)

func init() {
	router.AddPreRouteFunc(CheckSession)
	router.AddPreRouteFunc(CheckCookies)
}

func CSet(cv reflect.Value, k string, v interface{}) {
	cv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
}

func CGet(cv reflect.Value, k string) interface{} {
	return cv.MapIndex(reflect.ValueOf(k)).Interface()
}

func CheckSession(controller reflect.Value) {
	path := CGet(controller, "path").(string)

	if strings.HasPrefix(path, "/css") || strings.HasPrefix(path, "/favico") || strings.HasPrefix(path, "/res") || strings.HasPrefix(path, "/js") {
		return
	}

	session := smanager.GetSession(CGet(controller, "request").(*http.Request))
	CSet(controller, "session", session)
}

func CheckCookies(controller reflect.Value) {
	request := CGet(controller, "request").(*http.Request)
	cookie, _ := request.Cookie("cookiesAccepted")

	if cookie != nil {
		CSet(controller, "cookiesAccepted", "true")
	}

}
