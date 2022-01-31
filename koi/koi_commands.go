package koi

import "strings"

// Returns the first argument which does not start with a dash
// An empty string means no arg
func GetCommand(args []string) string {
	// Generate the argsWithArguments using this command:
	//: kubectl options | cut -d: -f1 | tr ',' '\n' | cut -d= -f1 | grep "^ *-" | awk '{printf "\"%s\", ", $1}' | pbcopy
	argsWithArguments := []string{
		"--add-dir-header", "--alsologtostderr", "--as", "--as-group", "--as-uid", "--cache-dir", "--certificate-authority", "--client-certificate", "--client-key", "--cluster", "--context", "--insecure-skip-tls-verify", "--kubeconfig", "--log-backtrace-at", "--log-dir", "--log-file", "--log-file-max-size", "--log-flush-frequency", "--logtostderr", "--match-server-version", "-n", "--namespace", "--one-output", "--password", "--profile", "--profile-output", "--request-timeout", "-s", "--server", "--skip-headers", "--skip-log-headers", "--stderrthreshold", "--tls-server-name", "--token", "--user", "--username", "-v", "--v", "--vmodule", "--warnings-as-errors",
	}

	skipNumArgs := 0

LOOPING_OVER_ARGS:
	for _, arg := range args {
		if skipNumArgs > 0 {
			skipNumArgs--
			continue
		}
		for _, check := range argsWithArguments {
			if arg == check {
				skipNumArgs = 1
				continue LOOPING_OVER_ARGS
			}

			if strings.HasPrefix(arg, check+"=") {
				continue LOOPING_OVER_ARGS
			}
		}

		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}
