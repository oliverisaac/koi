package koi

import "testing"

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "Basic koi get pods should return get",
			args: []string{"get", "pods"},
			want: "get",
		},
		{
			name: "Specifying the namespace should not cause poblems",
			args: []string{"-n", "koi", "containers"},
			want: "containers",
		},
		{
			name: "No command args should return no command",
			args: []string{"-n", "koi"},
			want: "",
		},
		{
			name: "No args should return no command",
			args: []string{},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCommand(tt.args); got != tt.want {
				t.Errorf("GetCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
