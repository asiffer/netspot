// dashboard.go

package api

import (
	"embed"
	"io/fs"
	"net/http"
	"netspot/analyzer"
	"netspot/config"
	"netspot/miner"
	"text/template"
	"time"
)

//go:embed web/index.html web/static/css/*.css web/static/images/* web/static/fonts/*.otf
var content embed.FS
var prefix = "web"
var root, _ = fs.Sub(content, prefix)

// DashboardHandler serves the dashboard
func DashboardHandler(w http.ResponseWriter, r *http.Request) {

	// unflatten config
	uconf := config.GetConfig(true)

	// add extra parameters
	uconf["isDeviceInterface"] = miner.IsDeviceInterface()
	uconf["isRunning"] = analyzer.IsRunning()

	// modifiers
	funcMap := template.FuncMap{
		// The name "toSeconds" is what the function will be called in the template text.
		"toSeconds": func(i interface{}) float64 {
			switch v := i.(type) {
			case time.Duration:
				return v.Seconds()
			case float64:
				return v / 1e9
			default:
				return 0.
			}
		},
	}
	t := template.New("Title").Funcs(funcMap)

	// Déclaration des fichiers à parser
	// t = template.Must(t.Parse(dashbardTemplate))
	// t = template.Must(t.ParseFiles("/home/asr/go/src/netspot/api/templates/main.html"))

	data, err := content.ReadFile("web/index.html")
	if err != nil {
		apiLogger.Error().Msgf("Cannot read file 'index.html': %v", err)
	}
	t = template.Must(t.Parse(string(data)))
	w.WriteHeader(http.StatusOK)
	t.ExecuteTemplate(w, "layout", uconf)
}
