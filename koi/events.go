package koi

import "regexp"

func EventsCommand(exe string, args []string) (exitCode int, runError error) {
	cmdArg := []string{
		"get",
		"events",
		"--sort-by=.metadata.creationTimestamp",
	}

	cmdArg = copyArgsIntoNewArgs(args,
		cmdArg,
		[][]string{
			{"-n", "--namespace"},
			{"-o", "--output"},
			{"--context"},
		},
		[][]string{
			{"--all-namespaces"},
		},
	)

	filterRegex := regexp.MustCompile(`\bNormal\b`)
	filter := func(line string) (string, bool) {
		return line, !filterRegex.MatchString(line)
	}

	// If the output is set, then we don't want to do any filtering
	if val := extractValueArgumentFromArgs(cmdArg, "-o", "--output"); val != "" {
		filter = nil
	}

	return runCommandAndFilterOutput(exe, cmdArg, filter)
}
