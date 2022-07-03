package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
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

func (p *pyRunna) handleEvalRequest(rw http.ResponseWriter, r *http.Request) {
	type pyUserInput struct {
		Code         string
		EvaluationID int
	}
	type pyResult struct {
		Output  string
		Elapsed time.Duration
		Error   string
	}

	var src pyUserInput
	rd := io.LimitReader(r.Body, 600) // 600Bytes read max
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
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()
	cmd, err := p.eval(ctx, fpath)
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
