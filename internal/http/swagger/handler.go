package swagger

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

const swaggerUIHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
      html { box-sizing: border-box; overflow-y: scroll; }
      *, *:before, *:after { box-sizing: inherit; }
      body { margin: 0; background: #fafafa; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.addEventListener('load', function () {
        window.ui = SwaggerUIBundle({
          url: '/swagger/openapi.yaml',
          dom_id: '#swagger-ui'
        });
      });
    </script>
  </body>
</html>
`

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger":
			http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		case "/swagger/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(swaggerUIHTML))
		case "/swagger/openapi.yaml":
			spec, err := readSpecFile()
			if err != nil {
				http.Error(w, "swagger spec is unavailable", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/yaml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(spec)
		default:
			http.NotFound(w, r)
		}
	})
}

func readSpecFile() ([]byte, error) {
	paths := []string{
		filepath.Join("docs", "swagger", "swagger.yaml"),
	}

	executablePath, err := os.Executable()
	if err == nil {
		executableDir := filepath.Dir(executablePath)
		paths = append(paths, filepath.Join(executableDir, "docs", "swagger", "swagger.yaml"))
	}

	_, sourceFile, _, ok := runtime.Caller(0)
	if ok {
		sourceDir := filepath.Dir(sourceFile)
		paths = append(paths, filepath.Clean(filepath.Join(sourceDir, "..", "..", "..", "docs", "swagger", "swagger.yaml")))
	}

	for _, path := range paths {
		spec, readErr := os.ReadFile(path)
		if readErr == nil {
			return spec, nil
		}
	}

	return nil, os.ErrNotExist
}
