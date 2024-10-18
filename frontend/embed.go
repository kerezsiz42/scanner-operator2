package frontend

import _ "embed"

//go:embed bundle.js
var BundleJs []byte

//go:embed index.html
var IndexHtml []byte

//go:embed output.css
var OutputCss []byte
