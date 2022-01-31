package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/oliverisaac/koi/koi"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr = cmd.Run()
	exitCode = cmd.ProcessState.ExitCode()
	return exitCode, errors.Wrapf(runErr, "Failed to run command %q %q", command, args)
}
