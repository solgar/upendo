package pages

import (
	"testing"
)

func assert(trueStatement bool, msg string) {
	if !trueStatement {
		_t.Error(msg)
	}
}

var (
	_t *testing.T = nil
)

func TestTemplates(t *testing.T) {
}
