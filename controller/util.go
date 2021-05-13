package controller

func AddHeader(c map[string]interface{}, k, v string) {
	c["headers"].(map[string]string)[k] = v
}
