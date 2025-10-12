package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNew verifies that New creates a Parser with goldmark configured
func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.md == nil {
		t.Fatal("Parser.md is nil")
	}
}

// TestParse tests the Parse method with valid markdown and frontmatter
func TestParse(t *testing.T) {
	p := New()
	content := []byte(`---
title: Test Post
date: 2024-01-15T10:00:00Z
description: A test post
tags: [test, example]
draft: false
---

# Hello World

This is **bold** and this is *italic*.

- Item 1
- Item 2
`)

	post, err := p.Parse(content, "2024-01-15-test-post.md")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	// Verify frontmatter
	if post.Title != "Test Post" {
		t.Errorf("Title = %q, want %q", post.Title, "Test Post")
	}

	expectedDate := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	if !post.Date.Equal(expectedDate) {
		t.Errorf("Date = %v, want %v", post.Date, expectedDate)
	}

	if post.Description != "A test post" {
		t.Errorf("Description = %q, want %q", post.Description, "A test post")
	}

	if len(post.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(post.Tags))
	}

	if post.Draft {
		t.Errorf("Draft = true, want false")
	}

	// Verify slug generation
	if post.Slug != "test-post" {
		t.Errorf("Slug = %q, want %q", post.Slug, "test-post")
	}

	// Verify keywords are generated from tags
	expectedKeywords := "test, example"
	if post.Keywords != expectedKeywords {
		t.Errorf("Keywords = %q, want %q", post.Keywords, expectedKeywords)
	}

	// Verify content conversion
	htmlContent := string(post.Content)
	if !strings.Contains(htmlContent, "<h1") {
		t.Errorf("Content doesn't contain h1 heading. Got: %s", htmlContent)
	}
	if !strings.Contains(htmlContent, "<strong>bold</strong>") {
		t.Errorf("Content doesn't contain bold text")
	}
	if !strings.Contains(htmlContent, "<em>italic</em>") {
		t.Errorf("Content doesn't contain italic text")
	}
	if !strings.Contains(htmlContent, "<ul>") {
		t.Errorf("Content doesn't contain list")
	}

	// Verify raw content is preserved
	if !strings.Contains(post.RawContent, "# Hello World") {
		t.Errorf("RawContent doesn't contain original markdown")
	}
}

// TestParse_DraftPost tests parsing a draft post
func TestParse_DraftPost(t *testing.T) {
	p := New()
	content := []byte(`---
title: Draft Post
date: 2024-01-15T10:00:00Z
description: A draft
tags: []
draft: true
---

This is a draft.
`)

	post, err := p.Parse(content, "draft.md")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if !post.Draft {
		t.Errorf("Draft = false, want true")
	}
}

// TestParse_InvalidFrontmatter tests parsing with invalid frontmatter
func TestParse_InvalidFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "no frontmatter delimiters",
			content: "This is just markdown content",
		},
		{
			name:    "single delimiter only",
			content: "---\ntitle: Test\n",
		},
		{
			name: "invalid YAML",
			content: `---
title: Test
invalid yaml here: [unclosed
---
Content`,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := p.Parse([]byte(tt.content), "test.md")
			if err == nil {
				t.Error("Parse() succeeded, want error")
			}
		})
	}
}

// TestParse_EmptyTags tests parsing with no tags
func TestParse_EmptyTags(t *testing.T) {
	p := New()
	content := []byte(`---
title: No Tags
date: 2024-01-15T10:00:00Z
description: Post with no tags
tags: []
draft: false
---

Content here.
`)

	post, err := p.Parse(content, "no-tags.md")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(post.Tags) != 0 {
		t.Errorf("len(Tags) = %d, want 0", len(post.Tags))
	}

	if post.Keywords != "" {
		t.Errorf("Keywords = %q, want empty string", post.Keywords)
	}
}

// TestParse_GoldmarkFeatures tests various goldmark features
func TestParse_GoldmarkFeatures(t *testing.T) {
	tests := []struct {
		name        string
		markdown    string
		wantHTML    string
		wantMissing string
	}{
		{
			name:     "GitHub Flavored Markdown - strikethrough",
			markdown: "This is ~~deleted~~ text",
			wantHTML: "<del>deleted</del>",
		},
		{
			name:     "GitHub Flavored Markdown - table",
			markdown: "| Col1 | Col2 |\n|------|------|\n| A    | B    |",
			wantHTML: "<table>",
		},
		{
			name:     "typographer - smart quotes",
			markdown: `"Hello"`,
			wantHTML: "&ldquo;Hello&rdquo;",
		},
		{
			name:     "auto heading IDs",
			markdown: "## My Heading",
			wantHTML: `id="my-heading"`,
		},
		{
			name:     "hard wraps",
			markdown: "Line 1\nLine 2",
			wantHTML: "<br",
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := []byte(`---
title: Test
date: 2024-01-15T10:00:00Z
draft: false
---

` + tt.markdown)

			post, err := p.Parse(content, "test.md")
			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			html := string(post.Content)
			if !strings.Contains(html, tt.wantHTML) {
				t.Errorf("Content doesn't contain %q\nGot: %s", tt.wantHTML, html)
			}

			if tt.wantMissing != "" && strings.Contains(html, tt.wantMissing) {
				t.Errorf("Content contains unwanted %q", tt.wantMissing)
			}
		})
	}
}

// TestParseFile tests parsing a real file
func TestParseFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "2024-01-15-test-post.md")

	content := `---
title: File Test
date: 2024-01-15T10:00:00Z
description: Testing file parsing
tags: [file, test]
draft: false
---

# File Content

This is from a file.
`

	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	p := New()
	post, err := p.ParseFile(filePath)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	if post.Title != "File Test" {
		t.Errorf("Title = %q, want %q", post.Title, "File Test")
	}

	if post.Slug != "test-post" {
		t.Errorf("Slug = %q, want %q", post.Slug, "test-post")
	}
}

// TestParseFile_NonExistent tests parsing a file that doesn't exist
func TestParseFile_NonExistent(t *testing.T) {
	p := New()
	_, err := p.ParseFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("ParseFile() succeeded, want error")
	}
}

// TestGenerateSlug tests the generateSlug function
func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{
			path: "content/posts/2024-01-15-my-first-post.md",
			want: "my-first-post",
		},
		{
			path: "2024-12-31-happy-new-year.md",
			want: "happy-new-year",
		},
		{
			path: "no-date-prefix.md",
			want: "no-date-prefix",
		},
		{
			path: "path/to/2024-01-15-nested-post.md",
			want: "nested-post",
		},
		{
			path: "simple.md",
			want: "simple",
		},
		{
			path: "no-extension",
			want: "no-extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := generateSlug(tt.path)
			if got != tt.want {
				t.Errorf("generateSlug(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// TestParse_MissingRequiredFields tests parsing with missing required fields
func TestParse_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "missing title",
			content: `---
date: 2024-01-15T10:00:00Z
---
Content`,
		},
		{
			name: "missing date",
			content: `---
title: Test
---
Content`,
		},
	}

	p := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := p.Parse([]byte(tt.content), "test.md")
			// Should parse but fields will be zero values
			if err != nil {
				t.Fatalf("Parse() failed: %v", err)
			}

			if tt.name == "missing title" && post.Title != "" {
				t.Errorf("Title = %q, want empty", post.Title)
			}

			if tt.name == "missing date" && !post.Date.IsZero() {
				t.Errorf("Date = %v, want zero", post.Date)
			}
		})
	}
}

// TestParse_ComplexMarkdown tests parsing complex markdown with multiple features
func TestParse_ComplexMarkdown(t *testing.T) {
	p := New()
	content := []byte(`---
title: Complex Post
date: 2024-01-15T10:00:00Z
description: A complex post with many markdown features
tags: [markdown, test, complex]
draft: false
---

# Main Heading

This is a paragraph with **bold**, *italic*, and ` + "`code`" + `.

## Subheading

Here's a list:
- Item 1
- Item 2
  - Nested item

1. Numbered item
2. Another numbered item

` + "```go" + `
func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

> This is a blockquote

[Link text](https://example.com)

![Alt text](image.jpg)
`)

	post, err := p.Parse(content, "complex.md")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	html := string(post.Content)

	// Verify various HTML elements are present
	expectedElements := []string{
		"<h1",
		"<h2",
		"<strong>",
		"<em>",
		"<code>",
		"<ul>",
		"<ol>",
		"<pre>",
		"<blockquote>",
		"<a href=",
		"<img",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(html, elem) {
			t.Errorf("Content missing expected element: %s\nGot: %s", elem, html)
		}
	}
}
