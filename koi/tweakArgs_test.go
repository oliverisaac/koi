package koi

import (
	"reflect"
	"testing"
)

func Test_ApplyTweaksToArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplyTweaksToArgs(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("applyTweaksToArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
