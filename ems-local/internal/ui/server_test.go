package ui

import (
	"strings"
	"testing"
)

func TestEmbeddedWebAssets(t *testing.T) {
	if dashboardHTML == "" {
		t.Fatal("dashboardHTML is empty")
	}
	if !strings.Contains(dashboardHTML, "<!DOCTYPE html>") {
		t.Error("dashboardHTML missing <!DOCTYPE html>")
	}
	if !strings.Contains(dashboardHTML, "/*ACTIVE_PAGE*/") {
		t.Error("dashboardHTML missing /*ACTIVE_PAGE*/ injection placeholder")
	}

	if componentHTML == "" {
		t.Fatal("componentHTML is empty")
	}
	if !strings.Contains(componentHTML, "<!DOCTYPE html>") {
		t.Error("componentHTML missing <!DOCTYPE html>")
	}
	if !strings.Contains(componentHTML, "/*COMP_ID*/") {
		t.Error("componentHTML missing /*COMP_ID*/ injection placeholder")
	}
}
