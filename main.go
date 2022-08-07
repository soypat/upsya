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
	"go.etcd.io/bbolt"
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
		os.Exit(0) // finish succesfully in this case.
	}
	if len(flag.Args()) > 1 {
		flag.Usage()
		return fmt.Errorf("got too many arguments: %s", flag.Args())
	}
	if evalGlob == "" {
		flag.Usage()
		return fmt.Errorf("evalglob flag not defined")
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
	server := Server{
		tmpls:     tmpl,
		pyCommand: python,
		auth:      &authbase{},
	}
	if os.Getenv("GONTAINER_FS") == "" {
		server.jail = systemPython{}
	} else {
		server.jail = container{
			chrtDir: os.Getenv("GONTAINER_FS"),
			workDir: "/home",
		}
	}
	err = server.jail.MkdirAll("tmp", 0777)
	if err != nil {
		return fmt.Errorf("creating tmp directory in jail: %s", err)
	}
	err = server.ParseAndEvaluateGlob(evalGlob)
	if err != nil {
		return err
	}
	db, err := bbolt.Open("kvdb.bbolt", 0777, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("evalresults"))
		return err
	})
	if err != nil {
		return err
	}
	server.kvdb = db
	server.submitted = make(map[uint64]struct{})
	// Set endpoints.
	smux.Handle("/assets/", http.FileServer(http.FS(assetFS)))
	smux.Handle("/py/evals/", http.StripPrefix("/py/evals/", http.HandlerFunc(server.handleListEvaluations)))
	smux.HandleFunc("/py/run/", server.handleRun)
	smux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/py/evals/", http.StatusTemporaryRedirect)
	})
	smux.HandleFunc("/auth/", server.handleAuth)

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
