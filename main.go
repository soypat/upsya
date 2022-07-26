package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

var (
	// debug mode enabled/disabled
	debug bool
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
	var addr, evalGlob string
	flag.StringVar(&addr, "http", ":8080", "Address on which to serve http.")
	flag.StringVar(&evalGlob, "evalglob", "", "Evaluation base directory. Testdata available in source directory under \"testdata/evaluations/*/*.py\".")
	flag.BoolVar(&debug, "debug", false, "Enable debugging mode with extra help for users.")
	help := flag.Bool("help", false, "summon help")
	flag.Parse()
	if *help {
		flag.Usage()
		log.Println("help called.")
		os.Exit(0)
	}
	if len(flag.Args()) > 1 {
		flag.Usage()
		log.Println("got too many arguments:", flag.Args())
		os.Exit(1)
	}
	if evalGlob == "" {
		flag.Usage()
		log.Println("evalglob flag not defined")
		os.Exit(1)
	}
	if debug {
		log.Println("debug mode enabled")
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
	evaluator := Server{
		tmpls:     tmpl,
		pyCommand: python,
		auth:      newauthbase(),
	}
	if os.Getenv("GONTAINER_FS") == "" {
		evaluator.jail = systemPython{}
	} else {
		evaluator.jail = container{
			chrtDir: os.Getenv("GONTAINER_FS"),
			workDir: "/home",
		}
	}
	err = evaluator.jail.MkdirAll("tmp", 0777)
	if err != nil {
		return fmt.Errorf("creating tmp directory in jail: %s", err)
	}
	err = evaluator.ParseAndEvaluateGlob(evalGlob)
	if err != nil {
		return err
	}
	// Set endpoints.
	smux.Handle("/assets/", http.FileServer(http.FS(assetFS)))

	smux.Handle("/py/evals/", http.StripPrefix("/py/evals/", http.HandlerFunc(evaluator.handleListEvaluations)))
	smux.HandleFunc("/py/run/", evaluator.handleRun)
	smux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/py/evals/", http.StatusTemporaryRedirect)
	})
	smux.HandleFunc("/auth/", evaluator.handleAuth)
	// Wrapping middleware for all http requests.
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
