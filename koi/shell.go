package koi

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ShellInvocation struct {
	name      string
	namespace string
	context   string
	image     string
	reason    string
	command   []string
	timeout   time.Duration
}

func ShellCommand(exe string, args []string) (exitCode int, runError error) {
	shell := ShellInvocation{}

	defaultPodPrefix := defaultEnv("USER", "koi")
	defaultPodName := fmt.Sprintf("%s-shell-%d", defaultPodPrefix, rand.Intn(1000))

	f := flag.NewFlagSet("shell", flag.ExitOnError)
	f.StringVarP(&shell.namespace, "namespace", "n", "", "The namespace to use")
	f.StringVarP(&shell.context, "context", "x", "", "The context to use")
	f.StringVarP(&shell.image, "image", "i", "oliverisaac/alpine-nettools:latest", "The image to use")
	f.StringVarP(&shell.reason, "reason", "r", "", "The reason for the shell")
	f.StringVar(&shell.name, "name", defaultEnv("KSHELL_NAME", defaultPodName), "The reason for the shell")
	debug := f.Bool("debug", false, "Enable debug logging")
	f.DurationVarP(&shell.timeout, "timeout", "t", 2*time.Minute, "Startup timeout duration (e.g. 2m)")

	f.Parse(args)

	if *debug {
		log.SetLevel(log.TraceLevel)
	}

	shell.command = f.Args()

	for shell.reason == "" {
		log.Error("You must provide a reason for the shell")
		fmt.Print("Enter a reason for this shell: ")
		_, err := fmt.Scanf("%s", &shell.reason)
		if err != nil {
			return 1, fmt.Errorf("Failed to read reason: %w", err)
		}
	}

	podJSON, err := getPodJSON(shell)
	if err != nil {
		return 1, fmt.Errorf("Failed to generate pod JSON: %w", err)
	}

	kubectlArgs := []string{"kubectl"}
	if shell.context != "" {
		kubectlArgs = append(kubectlArgs, "--context", shell.context)
	}

	if shell.namespace != "" {
		kubectlArgs = append(kubectlArgs, "--namespace", shell.namespace)
	}

	// Set the shell pod running
	log.Debug("Creating shell pod")
	err = runExternalCommand(strings.NewReader(podJSON), append(kubectlArgs, "apply", "-f", "-")...)
	if err != nil {
		return 1, fmt.Errorf("failed to create shell pod: %w", err)
	}

	// Clean up when we're done
	defer runExternalCommand(nil, append(kubectlArgs, "delete", "--wait=false", "pod", shell.name)...)

	log.Info("Waiting for shell pod to be ready...")
	err = runExternalCommand(strings.NewReader(podJSON), append(kubectlArgs, "wait", "--for=condition=ready", "pod", shell.name)...)
	if err != nil {
		return 1, fmt.Errorf("failed to wait for shell pod to be ready: %w", err)
	}

	if len(shell.command) == 0 {
		log.Debug("Running default shell command")
		command := "bash -l || sh -l"
		if os.Getenv("KSHELL_VI_MODE") == "true" {
			command = "bash -o vi -l || sh -l"
		}
		shell.command = []string{"sh", "-c", command}
	}

	execCommand := append(kubectlArgs, "exec", "-it", "-c", "shell", "pod/"+shell.name, "--")
	execCommand = append(execCommand, shell.command...)
	err = runExternalCommand(nil, execCommand...)
	if err != nil {
		return 1, fmt.Errorf("failed to exec into shell pod: %w", err)
	}

	return
}

func runExternalCommand(stdin io.Reader, command ...string) error {
	if len(command) == 0 {
		return fmt.Errorf("No command provided")
	}
	log.Debug("Running command: ", command)
	cmd := exec.Command(command[0], command[1:]...)
	if stdin == nil {
		stdin = os.Stdin
	}
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Failed to run command: %w", err)
	}
	return nil
}

func defaultEnv(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func getPodJSON(config ShellInvocation) (string, error) {
	// Generate the pod object
	retObj := k8sv1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.name,
			Annotations: map[string]string{
				"admission.stackrox.io/break-glass": config.reason,
			},
		},
		Spec: k8sv1.PodSpec{
			Containers: []k8sv1.Container{
				{
					Name:  "shell",
					Image: config.image,
					Command: []string{
						"/bin/sh",
						"-c",
						"sleep 1000000",
					},
				},
			},
		},
	}

	// convert to json and return
	ret, err := json.MarshalIndent(retObj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Failed to marshal pod object: %w", err)
	}
	return string(ret), nil
}
