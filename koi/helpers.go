package koi

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
					logrus.Error(errors.Wrapf(err, "Error parsing flags for %q", argumentFlags))
				} else {
					return ret
				}
			}
		}
	}
	return false
}

func copyImportantArgsIntoNewArgs(oldArgs, newArgs []string) []string {
	importantValueArgs := [][]string{
		{"-n", "--namespace"},
		{"--context"},
	}

	importantBoolArgs := [][]string{
		{"--all-namespaces"},
	}

	for _, ia := range importantValueArgs {
		val := extractValueArgumentFromArgs(oldArgs, ia...)
		if val != "" {
			newArgs = appendArgument(newArgs, ia[0], val)
		}
	}
	for _, ia := range importantBoolArgs {
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
	logrus.Tracef("going to run command: %q %q", exe, args)
	ioReader, ioWriter := io.Pipe()
	scanner := bufio.NewScanner(ioReader)

	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = ioWriter
	cmd.Stderr = os.Stderr

	runErr := cmd.Start()
	if runErr != nil {
		return 1, errors.Wrapf(runErr, "Failed to start command %q", args)
	}

	doneChan := make(chan error)
	go func() {
		doneChan <- cmd.Wait()
	}()

	keepRunning := true
	for keepRunning {
		select {
		case runErr = <-doneChan:
			logrus.Trace("Command has exited")
			keepRunning = false
		default:
			keepRunning = true
		}
		if scanner.Scan() {
			line := scanner.Text()

			printLine := true
			for _, filter := range lineFilter {
				if line, printLine = filter(line); !printLine {
					break
				}
			}
			if printLine {
				_, runErr = os.Stdout.Write([]byte(line + "\n"))
				if runErr != nil {
					return 1, errors.Wrapf(runErr, "Failed to write to stdout")
				}
			}
		}
	}

	exitCode = cmd.ProcessState.ExitCode()
	return exitCode, errors.Wrapf(runErr, "Failed to run command %q", args)
}
