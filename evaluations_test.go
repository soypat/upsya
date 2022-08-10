package main

import (
	"strings"
	"testing"
)

func TestEvaluation(t *testing.T) {
	const stdinPrefix = "\nStdin cases:\n"
	for _, e := range []Evaluation{
		{
			Title:               "Title to the work",
			Description:         "This is my Description",
			Content:             "Content and more Content\nIt's content all the way down",
			Stdin:               "sa\nds\nThis could be random text, must be perfectly preserved between serializations\n---\nOr else\n",
			Solution:            "print(\"solve\")",
			SolutionPlaceholder: "print(\"Hey now rock star\")\nprint(\"Get your game on hey now\")",
			SolutionSuffix:      "print(\"Hey now star star\")\nprint(\"Get your star power on hey now\")\n",
			SolutionPrefix:      "print(\"Hey star\")\nprint(\"prefix me this one\")\n",
		},
	} {
		var ser strings.Builder
		e.serialize(&ser)
		serialized := ser.String()
		t.Log(serialized)
		parsed, err := parseEval(strings.NewReader(serialized))
		if err != nil {
			t.Error("parsing evaluation:", err)
			continue
		}
		assertEvalEqual(t, e, parsed)
		ser.Reset()
		parsed.serialize(&ser)
		reparsed, err := parseEval(strings.NewReader(ser.String()))
		if err != nil {
			t.Error("reparsing evaluation:", err)
			continue
		}
		assertEvalEqual(t, parsed, reparsed)
	}
}

func assertEvalEqual(t *testing.T, a, b Evaluation) {
	if a.Title != b.Title {
		t.Error("title not equal")
	}
	if a.Description != b.Description {
		t.Error("content not equal")
	}
	if a.Content != b.Content {
		t.Errorf("content not equal\n%q\n%q\n", a.Content, b.Content)
	}
	if a.Solution != b.Solution {
		t.Errorf("solution not equal\n%q\n%q\n", a.Solution, b.Solution)
	}
	if a.Stdin != b.Stdin {
		t.Errorf("stdin not equal\n%q\n%q\n", a.Stdin, b.Stdin)
	}
	if a.SolutionPlaceholder != b.SolutionPlaceholder {
		t.Errorf("solution placeholder not equal\n%q\n%q\n", a.SolutionPlaceholder, b.SolutionPlaceholder)
	}
	if a.SolutionSuffix != b.SolutionSuffix {
		t.Errorf("solution suffix not equal\n%q\n%q\n", a.SolutionSuffix, b.SolutionSuffix)
	}
}
