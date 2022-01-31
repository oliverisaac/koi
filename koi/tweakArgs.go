package koi

import (
	"os"
	"strings"
)

type defaultValueMapping struct {
	value          string
	alreadySet     bool
	flagsThatMatch []string
}

// applyTweaksToArgs modifies the args sent in so they work wtih kubectl
// Goals:
// The -x flag should become --context
func ApplyTweaksToArgs(args []string) []string {
	replacedFlags := map[string]bool{}
	shortHandReplacements := map[string]string{
		"-x": "--context",
	}

	defaultValuesForFlags := []*defaultValueMapping{
		{
			value: coalesceString(os.Getenv("KOI_CONTEXT"), os.Getenv("KOI_CTX")),
			flagsThatMatch: []string{
				"--context",
				"-x",
			},
		},
		{
			value: coalesceString(os.Getenv("KOI_NAMESPACE"), os.Getenv("KOI_NS")),
			flagsThatMatch: []string{
				"--namespace",
				"-n",
			},
		},
	}

	lastArgPos := 0

	for i, arg := range args {
		lastArgPos = i
		if arg == "--" {
			break
		}

		// If any of the args match teh shorthand, then we set it to the longform
		for shorthand, longform := range shortHandReplacements {
			if strings.HasPrefix(arg, shorthand+"=") || arg == shorthand {
				if _, ok := replacedFlags[shorthand]; !ok {
					args[i] = strings.Replace(arg, shorthand, longform, 1)
					arg = args[i]
					replacedFlags[shorthand] = true
				}
			}
		}
		// If any of the args match one of the defautl values, then we don't need to apply them
		for _, dv := range defaultValuesForFlags {
			if !dv.alreadySet && dv.value != "" && stringArrayContains(dv.flagsThatMatch, arg) {
				dv.alreadySet = true
			}
		}
	}

	outputArgs := make([]string, 0, len(args)+len(defaultValuesForFlags))

	outputArgs = append(outputArgs, args[0:lastArgPos]...)
	for _, dv := range defaultValuesForFlags {
		if dv.value != "" && !dv.alreadySet {
			outputArgs = append(outputArgs, dv.flagsThatMatch[0], dv.value)
		}
	}
	outputArgs = append(outputArgs, args[lastArgPos:]...)

	return outputArgs
}

func stringArrayContains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func coalesceString(check ...string) string {
	for _, v := range check {
		if v != "" {
			return v
		}
	}
	return ""
}
