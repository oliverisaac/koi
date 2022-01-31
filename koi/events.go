package koi

import "regexp"

func EventsCommand(exe string, args []string) (exitCode int, runError error) {
	cmdArg := []string{
		"get",
		"events",
		"--sort-by=.metadata.creationTimestamp",
	}

	cmdArg = copyImportantArgsIntoNewArgs(args, cmdArg)

	filterRegex := regexp.MustCompile(`\bNormal\b`)
	filter := func(line string) (string, bool) {
		return line, !filterRegex.MatchString(line)
	}

	return runCommandAndFilterOutput(exe, cmdArg, filter)
}
