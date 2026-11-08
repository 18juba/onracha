package httpapi

import (
	"net/http"
	"os"
)

const swaggerUIVersion = "5.34.1"

const swaggerUIHTML = `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>ONRACHA API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.34.1/swagger-ui.css">
  <style>html { box-sizing: border-box; overflow-y: scroll } *, *:before, *:after { box-sizing: inherit } body { margin: 0; background: #fafafa }</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.34.1/swagger-ui-bundle.js" crossorigin="anonymous"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        validatorUrl: null,
        persistAuthorization: false,
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
`

func (h *handler) swaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerUIHTML))
}

func (h *handler) openAPISpec(w http.ResponseWriter, _ *http.Request) {
	specification, err := os.ReadFile(h.config.OpenAPISpecPath)
	if err != nil {
		h.logger.Error("OpenAPI specification could not be read", "error", err)
		writeError(w, http.StatusInternalServerError, "openapi_unavailable", "A especificação da API está indisponível.")
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(specification)
}
