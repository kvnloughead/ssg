package ssg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kvnloughead/ssg/internal/parser"
)

// TestBuild tests the full Build function
func TestBuild(t *testing.T) {
	// Create temporary directories
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "posts")
	templatesDir := filepath.Join(tmpDir, "templates")
	staticDir := filepath.Join(tmpDir, "static", "css")
	outputDir := filepath.Join(tmpDir, "public")

	// Create directory structure
	if err := os.MkdirAll(contentDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(templatesDir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(staticDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Create config file
	configPath := filepath.Join(tmpDir, "config.yaml")
	configContent := `title: Test Blog
description: A test blog
baseUrl: https://test.com
author: Test Author
keywords: test, blog
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create a test post
	postContent := `---
title: Test Post
date: 2024-01-15T10:00:00Z
description: A test post
tags: [test]
draft: false
---

# Hello World

This is a test post.
`
	postPath := filepath.Join(contentDir, "2024-01-15-test-post.md")
	if err := os.WriteFile(postPath, []byte(postContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create draft post (should be excluded)
	draftContent := `---
title: Draft Post
date: 2024-01-16T10:00:00Z
description: A draft
tags: []
draft: true
---

Draft content.
`
	draftPath := filepath.Join(contentDir, "2024-01-16-draft.md")
	if err := os.WriteFile(draftPath, []byte(draftContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Create templates
	baseTemplate := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{template "content" .}}</body>
</html>`
	if err := os.WriteFile(filepath.Join(templatesDir, "base.html"), []byte(baseTemplate), 0600); err != nil {
		t.Fatal(err)
	}

	indexTemplate := `{{define "content"}}
<div>{{range .Posts}}<article>{{.Title}}</article>{{end}}</div>
{{end}}`
	if err := os.WriteFile(filepath.Join(templatesDir, "index.html"), []byte(indexTemplate), 0600); err != nil {
		t.Fatal(err)
	}

	postTemplate := `{{define "content"}}
<article>{{.Post.Title}}</article>
{{end}}`
	if err := os.WriteFile(filepath.Join(templatesDir, "post.html"), []byte(postTemplate), 0600); err != nil {
		t.Fatal(err)
	}

	// Create static file
	cssContent := "body { color: black; }"
	if err := os.WriteFile(filepath.Join(staticDir, "style.css"), []byte(cssContent), 0600); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Run build
	err = Build(configPath, outputDir)
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify index.html was created
	indexPath := filepath.Join(outputDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("index.html was not created")
	}

	// Verify post was created
	generatedPostPath := filepath.Join(outputDir, "posts", "test-post.html")
	if _, err := os.Stat(generatedPostPath); os.IsNotExist(err) {
		t.Error("test-post.html was not created")
	}

	// Verify draft was NOT created
	draftOutputPath := filepath.Join(outputDir, "posts", "draft.html")
	if _, err := os.Stat(draftOutputPath); !os.IsNotExist(err) {
		t.Error("draft.html should not have been created")
	}

	// Verify static files were copied
	cssOutputPath := filepath.Join(outputDir, "css", "style.css")
	if _, err := os.Stat(cssOutputPath); os.IsNotExist(err) {
		t.Error("CSS file was not copied")
	}

	// Verify index content
	indexHTML, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(indexHTML), "Test Post") {
		t.Error("Index page doesn't contain post title")
	}
	if strings.Contains(string(indexHTML), "Draft Post") {
		t.Error("Index page contains draft post (should be excluded)")
	}
}

// TestNewPost tests creating a new post
func TestNewPost(t *testing.T) {
	tmpDir := t.TempDir()
	contentDir := filepath.Join(tmpDir, "content", "posts")
	if err := os.MkdirAll(contentDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Change to temp directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Create new post
	title := "My Test Post"
	err = NewPost(title)
	if err != nil {
		t.Fatalf("NewPost() failed: %v", err)
	}

	// Verify file was created
	entries, err := os.ReadDir(contentDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(entries))
	}

	// Verify filename format (YYYY-MM-DD-my-test-post.md)
	filename := entries[0].Name()
	if !strings.HasSuffix(filename, "-my-test-post.md") {
		t.Errorf("Filename = %q, want suffix '-my-test-post.md'", filename)
	}

	// Verify frontmatter
	content, err := os.ReadFile(filepath.Join(contentDir, filename))
	if err != nil {
		t.Fatal(err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "title: "+title) {
		t.Error("Content doesn't contain title")
	}
	if !strings.Contains(contentStr, "draft: true") {
		t.Error("Content doesn't have draft: true")
	}
	if !strings.Contains(contentStr, "tags: []") {
		t.Error("Content doesn't have tags")
	}
}

// TestNewPost_SlugGeneration tests slug generation for various titles
func TestNewPost_SlugGeneration(t *testing.T) {
	tests := []struct {
		title    string
		wantSlug string
	}{
		{"Simple Title", "simple-title"},
		{"Title With Numbers 123", "title-with-numbers-123"},
		{"Title!!!With###Special@@@Characters", "titlewithspecialcharacters"},
		{"Multiple   Spaces", "multiple---spaces"}, // Multiple spaces create multiple hyphens
		{"ALL CAPS TITLE", "all-caps-title"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			tmpDir := t.TempDir()
			contentDir := filepath.Join(tmpDir, "content", "posts")
			if err := os.MkdirAll(contentDir, 0750); err != nil {
				t.Fatal(err)
			}

			origDir, _ := os.Getwd()
			defer os.Chdir(origDir)
			os.Chdir(tmpDir)

			err := NewPost(tt.title)
			if err != nil {
				t.Fatalf("NewPost() failed: %v", err)
			}

			entries, err := os.ReadDir(contentDir)
			if err != nil {
				t.Fatal(err)
			}

			filename := entries[0].Name()
			if !strings.Contains(filename, tt.wantSlug) {
				t.Errorf("Filename %q doesn't contain slug %q", filename, tt.wantSlug)
			}
		})
	}
}

// TestFilterDrafts tests draft filtering
func TestFilterDrafts(t *testing.T) {
	posts := []*parser.Post{
		{Title: "Published 1", Draft: false},
		{Title: "Draft 1", Draft: true},
		{Title: "Published 2", Draft: false},
		{Title: "Draft 2", Draft: true},
		{Title: "Published 3", Draft: false},
	}

	published := filterDrafts(posts)

	if len(published) != 3 {
		t.Errorf("len(published) = %d, want 3", len(published))
	}

	for _, post := range published {
		if post.Draft {
			t.Errorf("Published posts contain draft: %s", post.Title)
		}
	}
}

// TestParseAllPosts tests parsing multiple posts
func TestParseAllPosts(t *testing.T) {
	tmpDir := t.TempDir()
	postsDir := filepath.Join(tmpDir, "posts")
	if err := os.MkdirAll(postsDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Create test posts
	posts := []struct {
		filename string
		content  string
	}{
		{
			"2024-01-15-first.md",
			`---
title: First Post
date: 2024-01-15T10:00:00Z
draft: false
---
Content 1`,
		},
		{
			"2024-01-16-second.md",
			`---
title: Second Post
date: 2024-01-16T10:00:00Z
draft: false
---
Content 2`,
		},
		{
			"2024-01-17-third.md",
			`---
title: Third Post
date: 2024-01-17T10:00:00Z
draft: true
---
Content 3`,
		},
	}

	for _, post := range posts {
		path := filepath.Join(postsDir, post.filename)
		if err := os.WriteFile(path, []byte(post.content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-markdown file (should be ignored)
	if err := os.WriteFile(filepath.Join(postsDir, "readme.txt"), []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	p := parser.New()
	parsed, err := parseAllPosts(p, postsDir)
	if err != nil {
		t.Fatalf("parseAllPosts() failed: %v", err)
	}

	if len(parsed) != 3 {
		t.Errorf("len(parsed) = %d, want 3", len(parsed))
	}
}

// TestParseAllPosts_EmptyDirectory tests parsing an empty directory
func TestParseAllPosts_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	postsDir := filepath.Join(tmpDir, "posts")
	if err := os.MkdirAll(postsDir, 0750); err != nil {
		t.Fatal(err)
	}

	p := parser.New()
	parsed, err := parseAllPosts(p, postsDir)
	if err != nil {
		t.Fatalf("parseAllPosts() failed: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("len(parsed) = %d, want 0", len(parsed))
	}
}

// TestParseAllPosts_NonExistentDirectory tests parsing a non-existent directory
func TestParseAllPosts_NonExistentDirectory(t *testing.T) {
	p := parser.New()
	parsed, err := parseAllPosts(p, "/nonexistent/path")
	if err != nil {
		t.Fatalf("parseAllPosts() should not error on non-existent dir: %v", err)
	}

	if len(parsed) != 0 {
		t.Errorf("len(parsed) = %d, want 0", len(parsed))
	}
}

// TestLoadConfig tests loading site configuration
func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `title: My Blog
description: A test blog
baseUrl: https://example.com
author: John Doe
keywords: golang, blog
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatal(err)
	}

	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	if config.Title != "My Blog" {
		t.Errorf("Title = %q, want %q", config.Title, "My Blog")
	}
	if config.Description != "A test blog" {
		t.Errorf("Description = %q, want %q", config.Description, "A test blog")
	}
	if config.BaseURL != "https://example.com" {
		t.Errorf("BaseURL = %q, want %q", config.BaseURL, "https://example.com")
	}
	if config.Author != "John Doe" {
		t.Errorf("Author = %q, want %q", config.Author, "John Doe")
	}
	if config.Keywords != "golang, blog" {
		t.Errorf("Keywords = %q, want %q", config.Keywords, "golang, blog")
	}
}

// TestLoadConfig_NonExistent tests loading a non-existent config file
func TestLoadConfig_NonExistent(t *testing.T) {
	_, err := loadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Error("loadConfig() succeeded, want error")
	}
}

// TestLoadConfig_InvalidYAML tests loading invalid YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `title: Test
description: [unclosed bracket
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfig(configPath)
	if err == nil {
		t.Error("loadConfig() succeeded with invalid YAML, want error")
	}
}

// TestCopyStatic tests copying static files
func TestCopyStatic(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "static")
	dstDir := filepath.Join(tmpDir, "public")

	// Create source directory structure
	if err := os.MkdirAll(filepath.Join(srcDir, "css"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(srcDir, "images"), 0750); err != nil {
		t.Fatal(err)
	}

	// Create files
	files := map[string]string{
		"css/style.css":   "body { color: black; }",
		"images/logo.png": "fake png data",
		"robots.txt":      "User-agent: *",
	}

	for path, content := range files {
		fullPath := filepath.Join(srcDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}

	// Copy static files
	err := copyStatic(srcDir, dstDir)
	if err != nil {
		t.Fatalf("copyStatic() failed: %v", err)
	}

	// Verify files were copied
	for path := range files {
		dstPath := filepath.Join(dstDir, path)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("File %s was not copied", path)
		}
	}

	// Verify content
	cssPath := filepath.Join(dstDir, "css", "style.css")
	content, err := os.ReadFile(cssPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != files["css/style.css"] {
		t.Error("Copied file content doesn't match")
	}
}

// TestCopyStatic_NonExistentSource tests copying from non-existent directory
func TestCopyStatic_NonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	err := copyStatic("/nonexistent", tmpDir)
	if err != nil {
		t.Errorf("copyStatic() with non-existent source should not error, got: %v", err)
	}
}

// TestRenderer_Integration tests renderer with actual templates
func TestRenderer_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	if err := os.MkdirAll(templatesDir, 0750); err != nil {
		t.Fatal(err)
	}

	// Create templates
	baseTemplate := `<!DOCTYPE html>
<html>
<head><title>{{.Title}}</title></head>
<body>{{template "content" .}}</body>
</html>`
	if err := os.WriteFile(filepath.Join(templatesDir, "base.html"), []byte(baseTemplate), 0600); err != nil {
		t.Fatal(err)
	}

	postTemplate := `{{define "content"}}
<article><h1>{{.Post.Title}}</h1><div>{{.Post.Content}}</div></article>
{{end}}`
	if err := os.WriteFile(filepath.Join(templatesDir, "post.html"), []byte(postTemplate), 0600); err != nil {
		t.Fatal(err)
	}

	// Create renderer
	r, err := newRenderer(templatesDir)
	if err != nil {
		t.Fatalf("newRenderer() failed: %v", err)
	}

	// Create test post
	testPost := &parser.Post{
		Title:   "Test Post",
		Date:    time.Now(),
		Slug:    "test-post",
		Content: "<p>Test content</p>",
	}

	config := SiteConfig{
		Title:  "Test Site",
		Author: "Test Author",
	}

	outputPath := filepath.Join(outputDir, "test.html")

	// Change to temp directory so renderToFile can find templates
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	// Render post
	err = r.renderPost(testPost, config, outputPath)
	if err != nil {
		t.Fatalf("renderPost() failed: %v", err)
	}

	// Verify output
	html, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	htmlStr := string(html)
	if !strings.Contains(htmlStr, "Test Post") {
		t.Error("Rendered HTML doesn't contain post title")
	}
	if !strings.Contains(htmlStr, "Test content") {
		t.Error("Rendered HTML doesn't contain post content")
	}
}
