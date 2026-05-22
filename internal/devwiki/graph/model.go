package graph

import "time"

const (
	// SchemaVersion is the graph JSON schema version.
	SchemaVersion = 1
	// BuilderVersion is bumped when graph build semantics change.
	BuilderVersion = 1
)

// PageType identifies the DevWiki page layer represented by a graph node.
type PageType string

const (
	PageTypeCapability PageType = "capability"
	PageTypeFeature    PageType = "feature"
	PageTypeWorkflow   PageType = "workflow"
)

// IssueLevel describes whether a graph issue blocks the build.
type IssueLevel string

const (
	IssueError   IssueLevel = "error"
	IssueWarning IssueLevel = "warning"
)

// Issue describes a validation or parsing finding.
type Issue struct {
	Level   IssueLevel `json:"level"`
	Path    string     `json:"path"`
	Message string     `json:"message"`
}

// Page is the normalized frontmatter extracted from one DevWiki Markdown page.
type Page struct {
	Type                PageType
	Path                string
	Slug                string
	Title               string
	Summary             string
	Status              string
	Confidence          string
	SearchTerms         []string
	Features            []string
	Capabilities        []string
	Workflows           []string
	RelatedCapabilities []string
	RelatedFeatures     []string
	RelatedWorkflows    []string
}

// Project describes the DevWiki project represented by the graph.
type Project struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Root string `json:"root"`
}

// Node is a rendered graph node.
type Node struct {
	ID          string   `json:"id"`
	Type        PageType `json:"type"`
	Slug        string   `json:"slug"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Status      string   `json:"status"`
	Confidence  string   `json:"confidence"`
	Path        string   `json:"path"`
	SearchTerms []string `json:"search_terms"`
}

// Edge is a rendered graph relation between two nodes.
type Edge struct {
	ID      string   `json:"id"`
	Type    string   `json:"type"`
	Source  string   `json:"source"`
	Target  string   `json:"target"`
	Label   string   `json:"label"`
	Sources []string `json:"sources"`
}

// Document is the right-panel document summary for one graph node.
type Document struct {
	Type    PageType `json:"type"`
	Path    string   `json:"path"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
}

// Graph is the JSON payload consumed by the static graph page.
type Graph struct {
	SchemaVersion int                 `json:"schema_version"`
	Project       Project             `json:"project"`
	BuiltAt       time.Time           `json:"built_at"`
	Nodes         []Node              `json:"nodes"`
	Edges         []Edge              `json:"edges"`
	Documents     map[string]Document `json:"documents"`
	Warnings      []Issue             `json:"warnings"`
}
