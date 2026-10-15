package web

import "embed"

// Files contains the embedded web assets for the Local EMS dashboard.
//
//go:embed dashboard.html component.html
var Files embed.FS
