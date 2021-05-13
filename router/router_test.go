package router

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func assert(trueStatement bool, msg string) {
	if !trueStatement {
		_t.Error(msg)
	}
}

var (
	_t *testing.T
)

type ctra int
type ctrb int
type ctrc int

func (c *ctra) Name() string {
	return "dummyA"
}

func (c *ctrb) Name() string {
	return "dummyB"
}

func (c *ctrc) Name() string {
	return "dummyC"
}

func TestAddManyEntries(t *testing.T) {
	clearRoutingData()
	a := new(ctra)
	b := new(ctrb)
	c := new(ctrc)

	Add("GET", "/dummyA", a, "Index")
	Add("GET", "/dummyB", b, "Index")
	Add("GET", "/", c, "root")

	e := findRoutingEntry("GET", "/dummyA")
	assert(e != nil && e.key == "GET /dummyA" && e.handlerName == "Index", "Wrong key.")

	e = findRoutingEntry("GET", "/")
	assert(e != nil && e.key == "GET /" && e.handlerName == "root", "Wrong key.")
}

func TestRoutingEntryCreation(t *testing.T) {
	e, err := createRoutingEntry("GET", "/")
	assert(err == nil, "Error should be nil.")
	assert(e.key == "GET /", "Wrong path.")

	e, err = createRoutingEntry("GET", "/path")
	assert(err == nil, "Error should be nil.")
	assert(e.key == "GET /path", "Wrong path.")

	e, err = createRoutingEntry("GET", "/some/path")
	assert(err == nil, "Error should be nil.")
	assert(e.key == "GET /some/path", "Wrong path.")

	e, err = createRoutingEntry("GET", "/:someParam")
	assert(err == nil, "Error should be nil.")
	assert(e.key == "GET /:param", "Wrong path.")

	e, err = createRoutingEntry("GET", "/path/with/:someParam")
	assert(err == nil, "Error should be nil.")
	assert(e.key == "GET /path/with/:param", "Wrong path.")

	e, err = createRoutingEntry("GET", "/path/:with/two/:params")
	assert(err != nil && e == nil, "Two params not allowed")
}

func TestFindingRoutingEntry(t *testing.T) {
	clearRoutingData()
	c := new(ctra)
	Add("GET", "/", c, "dummyFunc")
	Add("GET", "/path", c, "dummyFunc")
	Add("GET", "/some/path", c, "dummyFunc")
	Add("GET", "/:someParam", c, "dummyFunc")
	Add("GET", "/path/with/:someParam", c, "dummyFunc")

	e := findRoutingEntry("GET", "/")
	assert(e.key == "GET /", "Wrong key.")

	e = findRoutingEntry("GET", "/justParam")
	assert(e.key == "GET /:param", "Wrong key.")

	e = findRoutingEntry("GET", "/no/path")
	assert(e == nil, "Wrong key.")
}

type testResponseWriter struct{ header http.Header }

func (trw testResponseWriter) Write([]byte) (int, error)  { return 0, nil }
func (trw testResponseWriter) WriteHeader(statusCode int) {}
func (trw testResponseWriter) Header() http.Header        { return trw.header }

func TestSetHeaderValuesUsingMap(t *testing.T) {
	var w http.ResponseWriter = testResponseWriter{header: make(http.Header)}
	controller := map[string]interface{}{}

	controller["headers"] = map[string]string{"K1": "v1", "K2": "v2", "K3": "v3"}

	setHeaderValues(w, reflect.ValueOf(controller))

	headers := controller["headers"].(map[string]string)
	fmt.Println(w.Header())
	for k, v := range w.Header() {
		assert(headers[k] == v[0], "headers doesn't match")
	}

	fmt.Println(w.Header())
}
