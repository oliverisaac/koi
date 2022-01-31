package koi

import (
	"reflect"
	"testing"
)

func Test_extractValueArgumentFromArgs(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		argumentFlags []string
		want          string
	}{
		{
			name:          "If asking for -n return next arg",
			args:          []string{"-n", "koi"},
			argumentFlags: []string{"-n"},
			want:          "koi",
		},
		{
			name:          "If asking for -n return next arg if it exists",
			args:          []string{"-n"},
			argumentFlags: []string{"-n"},
			want:          "",
		},
		{
			name:          "If asking for -n return next arg if it exists",
			args:          []string{"-n", "koi"},
			argumentFlags: []string{"--namespace", "-n"},
			want:          "koi",
		},
		{
			name:          "If asking for -n return next part after equals sign if it exists",
			args:          []string{"-n=koi"},
			argumentFlags: []string{"--namespace", "-n"},
			want:          "koi",
		},
		{
			name:          "If flag has an equal sign in it, that shouldn't matter",
			args:          []string{"-n=koi=bob"},
			argumentFlags: []string{"--namespace", "-n"},
			want:          "koi=bob",
		},
		{
			name:          "Flags after the double dash should be ignored",
			args:          []string{"exec", "--", "-n=koi=bob"},
			argumentFlags: []string{"--namespace", "-n"},
			want:          "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractValueArgumentFromArgs(tt.args, tt.argumentFlags...); got != tt.want {
				t.Errorf("extractArgumentFromArgs() got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func Test_appendArgument(t *testing.T) {
	tests := []struct {
		name string
		args []string
		flag string
		val  string
		want []string
	}{
		{
			name: "Inserted args should go befor the double dash",
			args: []string{"get", "pods", "--", "bob"},
			flag: "--namespace",
			val:  "koi",
			want: []string{"get", "pods", "--namespace", "koi", "--", "bob"},
		},
		{
			name: "Empty values inserted",
			args: []string{"get", "pods", "--", "bob"},
			flag: "--all-namespaces",
			val:  "",
			want: []string{"get", "pods", "--all-namespaces", "--", "bob"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendArgument(tt.args, tt.flag, tt.val); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("appendArgument(): %v, want: %v", got, tt.want)
			}
		})
	}
}

func Test_extractBoolArgumentFromArgs(t *testing.T) {
	tests := []struct {
		name          string
		argumentFlags []string
		args          []string
		want          bool
	}{
		{
			name:          "If a flag is set then it should be true",
			argumentFlags: []string{"--all-namespaces"},
			args:          []string{"--all-namespaces"},
			want:          true,
		},
		{
			name:          "If a flag is not set then it should be true",
			argumentFlags: []string{"--all-namespaces"},
			args:          []string{},
			want:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractBoolArgumentFromArgs(tt.args, tt.argumentFlags...); got != tt.want {
				t.Errorf("extractBoolArgumentFromArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
