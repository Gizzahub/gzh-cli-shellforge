package cli

import (
	"testing"
)

func TestNewDeployCmd(t *testing.T) {
	cmd := newDeployCmd()

	if cmd.Use != "deploy" {
		t.Errorf("Use = %q, want %q", cmd.Use, "deploy")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	// Verify flags exist
	flags := []string{"build-dir", "dry-run", "backup", "verbose"}
	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Flag %q not found", flag)
		}
	}
}

func TestNewDeployCmd_DefaultBuildDir(t *testing.T) {
	cmd := newDeployCmd()

	flag := cmd.Flags().Lookup("build-dir")
	if flag == nil {
		t.Fatal("build-dir flag not found")
	}

	if flag.DefValue != "./build" {
		t.Errorf("build-dir default = %q, want %q", flag.DefValue, "./build")
	}
}

func TestNewDeployCmd_ShortFlags(t *testing.T) {
	cmd := newDeployCmd()

	// Check short flag for build-dir
	flag := cmd.Flags().ShorthandLookup("d")
	if flag == nil {
		t.Error("Short flag -d not found")
	}

	// Check short flag for verbose
	flag = cmd.Flags().ShorthandLookup("v")
	if flag == nil {
		t.Error("Short flag -v not found")
	}
}
