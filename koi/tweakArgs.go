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

// applyTweaksToArgs modifies the args sent in so they work with kubectl
// Goals:
// The -x flag should become --context
func ApplyTweaksToArgs(args []string) ([]string, string, string) {
	filterExe := ""
	filterCommand := ""

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

	finalArgs := []string{}
	endOfKoiArgs := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			endOfKoiArgs = true
		}

		if endOfKoiArgs {
			finalArgs = append(finalArgs, arg)
			continue
		}

		// If any of the args match the shorthand, then we set it to the longform
		for shorthand, longform := range shortHandReplacements {
			if strings.HasPrefix(arg, shorthand+"=") || arg == shorthand {
				if _, ok := replacedFlags[shorthand]; !ok {
					arg = strings.Replace(arg, shorthand, longform, 1)
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

		if arg == "-o" || arg == "--output" || strings.HasPrefix(arg, "--output=") || strings.HasPrefix(arg, "-o=") {
			var outputFormat string
			if strings.Contains(arg, "=") {
				outputFormat = strings.SplitN(arg, "=", 2)[1]
			} else {
				if i+1 < len(args) {
					outputFormat = args[i+1]
					i = i + 1
				}
			}
			if strings.HasPrefix(outputFormat, "jq") || strings.HasPrefix(outputFormat, "yq") {
				if strings.Contains(outputFormat, "=") {
					arg = "--" + outputFormat
				} else {
					arg = "--" + outputFormat + "=."
				}
			} else {
				arg = "--output=" + outputFormat
			}
		}

		if arg == "--yq" || arg == "--jq" {
			filterExe = strings.TrimPrefix(arg, "--")
			filterCommand = ""
			if i+1 < len(args) {
				filterCommand = args[i+1]
				if filterCommand == "--" {
					filterCommand = ""
				}
			}
			if filterCommand != "" {
				i = i + 1
			} else {
				filterCommand = "."
			}

			finalArgs = append(finalArgs, "--output=json")
			continue
		} else if strings.HasPrefix(arg, "--yq=") || strings.HasPrefix(arg, "--jq=") {
			filterConfig := strings.SplitN(arg, "=", 2)
			filterExe = strings.TrimPrefix(filterConfig[0], "--")
			filterCommand = filterConfig[1]
			finalArgs = append(finalArgs, "--output=json")
			continue
		}

		finalArgs = append(finalArgs, arg)
	}

	for _, dv := range defaultValuesForFlags {
		if dv.value != "" && !dv.alreadySet {
			finalArgs = appendArgument(finalArgs, dv.flagsThatMatch[0], dv.value)
		}
	}

	return finalArgs, filterExe, filterCommand
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
