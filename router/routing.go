package router

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/solgar/upendo/settings"
)

const (
	printError500ToUser = true
)

type Controller map[string]interface{}

var (
	ErrorsRouting = map[int]string{
		http.StatusBadRequest:          "GET /error/400", // BAD REQUEST
		http.StatusUnauthorized:        "GET /error/401", // UNAUTHORIZED
		http.StatusPaymentRequired:     "GET /error/402", // Payment Required
		http.StatusForbidden:           "GET /error/403", // FORBIDDEN
		http.StatusNotFound:            "GET /error/404", // NOT FOUND
		http.StatusMethodNotAllowed:    "GET /error/405",
		http.StatusNotAcceptable:       "GET /error/406",
		http.StatusInternalServerError: "GET /error/500"} // INTERNAL SERVER ERROR

	routingTable       map[string]*routingEntry = make(map[string]*routingEntry)
	ignores            map[string]int           = make(map[string]int)
	paramNameReplacer  *regexp.Regexp           = nil
	allowedMethods                              = map[string]int{"GET": 1, "HEAD": 1, "POST": 1, "PUT": 1, "DELETE": 1, "TRACE": 1, "OPTIONS": 1, "CONNECT": 1, "PATCH": 1}
	preRouteFunctions  []func(reflect.Value)    = make([]func(reflect.Value), 0)
	postRouteFunctions []func(reflect.Value)    = make([]func(reflect.Value), 0)
	controllersTypes   map[string]reflect.Type  = make(map[string]reflect.Type)
)

type routingEntry struct {
	varPlace    int
	varName     string
	key         string
	handlerName string
	controller  reflect.Type
}

type routingContext struct {
	callChain []string
	errorCtx  *errorContext
}

func (ctx *routingContext) printCallChain() {
	fmt.Println("Printing routing context:")
	for i, call := range ctx.callChain {
		for j := 0; j < i; j++ {
			fmt.Print("\t")
		}
		fmt.Println("=>", call)
	}
}

func createRoutingContext(rootCall string) *routingContext {
	ctx := &routingContext{}
	if rootCall != "" {
		ctx.callChain = make([]string, 1)
		ctx.callChain[0] = rootCall
	} else {
		ctx.callChain = make([]string, 0)
	}
	return ctx
}

func panicOnError(e error) {
	panic(e.Error())
}

func clearRoutingData() {
	routingTable = make(map[string]*routingEntry)
}

func createRoutingEntry(method, path string) (*routingEntry, error) {
	//var err error
	if strings.Count(path, "/:") > 1 {
		return nil, errors.New("Error while creating routing entry: routing entry cannot have more than one parameter.")
	}

	method = strings.ToUpper(method)
	_, ok := allowedMethods[method]

	if !ok {
		return nil, errors.New("Error while creating routing entry: unknown method: " + method)
	}

	key := method + " "

	entry := &routingEntry{}
	entry.varPlace = strings.Index(path, "/:")

	if entry.varPlace != -1 {
		varNameAndRest := path[entry.varPlace+2:]
		if len(varNameAndRest) == 0 {
			return nil, errors.New("Error while creating routing entry: Invalid param name in path: " + path)
		}
		nextSep := strings.Index(varNameAndRest, "/")
		if nextSep != -1 {
			entry.varName = varNameAndRest[:nextSep]
			key += path[:entry.varPlace] + "/:param" + varNameAndRest[nextSep:]
		} else {
			entry.varName = varNameAndRest
			key += path[:entry.varPlace] + "/:param"
		}

		entry.key = key
	} else {
		entry.key = method + " " + path
	}

	return entry, nil
}

func assertControllerMapKind(i interface{}) {
	if reflect.ValueOf(i).Kind() != reflect.Map {
		panic("Controller \"" + reflect.TypeOf(i).Name() + "\" kind is not Map.")
	}
}

func AddPreRouteFunc(pre func(reflect.Value)) {
	preRouteFunctions = append(preRouteFunctions, pre)
}

func AddPostRouteFunc(post func(reflect.Value)) {
	postRouteFunctions = append(preRouteFunctions, post)
}

func AddPath(path string, controller interface{}, methodName string) {
	space := strings.Index(path, " ")
	Add(path[:space], path[space+1:], controller, methodName)
}

func Add(method, path string, controller interface{}, methodName string) {
	entry, err := createRoutingEntry(method, path)
	if err != nil {
		panic(err)
	}
	entry.controller = reflect.TypeOf(controller)
	entry.handlerName = methodName
	_, ok := routingTable[entry.key]
	if !ok {
		routingTable[entry.key] = entry
	} else {
		panic("Cannot add key: " + entry.key + ". Already in routing table.")
	}
	controllersTypes[entry.controller.Name()] = entry.controller
}

func AddIgnoredPath(path string) {
	ignores[path] = 0
}

func RemoveIgnoredPath(path string) {
	delete(ignores, path)
}

func RouteRequest(w http.ResponseWriter, r *http.Request) {
	_, ok := ignores[r.URL.Path]
	if ok {
		return
	}

	ctx := createRoutingContext("")

	if routeRequestSimple(w, r, ctx) == false {
		routeRequestUsingKey(w, r, ErrorsRouting[http.StatusNotFound], ctx)
	}
}

func routeRequestSimple(w http.ResponseWriter, r *http.Request, ctx *routingContext) bool {
	return routeRequest(w, r, r.Method, r.URL.Path, ctx)
}

func routeRequestUsingKey(w http.ResponseWriter, r *http.Request, key string, ctx *routingContext) bool {
	space := strings.Index(key, " ")
	return routeRequest(w, r, key[:space], key[space+1:], ctx)
}

func findRoutingEntry(method, path string) *routingEntry {
	key := method + " " + path
	e := routingTable[key]
	if e != nil {
		return e
	}

	prevIdx := -1
	sepIdx := len(key)
	for {
		prevIdx = sepIdx
		sepIdx = strings.LastIndex(key[:prevIdx], "/")

		if sepIdx == -1 {
			return nil
		}

		k := key[:sepIdx+1] + ":param" + key[prevIdx:]
		e = routingTable[k]
		if e != nil {
			return e
		}
	}
}

func MakeCleanParams(params map[string]interface{}) map[string]interface{} {
	p := make(map[string]interface{})
	p["__writer"] = params["__writer"]
	p["writer"] = params["writer"]
	p["request"] = params["request"]
	p["method"] = params["method"]
	p["path"] = params["path"]
	return p
}

// needed for stack trace
type errorContext struct {
	StackTrace []string
	err        interface{}
}

//var stackTraces map[*http.Request][]string = make(map[*http.Request][]string)
var stackTraces map[*http.Request]*errorContext = make(map[*http.Request]*errorContext)
var criticalSection *sync.Mutex = &sync.Mutex{}

// this should be optimised!
func routeRequest(w http.ResponseWriter, r *http.Request, method, path string, ctx *routingContext) (success bool) {
	// if something happens before success is set to true consider it to be bad
	success = false

	ctx.callChain = append(ctx.callChain, method+" "+path)
	if len(ctx.callChain) > settings.RoutingChainMax {
		fmt.Println("Error: Call chain exceeded RoutingChainMax.")
		ctx.printCallChain()
		// to break call chain
		success = true
	}

	// recover from any panic and redirect to 500 page
	defer func() {
		e := recover()
		if e != nil {
			buf := make([]byte, 3048)
			runtime.Stack(buf, false)
			success = true
			if method+" "+path == ErrorsRouting[http.StatusInternalServerError] {
				fmt.Println("Internal server error occured during processing page for \"Internal Server Error\". Kinda funny... After this error previous error will be printed.")
				ctx.printCallChain()
				fmt.Println("Error:", e)
				st := strings.Split(string(buf), "\n")[3:]
				for _, trace := range st {
					fmt.Println(trace)
				}
				fmt.Println()
				fmt.Println("Previous error:", ctx.errorCtx.err)
				for _, trace := range ctx.errorCtx.StackTrace[3:] {
					fmt.Println(trace)
				}

			} else {
				// stack trace critical section lock and unlock
				criticalSection.Lock()
				errCtx := &errorContext{strings.Split(string(buf), "\n"), e}
				stackTraces[r] = errCtx
				ctx.errorCtx = errCtx
				criticalSection.Unlock()
				routeRequestUsingKey(w, r, ErrorsRouting[http.StatusInternalServerError], ctx)
			}
		}
	}()

	entry := findRoutingEntry(method, path)
	if entry == nil {
		return
	}

	buff := new(bytes.Buffer)
	controller := reflect.MakeMap(entry.controller)

	// handle stack trace from recovered panic
	var sTrace []string = nil
	var errorFromRecover interface{} = nil

	criticalSection.Lock()
	if len(stackTraces) > 0 {
		errorCtx, ok := stackTraces[r]
		if ok {
			sTrace = errorCtx.StackTrace
			errorFromRecover = errorCtx.err
			delete(stackTraces, r)
		}
	}
	criticalSection.Unlock()
	if sTrace != nil {
		controller.SetMapIndex(reflect.ValueOf("__stacktrace"), reflect.ValueOf(sTrace))
		controller.SetMapIndex(reflect.ValueOf("__error"), reflect.ValueOf(errorFromRecover))
	}

	controller.SetMapIndex(reflect.ValueOf("ControllerName"), reflect.ValueOf(entry.controller.Name()))
	controller.SetMapIndex(reflect.ValueOf("__writer"), reflect.ValueOf(w))
	controller.SetMapIndex(reflect.ValueOf("writer"), reflect.ValueOf(buff))
	controller.SetMapIndex(reflect.ValueOf("request"), reflect.ValueOf(r))
	controller.SetMapIndex(reflect.ValueOf("method"), reflect.ValueOf(r.Method))
	controller.SetMapIndex(reflect.ValueOf("path"), reflect.ValueOf(path))
	controller.SetMapIndex(reflect.ValueOf("headers"), reflect.ValueOf(map[string]string{}))

	if entry.varPlace != -1 {
		varValueBeginStr := path[entry.varPlace+1:]
		nextSep := strings.Index(varValueBeginStr, "/")
		if nextSep == -1 {
			nextSep = len(varValueBeginStr)
		}
		varValue := varValueBeginStr[:nextSep]
		controller.SetMapIndex(reflect.ValueOf(entry.varName), reflect.ValueOf(varValue))
	}

	for _, f := range preRouteFunctions {
		f(controller)
	}

	handlerMethod := controller.MethodByName(entry.handlerName)
	if !handlerMethod.IsValid() {
		panic("Cannot route: Controller " + entry.controller.Name() + " doesn't have function " + entry.handlerName + ".")
	}
	handlerMethod.Call([]reflect.Value{})

	setHeaderValue(w, "Content-Type", controller)
	setHeaderValue(w, "Location", controller)

	setHeaderValues(w, controller)

	statusCode := setStatusCode(w, controller)

	if statusCode != 303 {
		buffAfter := controller.MapIndex(reflect.ValueOf("writer")).Interface().(*bytes.Buffer)
		if buffAfter != nil {
			buff = buffAfter
			_, err := w.Write(buff.Bytes())

			for _, f := range postRouteFunctions {
				f(controller)
			}

			if err != nil {
				panic(err)
			}
		}
	}

	success = true
	return
}

func setHeaderValue(w http.ResponseWriter, key string, controller reflect.Value) {
	value := controller.MapIndex(reflect.ValueOf(key))
	if value.IsValid() {
		w.Header().Set(key, value.Interface().(string))
	}
}

func setHeaderValues(w http.ResponseWriter, c reflect.Value) {
	controller := c.Interface().(map[string]interface{})
	if h, ok := controller["headers"]; ok {
		headers := h.(map[string]string)
		for k, v := range headers {
			w.Header().Set(k, v)
		}
	}
}

func setStatusCode(w http.ResponseWriter, controller reflect.Value) int {
	value := controller.MapIndex(reflect.ValueOf("StatusCode"))
	if value.IsValid() {
		v := value.Interface().(int)
		w.WriteHeader(v)
		return v
	}
	return 200
}

func Redirect(c map[string]interface{}, path string) int {
	c["StatusCode"] = 303
	c["Location"] = path
	return 0
}

func RedirectToError(c map[string]interface{}, errVal int) {
	c["__redirected"] = true
	path := ErrorsRouting[errVal][4:]
	Redirect(c, path)
}

func PrintRouting() {
	for k, v := range routingTable {
		fmt.Println(k, "=>", v.controller.Name(), "=>", v.handlerName)
	}
}
