# SSG - Static Site Generator

[![ci](https://github.com/kvnloughead/ssg/actions/workflows/ci.yml/badge.svg)](https://github.com/kvnloughead/ssg/actions/workflows/ci.yml)
[![Code style: djLint](https://img.shields.io/badge/html%20style-djLint-blue.svg)](https://github.com/djlint/djlint)

A minimal, fast static site generator written in Go. Converts markdown posts with YAML frontmatter into a complete static website. JavaScript is only used for progressive enhancements.

## Features

- **Markdown to HTML** - Parses markdown posts into HTML with [Goldmark](github.com/yuin/goldmark-highlighting/v2)
  - Syntax highlighting of fenced code blocks
  - Highlight specific lines of code by adding `{hl_lines=[1,3,5]}` after the fence and language name
  - Footnotes: `[^1]`
- **Copy buttons on code blocks** - this feature uses JS
- **YAML Frontmatter** - Rich metadata support (title, date, description, tags, draft status)
- **Draft Posts** - Mark posts as drafts to exclude them from the build. Posts are marked as drafts when they are created
- **Local Dev Server** - Built-in HTTP server for previewing your site locally
- **Live Reload** - Hot reload support with Air (optional)
- **Fast Builds** - Efficient single-binary executable with no external dependencies

## Installation

### Prerequisites

- Go 1.21 or higher

### Quick Start

Fork the repo and follow these steps:

```bash
# Clone the forked repo
git clone https://github.com/yourusername/ssg.git
cd ssg

# Install dependencies
make deps

# Recommended: install linting dependencies and pre-push hook
make init

# Create a config file
cat > config.yaml << EOF
title: Your Blog Title
description: An SSG built with ssg
baseUrl: https://yourblog.com
author: Your Name
keywords: Some, Keywords
EOF
```

The following templates are provided:

- `base.html` - Main layout
- `index.html` - Home page (posts list)
- `post.html` - Individual post page

Adjust them and the CSS as desired.

### 3. Create your first post

Standard markdown features are parsed. The welcome post in `content/posts` shows

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

The binary has three commands: `build`, `serve`, and `new`. You can run them all with `make`:

```bash
make build
make serve
make new TITLE="My Title"
```

You can also run them like so:

```bash
go run ./cmd/ssg build [flags]               # Build the static site
go run ./cmd/ssg serve [flags]               # Serve the site locally
go run ./cmd/ssg new --title "My Title"      # Create a new post
```

Run `make help` or `go run ./cmd/ssg` for more info on the commands and flags.

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
│       └── 2024-01-15-welcome.md
├── templates/                # HTML templates
│   ├── base.html             # Base layout
│   ├── index.html            # Home page
│   └── post.html             # Post page
├── static/                   # Static assets
│   ├── css/
│   │   └── style.css
│   ├── images/
│   └── js/
|       └── scripts...
├── public/                   # Generated site (output)
├── config.yaml               # Site configuration
└── Makefile                  # Build automation and convenience targets
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

## Template Data

Templates have access to:

```go
type PageData struct {
    Site  SiteConfig        // Site config (title, author, etc.)
    Post  *parser.Post      // Current post (on post pages)
    Posts []*parser.Post    // All posts (on index page)
    Title string            // Page title
}
```

## CI Pipeline

The `Makefile` provides targets for:

- linting Go code with staticcheck
- security checking Go code with gosec
- linting html tempaltes with djlint
- validating HTML with vnu
- formatting Go code and Go templates
- running custom Go tests

The easiest way to interact with these are with the `ci/*` targets:

```
ci/test                  run the test job like CI
ci/lint                  run the lint job like CI (static analysis + security)
ci/format                run the format job like CI
ci/local                 run full CI pipeline locally
```

If the `pre-push` hook is enabled, `ci/local` is run before pushes are allowed. A GitHub action workflow replicating the local pipeline is run when merging or pushing into main.

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

## Acknowledgments

- Built with [goldmark](https://github.com/yuin/goldmark) for markdown parsing
- Uses [yaml.v3](https://github.com/go-yaml/yaml) for frontmatter
- Template linting with [djlint](https://djlint.com/)
