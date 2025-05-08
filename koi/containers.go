package koi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/rodaine/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func ContainersCommand(args []string) (exitCode int, runError error) {
	flags := pflag.NewFlagSet("kcontainers", pflag.ExitOnError)

	var kubeContext string
	flags.StringVarP(&kubeContext, "context", "c", "", "Context to get contianers in")

	var namespace string
	flags.StringVarP(&namespace, "namespace", "n", "", "Namespace to get contianers in")

	var allNamespaces bool
	flags.BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Get containers in all namespaces")

	var writeInColor bool
	flags.BoolVar(&writeInColor, "color", WritingToTerminal(), "Configure color output")

	err := flags.Parse(args)
	if err != nil {
		return 1, errors.Wrap(err, "parsing flags")
	}

	logrus := logrus.WithFields(logrus.Fields{
		"context":        kubeContext,
		"namespace":      namespace,
		"all-namespaces": allNamespaces,
	})

	logrus.Debug("Going to run kcontainers")

	if allNamespaces {
		namespace = ""
	}

	pods, err := getPodsByNamespace(kubeContext, namespace)
	if err != nil {
		return -1, errors.Wrap(err, "getting pods")
	}

	tbl := table.New("Namespace", "Pod", "Container", "Init", "Status")

	for _, pod := range pods {
		ns := pod.GetNamespace()
		podName := pod.GetName()
		statusByContainer := map[string]string{}

		for _, status := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
			statusName := "unknown"
			statusColor := color.WhiteString
			if status.Ready {
				statusName = "ready"
				statusColor = color.GreenString
			} else if status.State.Running != nil {
				statusName = "running"
				statusColor = color.CyanString
			} else if status.State.Terminated != nil {
				statusName = "terminated"
				statusColor = color.MagentaString
			} else if status.State.Waiting != nil {
				statusName = "waiting"
				statusColor = color.YellowString
			}

			if !writeInColor {
				statusColor = fmt.Sprintf
			}
			statusByContainer[status.Name] = statusColor("%s", statusName)
		}

		for _, container := range pod.Spec.InitContainers {
			status, ok := statusByContainer[container.Name]
			if !ok {
				status = "unknown"
			}
			tbl.AddRow(ns, podName, container.Name, true, status)
		}

		for _, container := range pod.Spec.Containers {
			status, ok := statusByContainer[container.Name]
			if !ok {
				status = "unknown"
			}
			tbl.AddRow(ns, podName, container.Name, false, status)
		}
	}

	if writeInColor {
		tbl.WithHeaderFormatter(color.New(color.FgHiWhite, color.Underline).SprintfFunc())
	}
	tbl.Print()

	return 0, nil
}

func WritingToTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		panic(err)
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func getKubeClient(kubeContext string) (*kubernetes.Clientset, error) {
	kubeconfigPath, ok := os.LookupEnv("KUBE_CONFIG")
	if !ok {
		kubeconfigPath = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}
	if _, err := os.Stat(kubeconfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig not found at %s", kubeconfigPath)
	}

	// Load the kubeconfig with the specified context
	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: kubeContext,
	}
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("building rest config: %w", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}
	return clientset, nil
}

func getPodsByNamespace(kubeContext string, namespace string) ([]v1.Pod, error) {
	client, err := getKubeClient(kubeContext)

	podList, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	return podList.Items, nil
}
