package main

import (
	"strings"
	"testing"
)

func TestEvaluation(t *testing.T) {
	const stdinPrefix = "\nStdin cases:\n"
	for _, e := range []Evaluation{
		{
			Title:       "\n MEga c \n ",
			Description: "\n MEga c \n ",
			Content:     "\n MEga c \n ",
			Stdin:       "sa\nds\n",
			Solution:    "\n MEga sadad\n",
		},
	} {
		var ser strings.Builder
		e.serialize(&ser)
		parsed, err := parseEval(strings.NewReader(ser.String()))
		if err != nil {
			t.Error("parsing evaluation:", err)
			continue
		}
		ser.Reset()
		parsed.serialize(&ser)
		reparsed, err := parseEval(strings.NewReader(ser.String()))
		if err != nil {
			t.Error("reparsing evaluation:", err)
			continue
		}
		assetEvalEqual(t, parsed, reparsed)
	}
}

func assetEvalEqual(t *testing.T, a, b Evaluation) {
	if a.Title != b.Title {
		t.Error("title not equal")
	}
	if a.Description != b.Description {
		t.Error("content not equal")
	}
	if a.Content != b.Content {
		t.Errorf("content not equal\n%q\n%q", a.Content, b.Content)
	}
	if a.Solution != b.Solution {
		t.Errorf("solution not equal\n%q\n%q", a.Solution, b.Solution)
	}
	if a.Stdin != b.Stdin {
		t.Errorf("stdin not equal\n%q\n%q", a.Stdin, b.Stdin)
	}
}
