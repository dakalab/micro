package micro

import (
	"bytes"
	"html/template"
	"net/http"
)

// RedocOpts is configures for redoc
type RedocOpts struct {
	// Route is the route in http server, should include / at the beginning, default is /docs.
	// Currently it can not be root route "/", see https://github.com/grpc-ecosystem/grpc-gateway/issues/1909
	Route string
	// SpecURLs are the urls to find the spec for, format: name -> url
	SpecURLs map[string]string
	// RedocURL is the js that generates the redoc site, defaults to: https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js
	RedocURL string
	// Title is the page title, default to: API documentation
	Title string
	// Up is whether to boot up the redoc endpoints
	Up bool
}

// EnsureDefaults sets default redoc options
func (redoc *RedocOpts) EnsureDefaults() {
	if redoc.Route == "" || redoc.Route == "/" {
		redoc.Route = "/docs"
	}

	if redoc.SpecURLs == nil {
		redoc.AddSpec("Service", "/swagger.json")
	}

	if redoc.RedocURL == "" {
		redoc.RedocURL = "https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"
	}

	if redoc.Title == "" {
		redoc.Title = "API documentation"
	}
}

// AddSpec adds a spec url with name
func (redoc *RedocOpts) AddSpec(name, url string) *RedocOpts {
	if redoc.SpecURLs == nil {
		redoc.SpecURLs = make(map[string]string)
	}

	redoc.SpecURLs[name] = url

	return redoc
}

// Serve is the HandlerFunc for Redoc
func (redoc *RedocOpts) Serve(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {

	redoc.EnsureDefaults()

	tmpl := template.Must(template.New("redoc").Parse(redocTemplate))

	buf := bytes.NewBuffer(nil)
	tmpl.Execute(buf, redoc)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

const (
	redocTemplate = `<!DOCTYPE html>
<html>
  <head>
    <title>{{ .Title }}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
      body {
        margin: 0;
        padding-top: 40px;
      }
      nav {
        position: fixed;
        top: 0;
        width: 100%;
        z-index: 100;
      }
      #links_container {
          margin: 0;
          padding: 0;
          background-color: #0033a0;
      }
      #links_container li {
          display: inline-block;
          padding: 10px;
          color: white;
          cursor: pointer;
      }
    </style>
  </head>
  <body>

    <!-- Top navigation placeholder -->
    <nav>
      <ul id="links_container">
      </ul>
    </nav>

    <redoc scroll-y-offset="body > nav"></redoc>

    <script src="{{ .RedocURL }}"></script>
    <script>
      // list of APIS
      var apis = [
				{{range $key, $value := .SpecURLs}}
        {
          name: {{ $key }},
          url: {{ $value }}
        },
				{{end}}
      ];
      // initially render first API
      Redoc.init(apis[0].url);
      function onClick() {
        var url = this.getAttribute('data-link');
        Redoc.init(url);
      }
      // dynamically building navigation items
      var $list = document.getElementById('links_container');
      apis.forEach(function(api) {
        var $listitem = document.createElement('li');
        $listitem.setAttribute('data-link', api.url);
        $listitem.innerText = api.name;
        $listitem.addEventListener('click', onClick);
        $list.appendChild($listitem);
      });
    </script>
  </body>
</html>
`
)
