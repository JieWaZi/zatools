package main

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
)

func TestReportCLIErrorPrintsGenericErrors(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	code := reportCLIError(errors.New("boom"), &stderr)
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if stderr.String() != "boom\n" {
		t.Fatalf("stderr = %q, want %q", stderr.String(), "boom\n")
	}
}

func TestReportCLIErrorSuppressesExitStatusNoise(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("sh", "-c", "exit 7")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit error")
	}

	var stderr bytes.Buffer
	code := reportCLIError(err, &stderr)
	if code != 7 {
		t.Fatalf("code = %d, want 7", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
