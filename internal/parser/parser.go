// Package parser handles parsing markdown files with frontmatter. It utilizes
// yuin/goldmark and gopkg.in/yaml.v3.
package parser

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v3"
)

// Post represents a parsed markdown post with frontmatter
type Post struct {
	Title       string
	Date        time.Time
	Slug        string
	Description string
	Tags        []string
	Keywords    string // Comma-separated string of tags
	Draft       bool
	Content     template.HTML // Unescaped HTML content
	RawContent  string        // Original markdown
}

// Frontmatter represents the YAML frontmatter
type Frontmatter struct {
	Title       string    `yaml:"title"`
	Date        time.Time `yaml:"date"`
	Description string    `yaml:"description"`
	Tags        []string  `yaml:"tags"`
	Draft       bool      `yaml:"draft"`
}

// Parser handles markdown parsing with goldmark
type Parser struct {
	md goldmark.Markdown
}

// New creates a new Parser with goldmark configured.
//   - Extensions: GitHub Flavored, footnotes, smart punctuation
//   - Auto-generate heading ID's
//   - newlines -> <br>
func New() *Parser {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,         // GitHub Flavored Markdown
			extension.Footnote,    // Footnote support
			extension.Typographer, // Smart punctuation
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(), // Auto-generate heading IDs
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(), // Convert newlines to <br>
			html.WithXHTML(),     // Use more strict XML-style tags
		),
	)

	return &Parser{md: md}
}

// ParseFile reads and parses a markdown file with YAML frontmatter.
//
// This is the main entry point for parsing posts. It reads the file from disk
// and delegates to Parse() for the actual parsing logic.
//
// Parameters:
//   - path: Filesystem path to the markdown file (e.g., "content/posts/2024-01-15-my-post.md")
//
// Returns a Post struct containing parsed frontmatter and converted HTML content,
// or an error if file reading or parsing fails.
func (p *Parser) ParseFile(path string) (*Post, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return p.Parse(content, path)
}

// Parse parses markdown content with YAML frontmatter into a Post struct.
//
// Expected format:
//
//	---
//	title: Post Title
//	date: 2024-01-15T10:00:00Z
//	description: "Post description"
//	tags: [tag1, tag2]
//	draft: false
//	---
//
//	Markdown content here...
//
// Process:
//  1. Splits content on "---" delimiters to extract frontmatter
//  2. Parses YAML frontmatter into structured data
//  3. Converts markdown to HTML using goldmark (with GFM, footnotes, etc.)
//  4. Generates a URL-friendly slug from the filename
//  5. Returns a Post struct with both HTML (Content) and original markdown (RawContent)
//
// Parameters:
//   - content: Raw file content as bytes
//   - path: File path (used only for slug generation)
//
// Returns a Post struct or an error if parsing fails.
func (p *Parser) Parse(content []byte, path string) (*Post, error) {
	// Split frontmatter and content
	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid frontmatter format")
	}

	// Parse frontmatter
	var fm Frontmatter
	if err := yaml.Unmarshal(parts[1], &fm); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}

	// Parse markdown content
	var buf bytes.Buffer
	markdown := bytes.TrimSpace(parts[2])
	if err := p.md.Convert(markdown, &buf); err != nil {
		return nil, fmt.Errorf("converting markdown: %w", err)
	}

	// Generate slug from filename
	slug := generateSlug(path)

	post := &Post{
		Title:       fm.Title,
		Date:        fm.Date,
		Slug:        slug,
		Description: fm.Description,
		Tags:        fm.Tags,
		Keywords:    strings.Join(fm.Tags, ", "),

		Draft: fm.Draft,
		// #nosec G203 -- HTML output from goldmark md parser, not from user input
		Content:    template.HTML(buf.String()),
		RawContent: string(markdown),
	}

	return post, nil
}

// generateSlug creates a URL-friendly slug from a file path. It extracts the
// filename, removes the extension, and strips the date prefix if present.
//
// For example: "content/posts/2024-01-15-my-first-post.md" → "my-first-post"
//
// This slug is used in the final URL: /posts/my-first-post.html
//
// Parameters:
//   - path: File path to generate slug from
//
// Returns the slug string.
func generateSlug(path string) string {
	filename := filepath.Base(path)
	// Remove extension
	slug := strings.TrimSuffix(filename, filepath.Ext(filename))
	// Remove date prefix if present (YYYY-MM-DD-)
	if len(slug) > 11 && slug[4] == '-' && slug[7] == '-' && slug[10] == '-' {
		slug = slug[11:]
	}
	return slug
}
