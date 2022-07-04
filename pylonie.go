package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

type pyRunna struct {
	eval func(ctx context.Context, progPath string) (*exec.Cmd, error)
}

func cmdRunna(command string) pyRunna {
	return pyRunna{
		func(ctx context.Context, progPath string) (*exec.Cmd, error) {
			stat, err := os.Stat(progPath)
			if err != nil {
				return nil, err
			}
			if stat.IsDir() {
				return nil, errors.New("python program path is directory")
			}
			cmd := exec.CommandContext(ctx, command, path.Base(progPath))
			cmd.Dir = path.Dir(progPath)
			return cmd, nil
		},
	}
}

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

func (p *Evaluator) handleRun(rw http.ResponseWriter, r *http.Request) {
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
	os.Mkdir("tmp", 0777)
	tempdir, err := os.MkdirTemp("tmp/", "*")
	fpath := path.Join(tempdir, "prog.py")
	if err != nil {
		httpErr(rw, "creating temp dir", err, http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempdir)
	fp, err := os.Create(fpath)
	if err != nil {
		httpErr(rw, "creating program file", err, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(fp, strings.NewReader(src.Code))
	if err != nil {
		httpErr(rw, "copying code to file", err, http.StatusInternalServerError)
		return
	}
	eid, _ := strconv.Atoi(src.EvaluationID)
	if eid > 0 {
		// Find evaluation if this is an evaluation.
		for _, ev := range p.evals {
			if ev.ID() == int64(eid) {
				ej := evaluationJob{
					eval:     ev,
					filename: fpath,
				}
				log.Println("running evaluation for", eid)
				p.evaluate(r.Context(), &ej)
				json.NewEncoder(rw).Encode(ej)
				return
			}
		}
		httpErr(rw, "evaluation not found", nil, http.StatusBadRequest)
		return
	}
	log.Println("running interpreter")
	// Below is regular interpreter logic.
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()
	cmd, err := p.runner.eval(ctx, fpath)
	if err != nil {
		httpErr(rw, "preparing program command", err, http.StatusInternalServerError)
		return
	}
	start := time.Now()
	output, err := cmd.CombinedOutput()
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
	cases := strings.Split(job.eval.Stdin, "---")
	job.status = make([]int, len(cases))
	job.outputs = make([]string, len(cases))
	start := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
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
		cmd, err := ev.runner.eval(ctx, job.filename)
		if err != nil {
			return err
		}
		cmd.Stdin = strings.NewReader(cases[i])
		if ctx.Err() != nil {
			return ctx.Err()
		}
		output, err := cmd.CombinedOutput()
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
