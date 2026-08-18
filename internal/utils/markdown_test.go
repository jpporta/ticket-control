package utils

import (
	"strings"
	"testing"
)

func TestMarkdownToTypst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "headings",
			input:    "# Heading 1\n## Heading 2\n### Heading 3",
			expected: "= Heading 1\n== Heading 2\n=== Heading 3\n",
		},
		{
			name:     "bold and italic",
			input:    "This is **bold** and __also bold__ and _italic_.",
			expected: "This is *bold* and *also bold* and _italic_.\n",
		},
		{
			name:     "lists",
			input:    "* Item A\n* Item B\n1. Number 1\n2. Number 2",
			expected: "- Item A\n- Item B\n+ Number 1\n+ Number 2\n",
		},
		{
			name:     "horizontal rule",
			input:    "Before\n---\nAfter",
			expected: "Before\n#line(length: 100%, stroke: 0.5pt)\nAfter\n",
		},
		{
			name:     "code block passthrough",
			input:    "```go\n# Not a heading\nfunc main() {}\n```",
			expected: "```go\n# Not a heading\nfunc main() {}\n```\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MarkdownToTypst(tt.input)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.expected) {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}
