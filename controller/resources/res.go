package resources

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"upendo/router"
	"upendo/settings"
)

type Resources map[string]interface{}

func init() {
	router.Add("GET", "/res/:file", Resources{}, "ImageResource")
	router.Add("GET", "/css/:file", Resources{}, "CSSResource")
}

func (c Resources) ImageResource() {
	// fix: set content type
	w := c["writer"].(*bytes.Buffer)

	f, err := os.Open(settings.StartDir + "res/" + c["file"].(string))
	defer f.Close()

	if err != nil {
		router.RedirectToError(c, http.StatusNotFound)
	}

	var fbytes []byte
	fbytes, err = ioutil.ReadAll(f)

	if err != nil {
		fmt.Println("Error:", err)
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

func (c Resources) CSSResource() {
	w := c["writer"].(*bytes.Buffer)

	c["Content-Type"] = "text/css"

	f, err := os.Open(settings.StartDir + "css/" + c["file"].(string))
	defer f.Close()

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var fbytes []byte
	fbytes, err = ioutil.ReadAll(f)

	if err != nil {
		fmt.Println("Error:", err)
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
