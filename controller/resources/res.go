package resources

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"upendo/router"
	"upendo/settings"
)

type Resources map[string]interface{}

func init() {
	router.Add("GET", "/res/:file", Resources{}, "ImageResource")
	router.Add("GET", "/css/:file", Resources{}, "CSSResource")
	router.Add("GET", "/js/:file", Resources{}, "JSResource")
}

func (c Resources) SendResource() {
	file := c["file"].(string)
	if settings.IgnoreMapFiles && strings.HasSuffix(file, ".map") {
		return
	}

	w := c["writer"].(*bytes.Buffer)
	directory := c["directory"].(string)

	f, err := os.Open(settings.StartDir + directory + "/" + file)
	defer f.Close()

	if err != nil {
		fmt.Println("Error:", err)
		router.RedirectToError(c, http.StatusNotFound)
		return
	}

	var fbytes []byte
	fbytes, err = ioutil.ReadAll(f)

	if err != nil {
		fmt.Println("Error:", err)
		router.RedirectToError(c, http.StatusInternalServerError)
		return
	}

	writer := bufio.NewWriter(w)
	_, err = writer.Write(fbytes)
	writer.Flush()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func (c Resources) ImageResource() {
	// fix: set content type
	c["directory"] = "res"
	c.SendResource()
}

func (c Resources) CSSResource() {
	c["Content-Type"] = "text/css"
	c["directory"] = "css"
	c.SendResource()
}

func (c Resources) JSResource() {
	c["Content-Type"] = "application/javascript"
	c["directory"] = "js"
	c.SendResource()
}
