package main

import (
	"embed"
	"errors"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

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
	var addr, evalDir string
	flag.StringVar(&addr, "http", ":8080", "Address on which to serve http.")
	flag.StringVar(&evalDir, "evaldir", "", "Evaluation base directory. Testdata available in source directory under \"testdata/evaluations\".")
	help := flag.Bool("help", false, "summon help")
	flag.Parse()
	if *help {
		flag.Usage()
		log.Println("help called.")
		os.Exit(0)
	}
	if evalDir == "" {
		flag.Usage()
		log.Println("evaldir flag not defined")
		os.Exit(1)
	}
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
	err = evaluator.ParseAndEvaluateGlob(path.Join(evalDir, "*.py"))
	if err != nil {
		return err
	}
	smux.Handle("/assets/", http.FileServer(http.FS(assetFS)))
	smux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/py/evals", http.StatusTemporaryRedirect)
	})
	smux.HandleFunc("/py/evals", evaluator.handleListEvaluations)

	smux.HandleFunc("/py/run/", evaluator.handleRun)
	sv := userMiddleware(smux)
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
