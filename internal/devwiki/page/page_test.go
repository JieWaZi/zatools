package page

import (
	"strings"
	"testing"
)

func TestParseSectionsParsesDevwikiSections(t *testing.T) {
	sections, err := ParseSections([]byte(`# Title

<!-- devwiki:section id=card -->
## 导航卡

- 主题定位：VIP
<!-- /devwiki:section -->

<!-- devwiki:section id=core -->
## 核心内容

正文
<!-- /devwiki:section -->
`))
	if err != nil {
		t.Fatalf("ParseSections() error = %v", err)
	}
	if len(sections) != 2 {
		t.Fatalf("len(sections) = %d, want 2", len(sections))
	}
	if sections[0].ID != "card" || sections[0].Title != "导航卡" {
		t.Fatalf("section[0] = %#v", sections[0])
	}
	if !strings.Contains(sections[1].Content, "正文") {
		t.Fatalf("core content = %q", sections[1].Content)
	}
}

func TestParseSectionsRejectsDuplicateNestedMissingEndAndEmpty(t *testing.T) {
	tests := map[string]string{
		"duplicate": `<!-- devwiki:section id=card -->
## A
<!-- /devwiki:section -->
<!-- devwiki:section id=card -->
## B
<!-- /devwiki:section -->
`,
		"nested": `<!-- devwiki:section id=card -->
## A
<!-- devwiki:section id=core -->
## B
<!-- /devwiki:section -->
`,
		"missing end": `<!-- devwiki:section id=card -->
## A
`,
		"empty": `<!-- devwiki:section id=card -->

<!-- /devwiki:section -->
`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseSections([]byte(input)); err == nil {
				t.Fatal("ParseSections() error = nil, want error")
			}
		})
	}
}

func TestParseNormalizesFrontmatterAndSections(t *testing.T) {
	doc, err := Parse("wiki/topics/vip.md", []byte(`---
title: "VIP"
slug: "[[vip]]"
kind: topic
summary: "VIP topic"
workflows:
  - "wiki/workflows/vip-runtime.md"
related_topics:
  - "[[ha]]"
---
# VIP

<!-- devwiki:section id=card -->
## 导航卡
card
<!-- /devwiki:section -->
`))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if doc.Meta.Slug != "vip" {
		t.Fatalf("Slug = %q, want vip", doc.Meta.Slug)
	}
	if len(doc.Meta.Workflows) != 1 || doc.Meta.Workflows[0] != "vip-runtime" {
		t.Fatalf("Workflows = %#v", doc.Meta.Workflows)
	}
	if len(doc.Meta.RelatedTopics) != 1 || doc.Meta.RelatedTopics[0] != "ha" {
		t.Fatalf("RelatedTopics = %#v", doc.Meta.RelatedTopics)
	}
	if _, ok := doc.SectionByID("card"); !ok {
		t.Fatal("missing card section")
	}
}
