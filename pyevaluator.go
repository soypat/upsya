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
)

const (
	pyTimeout = 500 * time.Millisecond
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

func (ev *Evaluator) handleRun(rw http.ResponseWriter, r *http.Request) {
	type pyUserInput struct {
		Code         string
		EvaluationID string
	}
	type pyResult struct {
		Output  string
		Elapsed time.Duration
		Error   string
	}
	var src pyUserInput
	rd := io.LimitReader(r.Body, 5000) // 600Bytes read max
	err := json.NewDecoder(rd).Decode(&src)
	if err != nil {
		httpErr(rw, "", err, http.StatusBadRequest)
		return
	}
	if err := assertSafePython(src.Code); err != nil {
		json.NewEncoder(rw).Encode(pyResult{
			Error: err.Error(),
		})
		return
	}
	tempdir := "tmp"
	ev.jail.Mkdir(tempdir, 0777)
	fpath := filepath.Join(tempdir, "prog.py")
	defer ev.jail.RemoveAll(tempdir)
	fp, err := ev.jail.CreateFile(fpath)
	if err != nil {
		httpErr(rw, "creating program file", err, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(fp, strings.NewReader(src.Code))
	if err != nil {
		httpErr(rw, "copying code to file", err, http.StatusInternalServerError)
		return
	}
	eid, _ := strconv.ParseUint(src.EvaluationID, 10, 64)
	if eid > 0 {
		// Find evaluation if this is an evaluation.
		for _, eval := range ev.evals {
			if eval.ID() == uint64(eid) {
				ej := evaluationJob{
					eval:     eval,
					filename: fpath,
				}
				log.Println("running evaluation for", eid)
				ev.evaluate(r.Context(), &ej)
				json.NewEncoder(rw).Encode(ej)
				return
			}
		}
		httpErr(rw, "evaluation not found", nil, http.StatusBadRequest)
		return
	}
	log.Println("running interpreter")
	// Below is regular interpreter logic.
	ctx, cancel := context.WithTimeout(r.Context(), pyTimeout)
	defer cancel()
	cmd := ev.jail.Command(ctx, pyTimeout, "", ev.pyCommand, fpath)
	// cmd, err := p.runner.eval(ctx, th)
	if err != nil {
		httpErr(rw, "preparing program command", err, http.StatusInternalServerError)
		return
	}
	start := time.Now()
	output, err := limitCombinedOutput(cmd, 800)
	// output, err := cmd.CombinedOutput()
	result := pyResult{
		Output:  string(output),
		Elapsed: time.Since(start),
		Error: (func(err error) string {
			if err != nil {
				return err.Error()
			}
			return ""
		})(err),
	}
	json.NewEncoder(rw).Encode(result)
}

func (ev *Evaluator) evaluate(ctx context.Context, job *evaluationJob) error {
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
	comparison := "\"ours\"\n\"theirs\"\n\n"
	for i := range cases {
		cmd := ev.jail.Command(ctx, pyTimeout, "", ev.pyCommand, job.filename)
		cmd.Stdin = strings.NewReader(cases[i])
		if ctx.Err() != nil {
			return ctx.Err()
		}
		output, err := limitCombinedOutput(cmd, 1000) //cmd.CombinedOutput()
		job.outputs[i] = string(output)
		if err != nil {
			job.status[i] = -2
			if debug {
				job.outputs[i] += " " + err.Error()
			}
			continue
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
	if rd.N == 0 {
		return b.Bytes(), errors.New("too much output")
	}
	return b.Bytes(), err
}
