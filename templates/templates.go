package templates

import "embed"

//go:embed *.tmpl fonts/*.woff2
var FS embed.FS
