package main

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
)

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
	eval := cmdRunna(python)
	smux.Handle("/assets/", http.FileServer(http.FS(assetFS)))
	smux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		err := tmpl.Lookup("eval.tmpl").Execute(rw, struct{ EvaluationID int }{EvaluationID: 1})
		if err != nil {
			log.Println("error in landing:", err)
		}
	})
	smux.HandleFunc("/py/eval", eval.handleEvalRequest)
	sv := userMiddleware(smux)
	addr := ":8080"
	log.Println("Server started at http://127.0.0.1" + addr)
	return http.ListenAndServe(addr, sv)
}

var funcmap = template.FuncMap{
	"assetPath": func(asset string) string { return "assets/" + asset },
}
