package controller

import "github.com/solgar/upendo/router"

func AddHeader(c interface{}, k, v string) {
	cc := c.(router.Controller)
	cc.AddHeader(k, v)
}
