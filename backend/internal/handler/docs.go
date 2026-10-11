package handler

import (
	"net/http"

	"tabletalk/backend/api"
)

// swaggerUIVersion pins the Swagger UI assets loaded from the CDN.
const swaggerUIVersion = "5.33.1"

var swaggerUIPage = []byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TableTalk API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@` + swaggerUIVersion + `/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/docs/swagger.yaml",
      dom_id: "#swagger-ui",
      persistAuthorization: true,
      tryItOutEnabled: false,
    });
  </script>
</body>
</html>`)

// DocsUI serves Swagger UI for the embedded OpenAPI spec (enabled with API_DOCS_ENABLED).
func DocsUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(swaggerUIPage)
}

// DocsSpec serves the OpenAPI spec.
func DocsSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	_, _ = w.Write(api.Spec)
}
