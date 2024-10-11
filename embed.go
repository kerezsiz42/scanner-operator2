package root

import "embed"

//go:embed frontend/bundle.js
var BundleJs embed.FS

//go:embed frontend/index.html
var IndexHtml embed.FS

//go:embed frontend/output.css
var OutputCss embed.FS
