package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/oliverisaac/koi/koi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var version string = "unset-version"
var commit string = "unset-commit"

func main() {
	var exitCode int
	var err error

	logLevel := defaultEnv("KOI_LOG_LEVEL", "INFO")
	if ll, err := logrus.ParseLevel(logLevel); err == nil {
		logrus.SetLevel(ll)
	}

	exe := defaultEnv("KOI_KUBECTL_EXE", "kubectl")
	koiArgs, filterExe, filterCommand := koi.ApplyTweaksToArgs(os.Args[1:])

	requestedKoiCommand := koi.GetCommand(koiArgs)
	logrus.Debugf("Requested command: %s", requestedKoiCommand)
	if requestedKoiCommand == "events" {
		exitCode, err = koi.EventsCommand(exe, koiArgs)
	} else if requestedKoiCommand == "fish" {
		exitCode, err = koi.FishCommand(exe, koiArgs)
	} else if requestedKoiCommand == "version" {
		fmt.Printf("Koi version: %s (%s)\n", version, commit)
		exitCode, err = runAttachedCommand(exe, filterExe, filterCommand, koiArgs)
	} else if requestedKoiCommand == "export" {
		exitCode, err = koi.ExportCommand(os.Stdin, os.Stdout)
	} else if requestedKoiCommand == "shell" {
		koiArgs = removeArg(koiArgs, "shell")
		exitCode, err = koi.ShellCommand(exe, koiArgs)
	} else if requestedKoiCommand == "containers" {
		koiArgs = removeArg(koiArgs, "containers")
		exitCode, err = koi.ContainersCommand(koiArgs)
	} else {
		exitCode, err = runAttachedCommand(exe, filterExe, filterCommand, koiArgs)
	}

	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to run the command"))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func removeArg(args []string, arg string) []string {
	for i, a := range args {
		if a == arg {
			return append(args[:i], args[i+1:]...)
		}
	}
	return args
}

func defaultEnv(env string, defaultVal string) string {
	if val, ok := os.LookupEnv(env); ok {
		return val
	}
	return defaultVal
}

func runAttachedCommand(command string, filterExe string, filterCommand string, args []string) (exitCode int, runErr error) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	var yqCmd *exec.Cmd
	var cmdOut *io.PipeWriter
	if filterExe != "" {
		var yqIn *io.PipeReader
		yqIn, cmdOut = io.Pipe()
		filterArgs := []string{}
		if filterExe == "yq" {
			filterArgs = append(filterArgs, "-P")
		} else if filterExe == "jq" {
			filterArgs = append(filterArgs, "-r")
		}
		filterArgs = append(filterArgs, filterCommand)
		yqCmd = exec.Command(filterExe, filterArgs...)

		yqCmd.Stdin = yqIn
		yqCmd.Stdout = os.Stdout
		yqCmd.Stderr = os.Stderr

		cmd.Stdout = cmdOut
		cmd.Stderr = os.Stderr
	}

	if yqCmd != nil {
		yqErr := yqCmd.Start()
		if yqErr != nil {
			return 1, errors.Wrapf(yqErr, "Failed to start yq command")
		}
	}

	cmdErr := cmd.Start()
	if cmdErr != nil {
		return 1, errors.Wrapf(cmdErr, "Failed to start command %q %q", command, args)
	}

	ps, cmdErr := cmd.Process.Wait()

	if yqCmd != nil {
		cmdOut.Close()
		_, yqErr := yqCmd.Process.Wait()
		if yqErr != nil {
			return 1, errors.Wrapf(yqErr, "Failed to run yq command")
		}
	}

	exitCode = ps.ExitCode()
	return exitCode, errors.Wrapf(cmdErr, "Failed to run command %q %q", command, args)
}
