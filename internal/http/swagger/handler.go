package swagger

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type pageData struct {
	Spec string
}

var indexTemplate = template.Must(template.New("swagger-index").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Swagger Docs</title>
    <style>
      :root { color-scheme: light; }
      body {
        margin: 0;
        padding: 32px;
        font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
        background: #f7f7f5;
        color: #1f2328;
      }
      main {
        max-width: 1100px;
        margin: 0 auto;
      }
      h1 {
        margin-top: 0;
      }
      a {
        color: #0b57d0;
      }
      pre {
        overflow-x: auto;
        padding: 20px;
        background: #ffffff;
        border: 1px solid #d0d7de;
        border-radius: 10px;
        line-height: 1.45;
      }
    </style>
  </head>
  <body>
    <main>
      <h1>Swagger Docs</h1>
      <p>Raw OpenAPI spec: <a href="/swagger/openapi.yaml">/swagger/openapi.yaml</a></p>
      <pre>{{ .Spec }}</pre>
    </main>
  </body>
</html>
`))

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger":
			http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
		case "/swagger/":
			spec, err := readSpecFile()
			if err != nil {
				http.Error(w, "swagger spec is unavailable", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = indexTemplate.Execute(w, pageData{Spec: string(spec)})
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
