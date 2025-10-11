// Package builder orchestrates the site building process
package builder

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kvnloughead/ssg/internal/parser"
	"github.com/kvnloughead/ssg/internal/renderer"
	"gopkg.in/yaml.v3"
)

// SiteConfig represents the site configuration from config.yaml
type SiteConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	BaseURL     string `yaml:"baseUrl"`
	Author      string `yaml:"author"`
}

// Build generates the static site
func Build(configPath, outputDir string) error {
	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create parser
	p := parser.New()

	// Parse all posts
	posts, err := parseAllPosts(p, "content/posts")
	if err != nil {
		return fmt.Errorf("parsing posts: %w", err)
	}

	// Filter out drafts
	publishedPosts := filterDrafts(posts)

	// Sort posts by date (newest first)
	sort.Slice(publishedPosts, func(i, j int) bool {
		return publishedPosts[i].Date.After(publishedPosts[j].Date)
	})

	// Create renderer
	r, err := renderer.New("templates")
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
	}

	// Convert config
	rendererConfig := renderer.Config{
		Title:       config.Title,
		Description: config.Description,
		BaseURL:     config.BaseURL,
		Author:      config.Author,
	}

	// Clean and create output directory
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("cleaning output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Render index page
	indexPath := filepath.Join(outputDir, "index.html")
	if err := r.RenderIndex(publishedPosts, rendererConfig, indexPath); err != nil {
		return fmt.Errorf("rendering index: %w", err)
	}

	// Render individual post pages
	for _, post := range publishedPosts {
		postPath := filepath.Join(outputDir, "posts", post.Slug+".html")
		if err := r.RenderPost(post, rendererConfig, postPath); err != nil {
			return fmt.Errorf("rendering post %s: %w", post.Slug, err)
		}
	}

	// Copy static files
	if err := copyStatic("static", outputDir); err != nil {
		return fmt.Errorf("copying static files: %w", err)
	}

	fmt.Printf("Built %d posts to %s\n", len(publishedPosts), outputDir)
	return nil
}

// Serve starts a local development server
func Serve(port string) error {
	publicDir := "public"

	// Check if public directory exists
	if _, err := os.Stat(publicDir); os.IsNotExist(err) {
		return fmt.Errorf("public directory does not exist, run 'ssg build' first")
	}

	// Serve static files
	fs := http.FileServer(http.Dir(publicDir))
	http.Handle("/", fs)

	addr := ":" + port
	fmt.Printf("Serving site at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	// Initialize structured logger to stdout with default settings.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // include file and line number
	}))

	// Start HTTP server
	srv := &http.Server{
		Addr:              addr,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
		ReadHeaderTimeout: 60 * time.Second,
	}

	return srv.ListenAndServe()
}

// NewPost creates a new markdown post with frontmatter template
func NewPost(title string) error {
	// Create slug from title
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	var cleanSlug strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			cleanSlug.WriteRune(r)
		}
	}
	slug = cleanSlug.String()

	// Create filename with date
	date := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s-%s.md", date, slug)
	filepath := filepath.Join("content/posts", filename)

	// Create post template
	content := fmt.Sprintf(`---
title: %s
date: %s
description: ""
tags: []
draft: true
---

Write your post here...
`, title, time.Now().Format(time.RFC3339))

	// Write file
	if err := os.WriteFile(filepath, []byte(content), 0600); err != nil {
		return fmt.Errorf("writing post file: %w", err)
	}

	fmt.Printf("Created new post: %s\n", filepath)
	return nil
}

// loadConfig loads the site configuration from YAML
func loadConfig(path string) (*SiteConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config SiteConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// parseAllPosts parses all markdown files in a directory
func parseAllPosts(p *parser.Parser, dir string) ([]*parser.Post, error) {
	var posts []*parser.Post

	entries, err := os.ReadDir(dir)
	if err != nil {
		// If directory doesn't exist, return empty slice
		if os.IsNotExist(err) {
			return posts, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		post, err := p.ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		posts = append(posts, post)
	}

	return posts, nil
}

// filterDrafts removes draft posts from the list
func filterDrafts(posts []*parser.Post) []*parser.Post {
	var published []*parser.Post
	for _, post := range posts {
		if !post.Draft {
			published = append(published, post)
		}
	}
	return published
}

// copyStatic copies static files to output directory
func copyStatic(srcDir, dstDir string) error {
	// Check if static directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// No static files, that's OK
		return nil
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Destination path
		dstPath := filepath.Join(dstDir, relPath)

		if info.IsDir() {
			// Create directory
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}