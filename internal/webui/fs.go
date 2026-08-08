// Package webui embeds the browser demo.
package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:static
var static embed.FS

// Handler serves the embedded demo files.
func Handler() http.Handler {
	sub, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}
