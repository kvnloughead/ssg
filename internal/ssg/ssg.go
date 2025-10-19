// Package ssg provides static site generation functionality, including building,
// rendering, and serving the site.
package ssg

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kvnloughead/ssg/internal/parser"
	"gopkg.in/yaml.v3"
)

// SiteConfig represents the site configuration from config.yaml
type SiteConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	BaseURL     string `yaml:"baseUrl"`
	Author      string `yaml:"author"`
	Keywords    string `yaml:"keywords"`
}

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
}

// PageData holds data passed to templates
type PageData struct {
	Site  SiteConfig
	Post  *parser.Post
	Posts []*parser.Post
	Title string
}

// Build generates the static site by orchestrating parser and renderer.
//
// Flow:
//  1. Loads site configuration from config.yaml (title, author, etc.)
//  2. Creates a parser instance to handle markdown conversion
//  3. Parses all markdown files in content/posts/ using parser.ParseFile
//  4. Filters out draft posts and sorts by date (newest first)
//  5. Creates a renderer instance with templates from templates/
//  6. Renders posts.html with the list of posts using renderer.renderIndex
//  7. Renders individual post pages using renderer.renderPost
//  8. Copies static assets (CSS, images, etc.) to output directory
//
// Parameters:
//   - configPath: Path to config.yaml containing site metadata
//   - outputDir: Directory where generated HTML files will be written (usually "public")
//
// Returns an error if any step fails (config loading, parsing, rendering, or file I/O).
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
	r, err := newRenderer("templates")
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
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
	if err := r.renderIndex(publishedPosts, *config, indexPath); err != nil {
		return fmt.Errorf("rendering index: %w", err)
	}

	// Render individual post pages
	for _, post := range publishedPosts {
		postPath := filepath.Join(outputDir, "posts", post.Slug+".html")
		if err := r.renderPost(post, *config, postPath); err != nil {
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

// Serve starts a local development server to preview the generated site.
//
// Serves static files from the "public" directory on the specified port.
// This is a simple HTTP file server for local development only.
//
// Parameters:
//   - port: Port number to serve on (e.g., "3000" for localhost:3000)
//
// Returns an error if the public directory doesn't exist or server fails to start.
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

// NewPost creates a new markdown post file with YAML frontmatter template.
//
// Creates a new file in content/posts/ with the format: YYYY-MM-DD-slug.md
// The slug is generated from the title (lowercase, spaces to hyphens, alphanumeric only).
// The file is pre-populated with YAML frontmatter including title, date, and draft status.
//
// Parameters:
//   - title: Human-readable title for the post (e.g., "My First Post")
//
// Returns an error if file creation fails.
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

// newRenderer creates a new Renderer with all templates pre-loaded from the template directory.
//
// Uses template.ParseGlob to load all *.html files in the directory into a single
// template set. Each file is named by its filename (e.g., "base.html", "posts.html").
// Templates can reference each other using {{define}} blocks.
//
// Expected template structure:
//   - base.html: Main layout with {{template "posts" .}} placeholder
//   - posts.html: Defines {{define "posts"}} for the posts list page
//   - post.html: Defines {{define "posts"}} for individual post pages
//
// Parameters:
//   - templateDir: Directory containing HTML templates (e.g., "templates")
//
// Returns a Renderer instance or an error if template loading fails.
func newRenderer(templateDir string) (*Renderer, error) {
	// Load all templates
	tmpl, err := template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return &Renderer{templates: tmpl}, nil
}

// renderPost renders a single blog post page to an HTML file.
//
// Called by Build for each published post. Creates a PageData struct with
// the post content and site config, then calls renderToFile with "post.html" to
// render base.html + post.html's {{define "posts"}} block.
//
// Parameters:
//   - post: Parsed post struct from parser.ParseFile containing title, content, etc.
//   - config: Site configuration (title, author, etc.) for template rendering
//   - outputPath: Where to write the HTML file (e.g., "public/posts/my-post.html")
//
// Returns an error if rendering or file writing fails.
func (r *Renderer) renderPost(post *parser.Post, config SiteConfig, outputPath string) error {
	data := PageData{
		Site:  config,
		Post:  post,
		Title: post.Title,
	}

	return r.renderToFile("post.html", data, outputPath)
}

// renderIndex renders the home page with a list of all published posts.
//
// Called by Build to create the main posts.html page. Creates a
// PageData struct with all posts and site config, then calls renderToFile with
// "posts.html" to render base.html + posts.html's {{define "posts"}} block.
//
// Parameters:
//   - posts: Slice of all published posts (already filtered and sorted by builder)
//   - config: Site configuration (title, author, etc.) for template rendering
//   - outputPath: Where to write the HTML file (e.g., "public/posts.html")
//
// Returns an error if rendering or file writing fails.
func (r *Renderer) renderIndex(posts []*parser.Post, config SiteConfig, outputPath string) error {
	data := PageData{
		Site:  config,
		Posts: posts,
		Title: config.Title,
	}

	return r.renderToFile("posts.html", data, outputPath)
}

// renderToFile renders a page by combining base.html with a content template.
//
// This is where the template inheritance pattern is implemented:
//  1. Clones the pre-loaded base.html template (for a fresh copy)
//  2. Parses the content template file (posts.html or post.html) which contains
//     a {{define "posts"}} block
//  3. Executes base.html, which calls {{template "posts" .}} to inject the
//     appropriate content block
//  4. Writes the final HTML to the output file
//
// This allows index and post pages to share the same header/footer/nav from base.html
// while having different main content.
//
// Parameters:
//   - contentTemplate: Which content template to use ("posts.html" or "post.html")
//   - data: PageData struct containing site config and post(s) for template variables
//   - outputPath: Where to write the rendered HTML file
//
// Returns an error if template cloning, parsing, execution, or file writing fails.
func (r *Renderer) renderToFile(contentTemplate string, data PageData, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	// Parse base.html with the specific content template
	tmpl, err := r.templates.Lookup("base.html").Clone()
	if err != nil {
		return fmt.Errorf("cloning base template: %w", err)
	}

	// Add the specific content template
	if _, err := tmpl.ParseFiles(filepath.Join("templates", contentTemplate)); err != nil {
		return fmt.Errorf("parsing content template: %w", err)
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

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

// parseAllPosts parses all markdown files in a directory using the provided parser.
//
// Scans the directory for .md files and calls parser.ParseFile on each one.
// Returns an empty slice if the directory doesn't exist (not an error).
//
// Parameters:
//   - p: Parser instance to use for markdown conversion
//   - dir: Directory path containing markdown files (e.g., "content/posts")
//
// Returns a slice of parsed Post structs or an error if parsing fails.
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

// filterDrafts removes draft posts from the list based on the "draft" frontmatter field.
//
// Posts with draft: true in their frontmatter are excluded from the published site.
//
// Parameters:
//   - posts: Slice of all parsed posts
//
// Returns a new slice containing only non-draft posts.
func filterDrafts(posts []*parser.Post) []*parser.Post {
	var published []*parser.Post
	for _, post := range posts {
		if !post.Draft {
			published = append(published, post)
		}
	}
	return published
}

// copyStatic recursively copies static assets (CSS, images, etc.) to the output directory.
//
// Walks the source directory tree and copies all files and directories to the destination,
// preserving directory structure and file permissions. Returns nil if source doesn't exist.
//
// Parameters:
//   - srcDir: Source directory containing static files (e.g., "static")
//   - dstDir: Destination directory in the output (e.g., "public")
//
// Returns an error if copying fails.
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
