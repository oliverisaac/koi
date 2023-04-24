package koi

import (
	"os"
	"reflect"
	"testing"
)

func Test_ApplyTweaksToArgs(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		env               map[string]string
		want              []string
		wantFilterExe     string
		wantFilterCommand string
	}{
		{
			name: "-x flag should be changed to --context",
			args: []string{"-x"},
			want: []string{"--context"},
		},
		{
			name: "-x=bob flag should be changed to --context=bob",
			args: []string{"-x=bob"},
			want: []string{"--context=bob"},
		},
		{
			name: "-x should only be changed once",
			args: []string{"-x", "bob", "-x", "example"},
			want: []string{"--context", "bob", "-x", "example"},
		},
		{
			name: "-x should only be changed before a double-dash",
			args: []string{"-n", "bob", "--", "bash", "-x", "example"},
			want: []string{"-n", "bob", "--", "bash", "-x", "example"},
		},
		{
			name: "If KOI_NS is set, then the namespace should be set to that if it is not alreayd set",
			args: []string{},
			want: []string{"--namespace", "koi"},
			env: map[string]string{
				"KOI_NS": "koi",
			},
		},
		{
			name: "If KOI_CONTEXT is set, then the context should be set to that if it is not already set",
			args: []string{},
			want: []string{"--context", "koi"},
			env: map[string]string{
				"KOI_CONTEXT": "koi",
			},
		},
		{
			name: "If KOI_CONTEXT is set, then the context should be set to that if it is not already set",
			args: []string{"--context", "bob"},
			want: []string{"--context", "bob"},
			env: map[string]string{
				"KOI_CONTEXT": "koi",
			},
		},
		{
			name: "If KOI_CONTEXT is set, then the context should be set before double-dash",
			args: []string{"exec", "-it", "--", "--context", "ignroethis"},
			want: []string{"exec", "-it", "--context", "koi", "--", "--context", "ignroethis"},
			env: map[string]string{
				"KOI_CONTEXT": "koi",
			},
		},
		{
			name: "If both KOI_CONTEXt and KOI_NS are set, then they should both be used",
			args: []string{"exec", "-it", "--", "--context", "ignroethis"},
			want: []string{"exec", "-it", "--context", "koi", "--namespace", "koi", "--", "--context", "ignroethis"},
			env: map[string]string{
				"KOI_CTX": "koi",
				"KOI_NS":  "koi",
			},
		},
		{
			name:              "If yq is set, next arg is the filter exe",
			args:              []string{"exec", "--yq", ".containers"},
			want:              []string{"exec", "--output=json"},
			wantFilterExe:     "yq",
			wantFilterCommand: ".containers",
		},
		{
			name:              "If yq is set, next arg is the filter exe",
			args:              []string{"exec", "--jq"},
			want:              []string{"exec", "--output=json"},
			wantFilterExe:     "jq",
			wantFilterCommand: ".",
		},
		{
			name:              "If output is set to yq, then make a filter with that",
			args:              []string{"exec", "-o", "yq=.containers"},
			want:              []string{"exec", "--output=json"},
			wantFilterExe:     "yq",
			wantFilterCommand: ".containers",
		},
		{
			name:              "If output is set to yq then make a filter with .",
			args:              []string{"exec", "-o", "yq"},
			want:              []string{"exec", "--output=json"},
			wantFilterExe:     "yq",
			wantFilterCommand: ".",
		},
		{
			name:              "If output is set to yq then make a filter with .",
			args:              []string{"exec", "-o", "jq"},
			want:              []string{"exec", "--output=json"},
			wantFilterExe:     "jq",
			wantFilterCommand: ".",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				os.Setenv(key, val)
			}
			got, gotFilterExe, gotFilterCommand := ApplyTweaksToArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyTweaksToArgs() got: %v, want: %v", got, tt.want)
			}
			if gotFilterExe != tt.wantFilterExe {
				t.Errorf("applyTweaksToArgs() gotFilterExe: %v, want: %v", gotFilterExe, tt.wantFilterExe)
			}
			if gotFilterCommand != tt.wantFilterCommand {
				t.Errorf("applyTweaksToArgs() gotFilterCommand: %v, want: %v", gotFilterCommand, tt.wantFilterCommand)
			}
			for key := range tt.env {
				os.Unsetenv(key)
			}
		})
	}
}
