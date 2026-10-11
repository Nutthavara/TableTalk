// Package api embeds the OpenAPI description of the HTTP API (swagger.yaml).
package api

import _ "embed"

//go:embed swagger.yaml
var Spec []byte
