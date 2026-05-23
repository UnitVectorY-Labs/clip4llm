package main

import (
	"fmt"
	"runtime"
	"testing"
)

// TestGreet is a basic unit test for the Greet function
func TestGreet(t *testing.T) {
	main()
}

func TestFormatVersionOutput(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "adds v prefix when missing",
			version: "1.2.3",
			want:    fmt.Sprintf("clip4llm version v1.2.3 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		},
		{
			name:    "preserves existing v prefix",
			version: "v2.3.4",
			want:    fmt.Sprintf("clip4llm version v2.3.4 (%s, %s/%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatVersionOutput("clip4llm", tt.version)
			if got != tt.want {
				t.Fatalf("formatVersionOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}
