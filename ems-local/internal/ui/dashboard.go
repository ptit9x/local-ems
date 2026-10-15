package ui

import (
	"github.com/vmo/ems-platform/ems-local/web"
)

var dashboardHTML string

func init() {
	data, err := web.Files.ReadFile("dashboard.html")
	if err != nil {
		panic("failed to read embedded dashboard.html: " + err.Error())
	}
	dashboardHTML = string(data)
}
