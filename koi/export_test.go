package koi

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestExportCommand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantExitCode int
		wantOutput   string
		wantErr      bool
	}{
		{
			name:         "No input",
			input:        "",
			wantExitCode: 1,
			wantOutput:   "",
			wantErr:      true,
		},
		{
			name: "General pod",
			input: `
			apiVersion: v1
			kind: Pod
			metadata:
			  annotations:
				kubernetes.io/psp: eks.privileged
			  creationTimestamp: "2023-04-15T19:12:52Z"
			  generateName: vault-watcher-6584975cb5-
			  labels:
				app: vault-watcher
				pod-template-hash: 6584975cb5
			  name: vault-watcher-6584975cb5-m4gz5
			  namespace: vault
			  ownerReferences:
				- apiVersion: apps/v1
				  blockOwnerDeletion: true
				  controller: true
				  kind: ReplicaSet
				  name: vault-watcher-6584975cb5
				  uid: f64e07b6-83b1-4de8-893f-3b1bfd7b0e77
			  resourceVersion: "13861343"
			  uid: d7c55c36-c2e3-468d-ab5f-87c1b2e93cdc
			spec:
			  automountServiceAccountToken: true
			  containers:
				- args:
					- /config/config.yaml
				  image: ubunut:latest
				  imagePullPolicy: Always
				  name: vault-watcher
				  resources: {}
				  terminationMessagePath: /dev/termination-log
				  terminationMessagePolicy: File
			  dnsPolicy: ClusterFirst
			  enableServiceLinks: true
			  nodeName: example.node
			  preemptionPolicy: PreemptLowerPriority
			  priority: 0
			  restartPolicy: Always
			  schedulerName: default-scheduler
			  securityContext: {}
			  serviceAccount: vault-watcher
			  serviceAccountName: vault-watcher
			  terminationGracePeriodSeconds: 30
			status:
			  conditions:
				- lastProbeTime: null
				  lastTransitionTime: "2023-04-15T19:12:52Z"
				  status: "True"
				  type: Initialized
`,
			wantOutput: `
			apiVersion: v1
			kind: Pod
			metadata:
			  annotations:
				kubernetes.io/psp: eks.privileged
			  labels:
				app: vault-watcher
				pod-template-hash: 6584975cb5
			  name: vault-watcher-6584975cb5-m4gz5
			  namespace: vault
			spec:
			  automountServiceAccountToken: true
			  containers:
				- args:
					- /config/config.yaml
				  image: ubunut:latest
				  imagePullPolicy: Always
				  name: vault-watcher
				  resources: {}
				  terminationMessagePath: /dev/termination-log
				  terminationMessagePolicy: File
			  dnsPolicy: ClusterFirst
			  enableServiceLinks: true
			  preemptionPolicy: PreemptLowerPriority
			  priority: 0
			  restartPolicy: Always
			  schedulerName: default-scheduler
			  securityContext: {}
			  serviceAccount: vault-watcher
			  serviceAccountName: vault-watcher
			  terminationGracePeriodSeconds: 30
`,
		},
		{
			name: "Pod List",
			input: `
			apiVersion: v1
			items:
			- apiVersion: v1
			  kind: Pod
			  metadata:
				creationTimestamp: "2023-04-15T19:25:27Z"
				generateName: pod-
				name: pod-0
			- apiVersion: v1
			  kind: Pod
			  metadata:
				creationTimestamp: "2023-04-15T19:25:27Z"
				generateName: pod-
				name: pod-1
			`,
			wantOutput: `
			apiVersion: v1
			items:
			- apiVersion: v1
			  kind: Pod
			  metadata:
				name: pod-0
			- apiVersion: v1
			  kind: Pod
			  metadata:
				name: pod-1
			`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputStream := bytes.NewBufferString(strings.ReplaceAll(tt.input, "\t", "    "))
			output := &bytes.Buffer{}
			gotExitCode, err := ExportCommand(inputStream, output)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExitCode != tt.wantExitCode {
				t.Errorf("ExportCommand() = %v, want %v", gotExitCode, tt.wantExitCode)
			}

			var gotOutput interface{}
			var wantOutput interface{}
			if err := yaml.Unmarshal(output.Bytes(), &gotOutput); err != nil {
				t.Errorf("ExportCommand() error = %v", err)
			}
			wantOutputStr := strings.ReplaceAll(tt.wantOutput, "\t", "    ")
			if err := yaml.Unmarshal([]byte(wantOutputStr), &wantOutput); err != nil {
				t.Errorf("ExportCommand() error = %v", err)
			}
			if !reflect.DeepEqual(gotOutput, wantOutput) {
				t.Errorf("ExportCommand() = %+v, want %+v", gotOutput, wantOutput)
			}
		})
	}
}

func Test_deletePathIfExists(t *testing.T) {
	tests := []struct {
		name        string
		inputObject interface{}
		expect      interface{}
		path        []string
	}{
		{
			name: "Basic delete root",
			path: []string{"foo"},
			inputObject: map[string]interface{}{
				"foo": "bar",
			},
			expect: map[string]interface{}{},
		},
		{
			name: "Basic delete nested",
			path: []string{"a", "b", "c"},
			inputObject: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "d",
						"e": "f",
					},
				},
			},
			expect: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"e": "f",
					},
				},
			},
		},
		{
			name: "Handle array in path",
			path: []string{"items", "0", "name"},
			inputObject: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name": "foo",
						"age":  10,
					},
					map[string]interface{}{
						"name": "bar",
						"age":  20,
					},
				},
			},
			expect: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"age": 10,
					},
					map[string]interface{}{
						"name": "bar",
						"age":  20,
					},
				},
			},
		},
		{
			name: "Delete array",
			path: []string{"items", "[]", "name"},
			inputObject: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"name": "foo",
						"age":  10,
					},
					map[string]interface{}{
						"name": "bar",
						"age":  20,
					},
				},
			},
			expect: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"age": 10,
					},
					map[string]interface{}{
						"age": 20,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deletePathIfExists(tt.inputObject, tt.path...)
			if !reflect.DeepEqual(tt.inputObject, tt.expect) {
				t.Errorf("deletePathIfExists() = %v, want %v", tt.inputObject, tt.expect)
			}
		})
	}
}
