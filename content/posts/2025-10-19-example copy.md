---
title: Example Blog Post
date: 2025-10-10T00:00:00Z
description: An introduction to this static site generator and how it works
tags: [blogging, golang, ssg]
draft: false
---

This is an example blog post written in **Markdown** and generated using a custom static site generator written in Go.

## Features

This SSG supports:

- GitHub Flavored Markdown
- Footnotes[^1]
- YAML frontmatter
- Posts marked as drafts won't be published
- Tags and metadata

## Code Examples

Here's a simple Go function:

```go {hl_lines=[2]}
func greet(name string) string {
  return fmt.Sprintf("Hello, %s!", name)
}
```

Line two is highlighted by writing this after the code fence: `go {hl_lines=[2]}`. The copy button is a progressive enhancement that won't be rendered if JS is not used.

## Lists

Unordered list:
- Item one
- Item two
- Item three

Ordered list:
1. First step
2. Second step
3. Third step

## Quotes

> "The best way to predict the future is to invent it."
> — Alan Kay

## Footnotes

You can add footnotes to your posts. They'll be automatically linked[^2].

[^1]: Footnotes are automatically rendered at the bottom of the page.
[^2]: This is the second footnote demonstrating the feature.

---

That's it! Start writing your own posts with `ssg new --title "Your Title"`.