package appcontrollers

import (
	"bytes"
	"fmt"
	"strconv"
	. "upendo/controller"
	"upendo/pages"
	"upendo/router"
)

type ExampleController router.Controller

func init() {
	pages.RegisterFunction("echo", echo)

	router.Add("GET", "/", ExampleController{}, "Index")
}

func (c ExampleController) Index() {
	HandlePageTemplate(c, "index.html")
}

// FunctionWithoutParams is called from template so it has to return string value
func (c ExampleController) FunctionWithoutParams() string {
	c["mapKey"] = "this is controller (which is a map) mapKey value"
	return "FunctionWithoutParams()"
}

// WithParams is called from template so it has to return string value but writes response directly to writer
func (c ExampleController) WithParams(passedParam string, intParam int) string {
	w := c["writer"].(*bytes.Buffer)
	fmt.Fprint(w, "WithParams: "+passedParam+", intParam: "+strconv.Itoa(intParam))
	return ""
}

// echo is function registered as a template function
func echo(msg string) string {
	return msg
}
