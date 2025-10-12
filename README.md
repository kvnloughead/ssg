# SSG - Static Site Generator

[![Code style: djLint](https://img.shields.io/badge/html%20style-djLint-blue.svg)](https://github.com/djlint/djlint)



A minimal, fast static site generator written in Go. Converts markdown posts with YAML frontmatter into a complete static website.

## Features

- **Markdown to HTML** - Parse markdown files with GitHub Flavored Markdown, footnotes, and smart typography
- **YAML Frontmatter** - Rich metadata support (title, date, description, tags, draft status)
- **Template Inheritance** - DRY templates using Go's html/template with base layouts
- **Draft Posts** - Mark posts as drafts to exclude them from the build
- **Local Dev Server** - Built-in HTTP server for previewing your site
- **Live Reload** - Hot reload support with Air (optional)
- **Fast Builds** - Efficient single-binary executable with no external dependencies

## Installation

### Prerequisites

- Go 1.21 or higher

### Build from source

```bash
git clone https://github.com/yourusername/ssg.git
cd ssg
make build
```

The binary will be created at `bin/ssg`.

### Install globally

```bash
go install github.com/yourusername/ssg/cmd/ssg@latest
```

## Quick Start

### 1. Initialize your project

```bash
# Clone or create a new directory
mkdir my-blog && cd my-blog

# Create the required directories
mkdir -p content/posts templates static/css

# Create a config file
cat > config.yaml << EOF
title: My Blog
description: A blog built with SSG
baseUrl: https://yourblog.com
author: Your Name
keywords: Programming, Technology
EOF
```

### 2. Create templates

See `templates/` directory for examples. You need at minimum:
- `base.html` - Main layout
- `index.html` - Home page (posts list)
- `post.html` - Individual post page

### 3. Create your first post

```bash
ssg new --title "My First Post"
```

This creates a file like `content/posts/2024-01-15-my-first-post.md`:

```markdown
---
title: My First Post
date: 2024-01-15T10:00:00Z
description: "My first blog post"
tags: [introduction, hello]
draft: false
---

Write your post here in **markdown**!
```

### 4. Build and serve

```bash
# Build the site
ssg build

# Serve locally
ssg serve

# Or use live reload with Air
make run/air
```

Visit `http://localhost:8080` to see your site!

## Usage

### Commands

```bash
ssg build [flags]    # Build the static site
ssg serve [flags]    # Serve the site locally
ssg new [flags]      # Create a new post
```

### Build Command

```bash
ssg build --output public --config config.yaml
```

**Flags:**
- `--output` - Output directory (default: `public`)
- `--config` - Config file path (default: `config.yaml`)

### Serve Command

```bash
ssg serve --port 8080
```

**Flags:**
- `--port` - Port to serve on (default: `8080`)

### New Command

```bash
ssg new --title "My Post Title"
```

**Flags:**
- `--title` - Post title (required)

## Project Structure

```
.
├── cmd/
│   └── ssg/
│       └── main.go           # CLI entry point
├── internal/
│   ├── parser/
│   │   └── parser.go         # Markdown + frontmatter parser
│   └── ssg/
│       └── ssg.go            # Site generation logic
├── content/
│   └── posts/                # Your markdown posts
│       ├── 2024-01-15-first-post.md
│       └── 2024-01-16-second-post.md
├── templates/                # HTML templates
│   ├── base.html             # Base layout
│   ├── index.html            # Home page
│   └── post.html             # Post page
├── static/                   # Static assets
│   ├── css/
│   │   └── style.css
│   └── images/
├── public/                   # Generated site (output)
├── config.yaml               # Site configuration
└── Makefile                  # Build automation
```

## Configuration

Edit `config.yaml`:

```yaml
title: My Blog
description: A blog about programming and technology
baseUrl: https://yourblog.com
author: Your Name
keywords: Programming, Golang, Technology
```

## Frontmatter

Posts support the following frontmatter fields:

```yaml
---
title: Post Title              # Required
date: 2024-01-15T10:00:00Z    # Required (RFC3339 format)
description: Post description  # Optional
tags: [tag1, tag2]            # Optional
draft: false                   # Optional (default: false)
---
```

## Templates

Templates use Go's `html/template` syntax with inheritance:

### base.html

```html
<!DOCTYPE html>
<html>
  <head>
    <title>{{.Title}} | {{.Site.Title}}</title>
    <link rel="stylesheet" href="/css/style.css" />
  </head>
  <body>
    <header>
      <h1><a href="/">{{.Site.Title}}</a></h1>
    </header>
    <main>{{template "content" .}}</main>
    <footer>© {{.Site.Author}}</footer>
  </body>
</html>
```

### index.html

```html
{{define "content"}}
<div class="posts-list">
  {{range .Posts}}
  <article>
    <h2><a href="/posts/{{.Slug}}.html">{{.Title}}</a></h2>
    <time>{{.Date.Format "January 2, 2006"}}</time>
  </article>
  {{end}}
</div>
{{end}}
```

### Template Data

Templates have access to:

```go
type PageData struct {
    Site  SiteConfig        // Site config (title, author, etc.)
    Post  *parser.Post      // Current post (on post pages)
    Posts []*parser.Post    // All posts (on index page)
    Title string            // Page title
}
```

## Development

### Make Commands

```bash
make build                    # Build the binary
make generate                 # Generate the site
make serve                    # Serve the site
make test                     # Run tests
make test/cover              # Run tests with coverage
make lint                    # Run linters
make format                  # Format code
make ci/local                # Run full CI pipeline locally
```

### Live Reload with Air

For development with automatic rebuilding:

```bash
make run/air
```

This watches for changes and automatically rebuilds the site.

### Linting Templates

```bash
make lint/templates          # Lint templates with djlint
make format/templates        # Format templates with djlint
```

## Architecture

### Package Overview

- **`cmd/ssg`** - CLI application entry point
- **`internal/parser`** - Markdown + frontmatter parsing (reusable, no deps on other packages)
- **`internal/ssg`** - Site generation orchestration (uses parser, handles rendering and file I/O)

### Build Flow

1. Load `config.yaml`
2. Parse all markdown files in `content/posts/` → `parser.Post` structs
3. Filter out drafts, sort by date
4. Load templates from `templates/`
5. Render `index.html` with all posts
6. Render individual post pages
7. Copy static assets to output directory

### Template Inheritance

Templates use a base + content pattern:

1. `base.html` provides layout with `{{template "content" .}}`
2. `index.html` and `post.html` define `{{define "content"}}` blocks
3. At render time, base is cloned and content template is parsed into it
4. Result is a complete page with shared layout

## Testing

```bash
# Run all tests
make test

# Run with coverage
make test/cover

# Run specific package
go test ./internal/parser
go test ./internal/ssg
```

## Deployment

The generated `public/` directory contains a complete static site. Deploy to any static hosting:

### Netlify

```bash
# Build command
make generate

# Publish directory
public
```

### GitHub Pages

```bash
# Build and push to gh-pages branch
make generate
cd public
git init
git add .
git commit -m "Deploy"
git push -f git@github.com:username/repo.git main:gh-pages
```

### Any Static Host

Just upload the contents of `public/` to your web server.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow the patterns in `CLAUDE.md`
- Add tests for new functionality
- Update documentation
- Run `make ci/local` before submitting

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [goldmark](https://github.com/yuin/goldmark) for markdown parsing
- Uses [yaml.v3](https://github.com/go-yaml/yaml) for frontmatter
- Template linting with [djlint](https://djlint.com/)