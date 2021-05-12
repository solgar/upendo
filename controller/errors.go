package controller

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/solgar/upendo/router"
)

type ErrorsController router.Controller

func InstallErrors() {
	errorsPath := "GET /error/"
	router.AddPath(errorsPath+":errorCode", ErrorsController{}, "HandleErrors")
	for k := range router.ErrorsRouting {
		router.ErrorsRouting[k] = errorsPath + strconv.Itoa(k)
	}
}

func (c ErrorsController) HandleErrors() {
	w := c["writer"].(*bytes.Buffer)
	fmt.Fprintln(w, `<!DOCTYPE html><html><head></head><body>`)
	c.ErrorDescription()
	fmt.Fprintln(w, "</body></hmtl>")
}

func (c ErrorsController) ErrorDescription() string {
	w := c["writer"].(*bytes.Buffer)
	fmt.Fprintf(w, "Ooops! It seems that error %s occured. We are terribly sorry :( Try to refresh page or pick other link.", c["errorCode"])
	st := c["__stacktrace"]
	err := c["__error"]
	if err != nil {
		fmt.Fprint(w, "<p style=\"text-align: left;\">")
		fmt.Fprint(w, err)
		fmt.Fprint(w, "</p>")
	}
	var stackTrace []string
	fmt.Fprint(w, "<p style=\"text-align: left;\">")
	if st != nil {
		stackTrace = st.([]string)
		for _, trace := range stackTrace {
			fmt.Fprintf(w, "%s<br>", trace)
		}
	}
	fmt.Fprint(w, "</p>")

	return ""
}
