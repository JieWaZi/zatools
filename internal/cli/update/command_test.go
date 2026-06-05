package updatecmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"zatools/internal/app/updateapp"
)

type fakeUpdater struct {
	result updateapp.Result
	err    error
}

func (f fakeUpdater) Update(context.Context) (updateapp.Result, error) {
	return f.result, f.err
}

func TestCommandPrintsUpdatedBinaryPath(t *testing.T) {
	cmd := NewCommandWithService(fakeUpdater{
		result: updateapp.Result{
			Asset: "zatools_v9.9.9_linux_amd64.tar.gz",
			Path:  "/usr/local/bin/zatools",
		},
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	output := out.String()
	for _, want := range []string{"zatools_v9.9.9_linux_amd64.tar.gz", "/usr/local/bin/zatools"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestCommandPrintsDeferredWindowsReplacement(t *testing.T) {
	cmd := NewCommandWithService(fakeUpdater{
		result: updateapp.Result{
			Asset:    "zatools_v9.9.9_windows_amd64.tar.gz",
			Path:     `C:\Users\admin\AppData\Local\Programs\zatools\bin\zatools.exe`,
			Deferred: true,
		},
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute error = %v", err)
	}
	if !strings.Contains(out.String(), "restart") && !strings.Contains(out.String(), "重新打开") {
		t.Fatalf("deferred update output missing restart hint:\n%s", out.String())
	}
}
