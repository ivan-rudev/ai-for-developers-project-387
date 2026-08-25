package main

import (
	"testing"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/config"
)

func TestResolvePort(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr bool
	}{
		{name: "valid port", in: "9000", want: 9000},
		{name: "lower bound", in: "1", want: 1},
		{name: "upper bound", in: "65535", want: 65535},
		{name: "zero is invalid", in: "0", wantErr: true},
		{name: "above upper bound", in: "65536", wantErr: true},
		{name: "negative", in: "-1", wantErr: true},
		{name: "not a number", in: "abc", wantErr: true},
		{name: "empty", in: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePort(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolvePort(%q) = %d, want error", tt.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvePort(%q) unexpected error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("resolvePort(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestApplyPortFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		portEnv string
		want    int
		wantErr bool
	}{
		{name: "env overrides config", portEnv: "9000", want: 9000},
		{name: "unset uses config default", portEnv: "", want: 8080},
		{name: "invalid env errors", portEnv: "not-a-port", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			if cfg.Server.Port != 8080 {
				t.Fatalf("test precondition: DefaultConfig port = %d, want 8080", cfg.Server.Port)
			}
			env := map[string]string{"PORT": tt.portEnv}
			err := applyPortFromEnv(cfg, func(k string) string { return env[k] })
			if tt.wantErr {
				if err == nil {
					t.Fatalf("applyPortFromEnv() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("applyPortFromEnv() unexpected error: %v", err)
			}
			if cfg.Server.Port != tt.want {
				t.Fatalf("cfg.Server.Port = %d, want %d", cfg.Server.Port, tt.want)
			}
		})
	}
}
