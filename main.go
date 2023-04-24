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
	koiArgs := koi.ApplyTweaksToArgs(os.Args[1:])

	requestedKoiCommand := koi.GetCommand(koiArgs)
	if requestedKoiCommand == "events" {
		exitCode, err = koi.EventsCommand(exe, koiArgs)
	} else if requestedKoiCommand == "fish" {
		exitCode, err = koi.FishCommand(exe, koiArgs)
	} else if requestedKoiCommand == "version" {
		fmt.Printf("Koi version: %s (%s)\n", version, commit)
		exitCode, err = runAttachedCommand(exe, koiArgs...)
	} else if requestedKoiCommand == "export" {
		exitCode, err = koi.ExportCommand(os.Stdin, os.Stdout)
	} else {
		exitCode, err = runAttachedCommand(exe, koiArgs...)
	}

	if err != nil {
		log.Fatal(errors.Wrap(err, "Failed to run the command"))
		os.Exit(1)
	}
	os.Exit(exitCode)
}

func defaultEnv(env string, defaultVal string) string {
	if val, ok := os.LookupEnv(env); ok {
		return val
	}
	return defaultVal
}

func runAttachedCommand(command string, args ...string) (exitCode int, runErr error) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin

	yqIn, cmdOut := io.Pipe()
	yqCmd := exec.Command("yq", "-P")

	yqCmd.Stdin = yqIn
	yqCmd.Stdout = os.Stdout
	yqCmd.Stderr = os.Stderr

	cmd.Stdout = cmdOut
	cmd.Stderr = os.Stderr

	yqErr := yqCmd.Start()
	if yqErr != nil {
		return 1, errors.Wrapf(yqErr, "Failed to start yq command")
	}

	cmdErr := cmd.Start()
	if cmdErr != nil {
		return 1, errors.Wrapf(cmdErr, "Failed to start command %q %q", command, args)
	}

	log.Println("Waiting for cmd to finish...")
	ps, cmdErr := cmd.Process.Wait()
	cmdOut.Close()

	log.Println("Waiting for yq to finish...")
	_, yqErr = yqCmd.Process.Wait()
	if yqErr != nil {
		return 1, errors.Wrapf(yqErr, "Failed to run yq command")
	}

	exitCode = ps.ExitCode()
	return exitCode, errors.Wrapf(cmdErr, "Failed to run command %q %q", command, args)
}
