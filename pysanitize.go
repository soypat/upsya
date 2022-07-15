package main

import (
	"fmt"
	"regexp"
	"strings"
)

// reForbid sanitization structures
var reForbid = map[*regexp.Regexp]string{
	regexp.MustCompile(`exec|eval|globals|locals|write|breakpoint|getattr|memoryview|vars|super`): "forbidden function key '%s'",
	//regexp.MustCompile(`input\s*\(`):                           "no %s) to parse!",
	regexp.MustCompile(`tofile|savetxt|fromfile|fromtxt|load\s*\(`): "forbidden numpy function key '%s'",
	regexp.MustCompile("dump"):                                      "forbidden json function key '%s'",
	regexp.MustCompile("to_csv|to_json|to_html|to_clipboard|to_excel|to_hdf|to_feather|to_parquet|to_msgpack|to_stata|to_pickle|to_sql|to_gbq"): "forbidden pandas function key '%s'",
	regexp.MustCompile(`__\w+__`): "forbidden dunder function key '%s'",
}

// special treatment for imports since we may allow special imports such as math, numpy, pandas
var reImport = regexp.MustCompile(`^from[\s]+[\w]+|import[\s]+[\w]+`)

var allowedImports = map[string]bool{
	"math":       true,
	"numpy":      true,
	"pandas":     true,
	"json":       true,
	"itertools":  false,
	"processing": false,
	"os":         false,
}

func assertSafePython(sourceCode string) error {
	const pyMaxSourceLength = 1200
	if len(sourceCode) > pyMaxSourceLength {
		return fmt.Errorf("code snippet too long (%d/%d)", len(sourceCode), pyMaxSourceLength)
	}
	semicolonSplit := strings.Split(sourceCode, ";")
	newLineSplit := strings.Split(sourceCode, "\n")
	for _, v := range append(semicolonSplit, newLineSplit...) {
		for re, errF := range reForbid {
			str := re.FindString(strings.TrimSpace(v))
			if str != "" {
				return fmt.Errorf(errF, str)
			}
		}
		str := reImport.FindString(strings.TrimSpace(v))
		if str != "" {
			words := strings.Split(str, " ")
			if len(words) < 2 {
				return fmt.Errorf("unexpected import formatting: %s", str)
			}
			allowed, present := allowedImports[strings.TrimSpace(words[1])]
			if !present {
				return fmt.Errorf("import '%s' not in safelist:\n%s", strings.TrimSpace(words[1]), printSafeList())
			}
			if !allowed {
				return fmt.Errorf("forbidden import '%s'", strings.TrimSpace(words[1]))
			}
		}
	}
	return nil
}

// printSafeList shows user what imports can
// be used in interpreter
func printSafeList() (s string) {
	counter := 0
	for k, v := range allowedImports {
		if v {
			counter++
			if counter > 1 {
				s += ",  "
			}
			s += k
		}
	}
	return s
}
