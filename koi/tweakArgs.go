package koi

import (
	"strings"
)

// applyTweaksToArgs modifies the args sent in so they work wtih kubectl
// Goals:
// The -x flag should become --context
func ApplyTweaksToArgs(args []string) []string {
	replacedFlags := map[string]bool{}
	shortHandReplacements := map[string]string{
		"-x": "--context",
	}

	for i, arg := range args {
		if arg == "--" {
			break
		}
		for shorthand, longform := range shortHandReplacements {
			if strings.HasPrefix(arg, shorthand+"=") || arg == shorthand {
				if _, ok := replacedFlags[shorthand]; !ok {
					args[i] = strings.Replace(arg, shorthand, longform, 1)
					replacedFlags[shorthand] = true
				}
			}
		}
	}
	return args
}
