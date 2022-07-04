package main

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

const debug = true

var (
	//go:embed assets
	assetFS embed.FS
	//go:embed templates
	templateFS embed.FS
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Println("finished program")
}

func run() error {
	smux := http.NewServeMux()
	tmpl, err := template.New("base").Funcs(funcmap).ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return err
	}
	python := os.Getenv("PYTHON3")
	if python == "" {
		return errors.New("PYTHON3 env variable not set. must be python executable location or command name. i.e. \"python3\".")
	}
	evaluator := Evaluator{
		tmpls:  tmpl,
		runner: cmdRunna(python),
	}
	err = evaluator.ParseAndEvaluateGlob("evaluations/*.py")
	if err != nil {
		return err
	}
	smux.Handle("/assets/", http.FileServer(http.FS(assetFS)))
	smux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/py/evals", http.StatusTemporaryRedirect)
	})
	smux.HandleFunc("/py/evals", evaluator.handleListEvaluations)
	// smux.HandleFunc("/py/evals", evaluator.handleEvaluation)

	smux.HandleFunc("/py/run/", evaluator.handleRun)
	sv := userMiddleware(smux)
	addr := ":8080"
	log.Println("Server started at http://127.0.0.1" + addr)
	return http.ListenAndServe(addr, sv)
}

var funcmap = template.FuncMap{
	"assetPath": func(asset string) string { return "/assets/" + asset },
	"safe":      func(html string) template.HTML { return template.HTML(html) },
	"markdown": func(input string) template.HTML {
		result := blackfriday.Run([]byte(input))
		p := bluemonday.UGCPolicy()
		output := p.Sanitize(string(result))
		return template.HTML(output)
	},
}
