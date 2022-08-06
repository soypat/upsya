package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

type Server struct {
	eg        EvalGroup
	evalmap   map[uint64]Evaluation
	jail      jail
	pyCommand string // command string
	// runner pyRunna
	tmpls *template.Template
	auth  *authbase
	kvdb  *bbolt.DB
}

func (sv *Server) ParseAndEvaluateGlob(pattern string) error {
	FS := os.DirFS(".")
	matches, err := fs.Glob(FS, pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return errors.New("No evaluations found with glob " + pattern)
	}

	// We obtain base directory for forming evaluation groups.
	var shortlist []string
	baseDir := ""
	minBase := math.MaxInt
	for _, match := range matches {
		dir := filepath.Dir(match)
		l := filepath.SplitList(dir)
		if len(shortlist) > 0 && len(l) == len(shortlist) && shortlist[len(shortlist)-1] != l[len(l)-1] {
			dir = filepath.Dir(dir)
			l = l[:len(l)-1]
		}
		if len(l) <= minBase {
			minBase = len(l)
			baseDir = dir
			shortlist = l
		}
	}

	evaluationMap := make(map[uint64]Evaluation)
	// Create main evalgroup.
	egroup := EvalGroup{Dir: "Main"}
	originalJail := sv.jail
	defer func() {
		sv.jail = originalJail
	}()
	sv.jail = systemPython{}
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
		err = sv.evaluate(ctx, &ej)
		if err != nil {
			return fmt.Errorf("running %s solution (%s): %s", match, ej.Error, err)
		}
		eval.results = ej.outputs
		dir := strings.TrimPrefix(filepath.Dir(match), baseDir)
		egroup.MkDirAll(dir)
		err = egroup.AddEvaluation(dir, eval)
		if err != nil {
			return err
		}
		id := eval.ID()
		if existing, ok := evaluationMap[id]; ok {
			return errors.New("evaluation " + existing.Title + " ID collides with evaluation " + eval.Title)
		}
		evaluationMap[id] = eval
	}
	sv.eg = egroup
	sv.evalmap = evaluationMap
	return nil
}

func (sv *Server) handleListEvaluations(rw http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Has("ID") {
		sv.handleEvaluation(rw, r)
		return
	}
	group := sv.eg
	urlPath := "/" + r.URL.Path
	sv.eg.Walk(func(lvl int, path string, e *EvalGroup) error {
		if path == urlPath {
			group = *e
			return errors.New("sentinel error")
		}
		return nil
	})
	sv.tmpls.Lookup("all_evaluations.tmpl").Execute(rw, struct {
		Egroup EvalGroup
	}{
		Egroup: group,
	})
}

func (sv *Server) handleEvaluation(rw http.ResponseWriter, r *http.Request) {
	num := r.URL.Query().Get("ID")
	n, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		sv.httpErr(rw, "error parsing url", err, http.StatusBadRequest)
		return
	}
	eval, ok := sv.evalmap[n]
	if !ok {
		sv.httpErr(rw, "evaluation "+num+" not found", nil, http.StatusBadRequest)
		return
	}
	u, _ := sv.auth.getUserSession(r)
	err = sv.tmpls.Lookup("evaluation.tmpl").Execute(rw, struct {
		Eval Evaluation
		User User
	}{
		Eval: eval,
		User: u,
	})
	if err != nil {
		log.Println(err)
	}
}

type EvalGroup struct {
	Dir       string
	Evals     []Evaluation
	SubGroups []EvalGroup
}

func (eg EvalGroup) String() string {
	return eg.Dir
}

func (eg *EvalGroup) AddEvaluation(dir string, eval Evaluation) error {
	var errOK = errors.New("unreachable sentinel error")
	trim := func(s string) string {
		return strings.TrimFunc(s, func(r rune) bool { return r == filepath.Separator })
	}
	err := eg.Walk(func(lvl int, path string, e *EvalGroup) error {
		if trim(path) == trim(dir) {
			e.Evals = append(e.Evals, eval)
			return errOK
		}
		return nil
	})
	if errors.Is(err, errOK) {
		return nil
	}
	return errors.New("evalgroup dir " + dir + " not found")
}

func (eg *EvalGroup) MkDirAll(path string) {
	list := filepath.SplitList(path)
	if len(list) == 0 {
		return
	}
	eg.addir(list[0], list[1:])
}

func (eg *EvalGroup) addir(newdir string, todo []string) {
	newdir = strings.TrimPrefix(newdir, "/")
	var toAdd *EvalGroup
	found := false
	for i := range eg.SubGroups {
		if eg.SubGroups[i].Dir == newdir {
			found = true
			toAdd = &eg.SubGroups[i]
			break
		}
	}
	if !found {
		eg.SubGroups = append(eg.SubGroups, EvalGroup{Dir: newdir})
		toAdd = &eg.SubGroups[len(eg.SubGroups)-1]
	}
	if len(todo) == 0 {
		return
	}
	toAdd.addir(todo[0], todo[1:])
}

// Walk recursively traverses Evalgroup and subgroups depth-first.
func (eg *EvalGroup) Walk(fn func(lvl int, path string, e *EvalGroup) error) error {
	return eg.glob(0, string(filepath.Separator), fn)
}

func (eg *EvalGroup) glob(lvl int, path string, fn func(lvl int, path string, e *EvalGroup) error) error {
	err := fn(lvl, path, eg)
	if err != nil {
		return err
	}
	for i := 0; i < len(eg.SubGroups); i++ {
		subeg := &eg.SubGroups[i]
		err = subeg.glob(lvl+1, filepath.Join(path, subeg.Dir), fn)
		if err != nil {
			return err
		}
	}
	return nil
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

func (e Evaluation) ID() (sum uint64) {
	return nchashStr(strings.Join(e.results, ""))
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
	var hash uint64 = 14695981039346656037 // Seed the hash.
	for i := 0; i < len(b); i++ {
		hash ^= uint64(b[i])
		hash *= fnvPrime
	}
	return hash
}
