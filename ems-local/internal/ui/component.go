package ui

import (
	"github.com/vmo/ems-platform/ems-local/web"
)

// componentHTML is the standalone component detail page loaded from embedded web assets.
var componentHTML string

func init() {
	data, err := web.Files.ReadFile("component.html")
	if err != nil {
		panic("failed to read embedded component.html: " + err.Error())
	}
	componentHTML = string(data)
}
