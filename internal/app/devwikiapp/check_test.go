package devwikiapp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	common "zatools/internal/app/common"
	"zatools/internal/skills"
)

func TestCheckDocumentRejectsMissingRequiredSections(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "topics", "missing.md"), `---
title: Missing
slug: missing
kind: topic
summary: Missing sections
---
# Missing

<!-- devwiki:section id=card -->
## Card
card
<!-- /devwiki:section -->
`)
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Check(context.Background(), CheckOptions{Root: root, Types: []string{"document"}, Stdout: &out})
	if err == nil {
		t.Fatal("Check(document) error = nil, want missing section error")
	}
	if !strings.Contains(err.Error(), "document check failed") {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(out.String(), "missing required section \"core\"") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCheckDocumentAcceptsSpecificFile(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "topics", "ok.md"), validTopicDocument("ok"))
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "topics", "bad.md"), "# no frontmatter\n")
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Check(context.Background(), CheckOptions{
		Root:   root,
		Types:  []string{"document"},
		Paths:  []string{filepath.Join(root, "wiki", "topics", "ok.md")},
		Stdout: &out,
	})
	if err != nil {
		t.Fatalf("Check(document file) error = %v, output=%q", err, out.String())
	}
	if !strings.Contains(out.String(), "DevWiki document check passed") {
		t.Fatalf("output = %q", out.String())
	}
}

func TestCheckDocumentIgnoresSupportFilesWithoutSections(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "index.md"), "# Wiki Index\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "glossary.md"), "# Glossary\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "log.md"), "# Wiki Log\n\n- entry\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "topics", "ok.md"), validTopicDocument("ok"))

	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Check(context.Background(), CheckOptions{Root: root, Types: []string{"document"}, Stdout: &out})
	if err != nil {
		t.Fatalf("Check(document) error = %v, output=%q", err, out.String())
	}
	if strings.Contains(out.String(), "index.md") || strings.Contains(out.String(), "glossary.md") || strings.Contains(out.String(), "log.md") {
		t.Fatalf("support files should not be checked, output=%q", out.String())
	}
}

func TestCheckGraphDoesNotWriteOutputs(t *testing.T) {
	root := t.TempDir()
	mustWriteFileDevwikiApp(t, filepath.Join(root, "config", "project.yaml"), "project_name: Sample\nproject_slug: sample\n")
	mustWriteFileDevwikiApp(t, filepath.Join(root, "wiki", "topics", "vip.md"), validTopicDocument("vip"))
	service := NewServiceWithRuntime(common.Runtime{Workspace: skills.NewWorkspace(root)})
	var out bytes.Buffer
	err := service.Check(context.Background(), CheckOptions{Root: root, Types: []string{"graph"}, Stdout: &out})
	if err != nil {
		t.Fatalf("Check(graph) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".devwiki", "graph")); !os.IsNotExist(err) {
		t.Fatalf("graph check wrote output dir: %v", err)
	}
	if !strings.Contains(out.String(), "DevWiki graph check passed") {
		t.Fatalf("output = %q", out.String())
	}
}

func validTopicDocument(slug string) string {
	return `---
title: Topic
slug: ` + slug + `
kind: topic
summary: Topic summary
---
# Topic

<!-- devwiki:section id=card -->
## Card
card
<!-- /devwiki:section -->

<!-- devwiki:section id=core -->
## Core
core
<!-- /devwiki:section -->

<!-- devwiki:section id=explain -->
## Explain
explain
<!-- /devwiki:section -->
`
}
