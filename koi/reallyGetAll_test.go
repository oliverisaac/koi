package koi

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func Test_parseNamespacedResources(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "Splits lines and drops blanks",
			input: "pods\nconfigmaps\n\ndeployments.apps\n",
			want:  []string{"pods", "configmaps", "deployments.apps"},
		},
		{
			name:  "Trims surrounding whitespace",
			input: "  pods  \n\tconfigmaps\n",
			want:  []string{"pods", "configmaps"},
		},
		{
			name:  "Empty output yields nothing",
			input: "",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseNamespacedResources(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNamespacedResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildGetAllArgs(t *testing.T) {
	tests := []struct {
		name        string
		resources   []string
		namespace   string
		kubeContext string
		output      string
		want        []string
	}{
		{
			name:      "Joins every resource into one comma separated argument",
			resources: []string{"pods", "configmaps", "deployments.apps"},
			namespace: "koi",
			want:      []string{"get", "pods,configmaps,deployments.apps", "--namespace", "koi"},
		},
		{
			name:      "Output format is passed through",
			resources: []string{"pods"},
			namespace: "koi",
			output:    "json",
			want:      []string{"get", "pods", "--namespace", "koi", "--output", "json"},
		},
		{
			name:        "Context and yaml output are passed through",
			resources:   []string{"pods"},
			kubeContext: "prod",
			output:      "yaml",
			want:        []string{"get", "pods", "--context", "prod", "--output", "yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildGetAllArgs(tt.resources, tt.namespace, tt.kubeContext, tt.output); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildGetAllArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

// fakeKubectl writes a stub executable that records the arguments of every call,
// answers api-resources with the given resource names and exits 0 on get.
func fakeKubectl(t *testing.T, resources string) (exe string, calls func() []string) {
	t.Helper()

	dir := t.TempDir()
	logPath := filepath.Join(dir, "calls.log")
	exe = filepath.Join(dir, "kubectl")

	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$*\" >> " + logPath + "\n" +
		"if [ \"$1\" = \"api-resources\" ]; then\n" +
		"  printf '" + resources + "'\n" +
		"  exit 0\n" +
		"fi\n" +
		"exit 0\n"

	if err := os.WriteFile(exe, []byte(script), 0o755); err != nil {
		t.Fatalf("writing fake kubectl: %v", err)
	}

	return exe, func() []string {
		raw, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("reading calls log: %v", err)
		}
		return parseNamespacedResources(string(raw))
	}
}

func Test_ReallyGetAllCommand(t *testing.T) {
	exe, calls := fakeKubectl(t, "pods\\nconfigmaps\\ndeployments.apps")

	exitCode, err := ReallyGetAllCommand(exe, []string{"-n", "koi", "-o", "json"})
	if err != nil {
		t.Fatalf("ReallyGetAllCommand() error = %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("ReallyGetAllCommand() exit = %d, want 0", exitCode)
	}

	got := calls()
	want := []string{
		"api-resources --namespaced=true --verbs=list -o name",
		"get pods,configmaps,deployments.apps --namespace koi --output json",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("kubectl calls = %v, want %v", got, want)
	}
}

func Test_ReallyGetAllCommandWithNoResources(t *testing.T) {
	exe, _ := fakeKubectl(t, "")

	_, err := ReallyGetAllCommand(exe, []string{})
	if err == nil || !strings.Contains(err.Error(), "no namespaced resources") {
		t.Fatalf("ReallyGetAllCommand() error = %v, want no namespaced resources", err)
	}
}
