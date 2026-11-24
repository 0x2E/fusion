package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantVersion bool
		wantErr     bool
		wantHelp    bool
	}{
		{
			name:        "short version flag",
			args:        []string{"-v"},
			wantVersion: true,
		},
		{
			name:        "long version flag",
			args:        []string{"--version"},
			wantVersion: true,
		},
		{
			name: "no arguments",
		},
		{
			name:    "unknown flag",
			args:    []string{"--bogus"},
			wantErr: true,
		},
		{
			name:     "help flag",
			args:     []string{"-h"},
			wantErr:  true,
			wantHelp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer

			gotVersion, err := parseArgs(tt.args, &output)

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("parseArgs(%q) unexpected error: %v", tt.args, err)
				}
			} else if err == nil {
				t.Fatalf("parseArgs(%q) error = nil, want an error", tt.args)
			} else if gotHelp := errors.Is(err, flag.ErrHelp); gotHelp != tt.wantHelp {
				t.Fatalf("parseArgs(%q) ErrHelp = %t, want %t (error: %v)", tt.args, gotHelp, tt.wantHelp, err)
			}

			if tt.wantErr && !strings.Contains(output.String(), "Usage of fusion:") {
				t.Errorf("parseArgs(%q) output = %q, want usage text", tt.args, output.String())
			}

			if gotVersion != tt.wantVersion {
				t.Errorf("parseArgs(%q) showVersion = %t, want %t", tt.args, gotVersion, tt.wantVersion)
			}
		})
	}
}
