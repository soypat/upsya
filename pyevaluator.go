package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

const (
	pyTimeout = 500 * time.Millisecond
	// Maximum length of stdout output of evaluations
	pyMaxStdoutLen = 1200
)

type evaluationJob struct {
	eval     Evaluation
	filename string
	// case status 0: did not run, -2: program error, -1: incorrect result, 1: correct result.
	status  []int
	outputs []string
	Output  string
	Elapsed time.Duration
	Error   string
}

type PyUserInput struct {
	Code         string
	EvaluationID string
	UserID       uint64
}

func (sv *Server) handleRun(rw http.ResponseWriter, r *http.Request) {
	var src PyUserInput
	rd := io.LimitReader(r.Body, 5000) // 600Bytes read max
	err := json.NewDecoder(rd).Decode(&src)
	if err != nil {
		sv.httpErr(rw, "", err, http.StatusBadRequest)
		return
	}
	// Check if user logged in.
	user, err := sv.auth.getUserSession(r)
	if err != nil {
		json.NewEncoder(rw).Encode(struct{ Error, Output string }{
			Error: "Authentication error. Please login again",
		})
		return
	}
	// Check if code has been submitted already.
	hash := nchashStr(src.Code) + user.ID
	_, alreadySubmitted := sv.submitted[hash]
	if alreadySubmitted {
		json.NewEncoder(rw).Encode(struct{ Error, Output string }{
			Error: "You can't submit two identical programs.",
		})
		return
	}
	sv.submitted[hash] = struct{}{}
	// Check if python is "safe" for running.
	if err := assertSafePython(src.Code); err != nil {
		json.NewEncoder(rw).Encode(struct{ Error, Output string }{
			Error: err.Error(),
		})
		return
	}
	// Prepare jail directory.
	wd, err := sv.jail.Getwd()
	if err != nil {
		log.Println(err)
	}
	tempdir := filepath.Join(wd, "tmp")
	err = sv.jail.MkdirAll(tempdir, 0777)
	if err != nil {
		log.Println(err)
	}
	fpath := filepath.Join(tempdir, "prog.py")
	defer sv.jail.RemoveAll(tempdir)
	fp, err := sv.jail.CreateFile(fpath)
	if err != nil {
		sv.httpErr(rw, "creating program file", err, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(fp, strings.NewReader(src.Code))
	if err != nil {
		sv.httpErr(rw, "copying code to file", err, http.StatusInternalServerError)
		return
	}
	eid, _ := strconv.ParseUint(src.EvaluationID, 10, 64)
	if eid > 0 {
		// Find evaluation if this is an evaluation.
		for _, eval := range sv.evalmap {
			if eval.ID() == uint64(eid) {
				ej := evaluationJob{
					eval:     eval,
					filename: fpath,
				}
				log.Println("running evaluation for", eid)
				sv.evaluate(r.Context(), &ej)
				json.NewEncoder(rw).Encode(ej)
				go sv.saveEvaluationResult(src, ej) // non critical task
				return
			}
		}
		sv.httpErr(rw, "evaluation not found", nil, http.StatusBadRequest)
		return
	}
	log.Println("running interpreter")
	// Below is regular interpreter logic.
	ctx, cancel := context.WithTimeout(r.Context(), pyTimeout)
	defer cancel()
	cmd := sv.jail.Command(ctx, pyTimeout, "", sv.pyCommand, fpath)
	// cmd, err := p.runner.eval(ctx, th)
	if err != nil {
		sv.httpErr(rw, "preparing program command", err, http.StatusInternalServerError)
		return
	}
	start := time.Now()
	output, err := limitCombinedOutput(cmd, pyMaxStdoutLen)
	result := struct {
		Output  string
		Elapsed time.Duration
		Error   string
	}{
		Output:  string(output),
		Elapsed: time.Since(start),
		Error: func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	}
	json.NewEncoder(rw).Encode(result)
}

func (ev *Server) evaluate(ctx context.Context, job *evaluationJob) error {
	cases := job.eval.StdinCases()
	job.status = make([]int, len(cases))
	job.outputs = make([]string, len(cases))
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, pyTimeout)
	setJob := func(output string, err error) {
		job.Output = output
		job.Elapsed = time.Since(start)
		if err != nil {
			job.Error = err.Error()
		}
	}
	defer cancel()
	correct := 0
	comparison := "\"ours\"\n\"yours\"\n\n"
	for i := range cases {
		cmd := ev.jail.Command(ctx, pyTimeout, "", ev.pyCommand, job.filename)
		cmd.Stdin = strings.NewReader(cases[i])
		if ctx.Err() != nil {
			return ctx.Err()
		}
		output, err := limitCombinedOutput(cmd, pyMaxStdoutLen) //cmd.CombinedOutput()
		job.outputs[i] = string(output)
		if err != nil {
			if debug {
				job.status[i] = -2
				if debug {
					job.outputs[i] += " " + err.Error()
				}
				continue
			}
			setJob(string(output), err)
			return nil
		}
		if job.outputs[i] == job.eval.results[i] {
			correct++
			job.status[i] = 1
		} else {
			job.status[i] = -1
		}
	}
	if len(cases) == correct {
		setJob(fmt.Sprintf("all %d cases passed! ", correct), nil)
		return nil
	}
	if debug {
		// generate debug info
		for i := range cases {
			comparison += fmt.Sprintf("%q\n%q %d\n\n", job.eval.results[i], job.outputs[i], job.status[i])
		}
	}
	msg := fmt.Sprintf("%d/%d cases passed", correct, len(cases))
	if debug {
		msg += "\n" + comparison
	}
	setJob(msg, errors.New("Did not pass all cases"))
	return nil
}

const (
	timeKeyPadding = ".0000"
	timeKeyFormat  = "2006-01-02 15:04:05.9999"
)

func boltKey(t time.Time) []byte {
	// RFC3339 format allows for sortable keys. See https://github.com/etcd-io/bbolt#range-scans.
	key := []byte(t.Format(timeKeyFormat))
	diff := len(timeKeyFormat) - len(key)
	key = append(key, timeKeyPadding[len(timeKeyPadding)-diff:]...)
	return key
}

// saveEvaluationResult updates the server database with the python source code and evaluation results.
func (sv *Server) saveEvaluationResult(src PyUserInput, job evaluationJob) {
	defer func() {
		a := recover()
		if a != nil {
			log.Printf("CRITICAL: panic in saveEvaluationResult: %v", a)
		}
	}()
	err := sv.kvdb.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("evalresults"))
		if b == nil {
			return errors.New("evaluation result bucket not exist")
		}
		// Discarding error: Marshal fails in rare edge cases, we are assured it won't fail here.
		value, _ := json.Marshal(struct {
			Source PyUserInput
			Job    evaluationJob
		}{Source: src, Job: job})
		return b.Put(boltKey(time.Now()), value)
	})
	if err != nil {
		log.Printf("error saving evaluation: %v", err)
	}
}

// limitCombinedOutput returns the combined Stdout and Stderr output of
// the command while limiting it to n bytes.
func limitCombinedOutput(cmd *exec.Cmd, n int64) (output []byte, err error) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	defer stdout.Close()
	defer stderr.Close()
	rd := &io.LimitedReader{
		R: io.MultiReader(stdout, stderr),
		N: n,
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	defer cmd.Process.Kill()
	var b bytes.Buffer
	_, err = io.Copy(&b, rd)
	if err != nil {
		return nil, err
	}
	if rd.N == 0 {
		return b.Bytes(), errors.New("too much output")
	}
	return b.Bytes(), cmd.Wait()
}
