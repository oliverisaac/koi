package main

import (
	"log"
	"os"
	"os/exec"

	"github.com/oliverisaac/koi/koi"
	"github.com/pkg/errors"
)

func main() {
	var exitCode int
	var err error

	exe := defaultEnv("KOI_KUBECTL_EXE", "kubectl")
	koiArgs := koi.ApplyTweaksToArgs(os.Args[1:])

	exitCode, err = runAttachedCommand(exe, koiArgs...)
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
