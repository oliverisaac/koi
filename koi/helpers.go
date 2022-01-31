package koi

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func extractValueArgumentFromArgs(args []string, argumentFlags ...string) string {
	for i, a := range args {
		if a == "--" {
			break
		}
		for _, f := range argumentFlags {
			if a == f && len(args) > (i+1) {
				return args[i+1]
			}
			if strings.HasPrefix(a, f+"=") {
				return strings.SplitN(a, "=", 2)[1]
			}
		}
	}
	return ""
}

func extractBoolArgumentFromArgs(args []string, argumentFlags ...string) bool {
	for _, a := range args {
		if a == "--" {
			break
		}
		for _, f := range argumentFlags {
			if a == f {
				return true
			}
			if strings.HasPrefix(a, f+"=") {
				val := strings.SplitN(a, "=", 2)[1]
				ret, err := strconv.ParseBool(val)
				if err != nil {
					log.Error(errors.Wrapf(err, "Error parsing flags for %q", argumentFlags))
				} else {
					return ret
				}
			}
		}
	}
	return false
}

func copyArgsIntoNewArgs(oldArgs, newArgs []string, copyValueArgs, copyBoolArgs [][]string) []string {
	for _, ia := range copyValueArgs {
		val := extractValueArgumentFromArgs(oldArgs, ia...)
		if val != "" {
			newArgs = appendArgument(newArgs, ia[0], val)
		}
	}
	for _, ia := range copyBoolArgs {
		val := extractBoolArgumentFromArgs(oldArgs, ia...)
		if val {
			newArgs = appendArgument(newArgs, ia[0], "")
		}
	}
	return newArgs
}

func appendArgument(args []string, flag string, val string) []string {
	insertionPoint := 0
	for i, a := range args {
		insertionPoint = i
		if a == "--" {
			break
		}
	}

	ret := make([]string, 0, len(args)+2)
	ret = append(ret, args[0:insertionPoint]...)

	if val == "" {
		ret = append(ret, flag)
	} else {
		ret = append(ret, flag, val)
	}

	ret = append(ret, args[insertionPoint:]...)
	return ret
}

func runCommandAndFilterOutput(exe string, args []string, lineFilter ...func(line string) (modifiedLine string, shouldPrint bool)) (exitCode int, runError error) {
	log.Tracef("going to run command: %q %q", exe, args)

	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, errors.Wrapf(err, "Failed to create stdout pipe")
	}
	scanner := bufio.NewScanner(stdout)

	runErr := cmd.Start()
	if runErr != nil {
		return 1, errors.Wrapf(runErr, "Failed to start command %q", args)
	}

	log.Trace("Start scanning")
	for scanner.Scan() {
		line := scanner.Text()
		printLine := true
		for _, filter := range lineFilter {
			if filter != nil {
				if line, printLine = filter(line); !printLine {
					break
				}
			}
		}
		if printLine {
			_, runErr = os.Stdout.Write([]byte(line + "\n"))
			if runErr != nil {
				return 1, errors.Wrapf(runErr, "Failed to write to stdout")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 1, errors.Wrapf(err, "Failed to scan")
	}

	log.Trace("Done scanning")

	runErr = cmd.Wait()
	exitCode = cmd.ProcessState.ExitCode()
	return exitCode, errors.Wrapf(runErr, "Failed to run command %q", args)
}
