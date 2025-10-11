// Package renderer handles HTML template rendering
package renderer

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/kvnloughead/ssg/internal/parser"
)

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
}

// Config holds site configuration for templates
type Config struct {
	Title       string
	Description string
	BaseURL     string
	Author      string
}

// PageData holds data passed to templates
type PageData struct {
	Site  Config
	Post  *parser.Post
	Posts []*parser.Post
	Title string
}

// New creates a new Renderer with loaded templates
func New(templateDir string) (*Renderer, error) {
	// Load all templates
	tmpl, err := template.ParseGlob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("loading templates: %w", err)
	}

	return &Renderer{templates: tmpl}, nil
}

// RenderPost renders a single post page
func (r *Renderer) RenderPost(post *parser.Post, config Config, outputPath string) error {
	data := PageData{
		Site:  config,
		Post:  post,
		Title: post.Title,
	}

	return r.renderToFile("post.html", data, outputPath)
}

// RenderIndex renders the index/home page with a list of posts
func (r *Renderer) RenderIndex(posts []*parser.Post, config Config, outputPath string) error {
	data := PageData{
		Site:  config,
		Posts: posts,
		Title: config.Title,
	}

	return r.renderToFile("index.html", data, outputPath)
}

// renderToFile renders a template to a file
func (r *Renderer) renderToFile(templateName string, data PageData, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	// Execute template
	if err := r.templates.ExecuteTemplate(f, templateName, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	return nil
}