package scripts

import (
	"os"
	"strings"
	"testing"
)

func TestPowerShellInstallerDoesNotSplitSessionPathWithoutNullGuard(t *testing.T) {
	content, err := os.ReadFile("install.ps1")
	if err != nil {
		t.Fatalf("read install.ps1: %v", err)
	}

	script := string(content)
	if strings.Contains(script, "$env:Path.Split(") {
		t.Fatalf("install.ps1 must not call $env:Path.Split directly; empty Path triggers a null-value method error")
	}
}

func TestPowerShellInstallerDoesNotUseRuntimeInformationForArchitecture(t *testing.T) {
	content, err := os.ReadFile("install.ps1")
	if err != nil {
		t.Fatalf("read install.ps1: %v", err)
	}

	script := string(content)
	if strings.Contains(script, "RuntimeInformation]::OSArchitecture") {
		t.Fatalf("install.ps1 must not use RuntimeInformation.OSArchitecture; Windows PowerShell can return null there")
	}
}
