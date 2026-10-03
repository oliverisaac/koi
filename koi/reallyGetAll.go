package koi

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

// ReallyGetAllCommand lists every namespaced API resource in a single call.
// It asks the cluster which resources exist and support listing, then hands the
// whole comma separated list to one "kubectl get". The output flags (-o/--output)
// are passed straight through, so json and yaml work, and the default is
// kubectl's table output.
func ReallyGetAllCommand(exe string, args []string) (exitCode int, runError error) {
	flags := pflag.NewFlagSet("really-get-all", pflag.ExitOnError)

	var namespace string
	flags.StringVarP(&namespace, "namespace", "n", "", "Namespace to get resources in")

	var kubeContext string
	flags.StringVarP(&kubeContext, "context", "x", "", "Context to get resources in")

	var output string
	flags.StringVarP(&output, "output", "o", "", "Output format to pass to kubectl, for example json or yaml")

	if err := flags.Parse(args); err != nil {
		return 1, errors.Wrap(err, "parsing flags")
	}

	if extra := flags.Args(); len(extra) > 0 {
		return 1, fmt.Errorf("unexpected arguments: %v", extra)
	}

	rawResources, err := getNamespacedResources(exe, kubeContext)
	if err != nil {
		return 1, errors.Wrap(err, "identifying namespaced resources")
	}

	resources := parseNamespacedResources(rawResources)
	if len(resources) == 0 {
		return 1, fmt.Errorf("the cluster reported no namespaced resources")
	}

	return runCommandAndFilterOutput(exe, buildGetAllArgs(resources, namespace, kubeContext, output))
}

// getNamespacedResources asks kubectl for every namespaced resource that
// supports the list verb, in "resource.group" form, one per line.
func getNamespacedResources(exe string, kubeContext string) (string, error) {
	args := []string{"api-resources", "--namespaced=true", "--verbs=list", "-o", "name"}
	if kubeContext != "" {
		args = append(args, "--context", kubeContext)
	}

	cmd := exec.Command(exe, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "running %s %s: %s", exe, strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}

	return string(out), nil
}

func parseNamespacedResources(output string) []string {
	var resources []string
	for _, line := range strings.Split(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			resources = append(resources, line)
		}
	}
	return resources
}

func buildGetAllArgs(resources []string, namespace string, kubeContext string, output string) []string {
	args := []string{"get", strings.Join(resources, ",")}

	if kubeContext != "" {
		args = append(args, "--context", kubeContext)
	}
	if namespace != "" {
		args = append(args, "--namespace", namespace)
	}
	if output != "" {
		args = append(args, "--output", output)
	}

	return args
}
