package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Evaluator struct {
	evals  []Evaluation
	runner pyRunna
	tmpls  *template.Template
}

func (e *Evaluator) ParseAndEvaluateGlob(pattern string) error {
	FS := os.DirFS(".")
	matches, err := fs.Glob(FS, pattern)
	if err != nil {
		return err
	}
	for _, match := range matches {
		p, _ := FS.Open(match)
		if isDir(p) {
			continue
		}
		eval, err := parseEval(p)
		if err != nil {
			return fmt.Errorf("parsing file %s: %s", match, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2000*time.Millisecond)
		defer cancel()
		eval.results = make([]string, len(strings.Split(eval.Stdin, "---")))
		ej := evaluationJob{
			eval:     eval,
			filename: match,
		}
		err = e.evaluate(ctx, &ej)
		if err != nil {
			return fmt.Errorf("running %s solution (%s): %s", match, ej.Error, err)
		}
		eval.results = ej.outputs
		e.evals = append(e.evals, eval)
	}
	return nil
}

func (e *Evaluator) handleListEvaluations(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("ID") {
		e.handleEvaluation(rw, r)
		return
	}
	e.tmpls.Lookup("all_evaluations.tmpl").Execute(rw, e.evals)
}

func (e *Evaluator) handleEvaluation(rw http.ResponseWriter, r *http.Request) {
	num := r.URL.Query().Get("ID")
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		httpErr(rw, "error parsing url", err, http.StatusBadRequest)
		return
	}
	for _, ev := range e.evals {
		if n == ev.ID() {
			err := e.tmpls.Lookup("evaluation.tmpl").Execute(rw, &ev)
			if err != nil {
				log.Println(err)
			}
			return
		}
	}
}

type Evaluation struct {
	Title       string
	Description string
	Content     string
	Stdin       string
	Solution    string
	// Results is the standard output of the solution for each of the
	// standard input test cases.
	results []string
}

func (e *Evaluation) ID() (sum uint64) {
	return nchashStr(e.Title)
}

func (eval Evaluation) serialize(w io.Writer) (err error) {
	const stdinPrefix = "Stdin cases:"
	_, err = w.Write([]byte("\"\"\"\n" + eval.Title + "\n" + eval.Description + "\n===\n"))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(eval.Content + "\n\"\"\"\n" + eval.Solution))
	if err != nil {
		return err
	}
	if eval.Stdin != "" {
		_, err = w.Write([]byte("\n\"\"\"\n" + stdinPrefix + "\n" + eval.Stdin + "\"\"\"\n"))
	}
	return err
}

func parseEval(r io.Reader) (eval Evaluation, err error) {
	const stdinPrefix = "Stdin cases:"
	var s strings.Builder
	_, err = io.Copy(&s, r)
	if err != nil {
		return eval, err
	}
	splits := strings.Split(strings.ReplaceAll(s.String(), "\r", ""), `"""`)
	if len(splits) < 3 {
		return eval, errors.New("docstrings not found")
	}
	if len(splits) >= 4 {
		splits[3] = strings.TrimLeft(splits[3], " \n")
		if !strings.HasPrefix(splits[3], stdinPrefix) {
			return eval, fmt.Errorf("second docstring must be input and start with %q", stdinPrefix)
		}
		eval.Stdin = strings.TrimSpace(strings.TrimPrefix(splits[3], stdinPrefix))
		if strings.HasSuffix(splits[3], "\n") {
			eval.Stdin += "\n" // if last line had a newline add it again.
		}
	}
	wholeContent := splits[1]
	eval.Solution = strings.TrimSpace(splits[2])
	frontmatter, content, hasFront := strings.Cut(wholeContent, "===")
	eval.Content = strings.TrimSpace(content)
	if hasFront {
		frontmatter = strings.TrimSpace(frontmatter)
		title, description, hasDesc := strings.Cut(frontmatter, "\n")
		eval.Title = title
		if hasDesc {
			eval.Description = strings.TrimSpace(description)
		}
	}
	return eval, nil
}

func isDir(f fs.File) bool {
	if p, ok := f.(fs.ReadDirFile); ok {
		_, err := p.ReadDir(1)
		return err == nil
	}
	return false
}

func (e *Evaluation) StdinCases() []string {
	return strings.Split(e.Stdin, "---\n")
}

func nchashStr(s string) uint64 {
	return nchash([]byte(s))
}

// nchash is a non-cryptographic hash function.
func nchash(b []byte) uint64 {
	// Fowler-Noll-Vo (FNV) hash function.
	const fnvPrime = 1099511628211
	var hash uint64 = 14695981039346656037
	for i := 0; i < len(b); i++ {
		hash ^= uint64(b[i])
		hash *= fnvPrime
	}
	return hash
}
