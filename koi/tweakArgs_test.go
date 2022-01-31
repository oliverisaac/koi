package koi

import (
	"os"
	"reflect"
	"testing"
)

func Test_ApplyTweaksToArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
		env  map[string]string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				os.Setenv(key, val)
			}
			if got := ApplyTweaksToArgs(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyTweaksToArgs() got: %v, want: %v", got, tt.want)
			}
			for key := range tt.env {
				os.Unsetenv(key)
			}
		})
	}
}
