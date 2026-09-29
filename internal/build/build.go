// Package build holds build-time metadata injected via ldflags.
package build

// Set via -ldflags at build time. See Makefile.
var (
	Version   = "dev"
	BuildTime = "unknown"
)
