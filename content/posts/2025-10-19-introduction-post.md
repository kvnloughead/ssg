---
title: Introduction Post
date: 2025-10-19T00:00:00Z
description: An introduction to this static site generator and how it works
tags: [blogging, golang, ssg]
draft: false
---

This is an example blog post written in **Markdown** and generated using a custom static site generator written in Go.

## Features

This SSG supports:

- [GitHub Flavored Markdown](https://docs.github.com/en/get-started/writing-on-github/getting-started-with-writing-and-formatting-on-github/basic-writing-and-formatting-syntax)
- Footnotes[^1]
- YAML frontmatter
- Posts marked as drafts won't be published
- Tags and metadata
- [HTML in markdown](../posts/html-inside-markdown.html)

## Progressive Enhancements

These features require JavaScript, but the application works fine without them.

- Copy buttons in code blocks

    ```go
    // Try me
    ```

## Standard markdown features

### Code Blocks

Code blocks are syntax highlighted, and support highlighting one or more lines of code[^2].

```go {hl_lines=[2]}
func greet(name string) string {
  return fmt.Sprintf("Hello, %s!", name)
}
```

### Lists

Unordered list:
- Item one
- Item two
- Item three

Ordered list:
1. First step
2. Second step
3. Third step

### Quotes

> "The best way to predict the future is to invent it."
> — Alan Kay

[^1]: You can add footnotes to your posts. with this syntax: `[^1]`. Add the corresponding `[^1]` to the bottom of the page and they will be linked.
[^2]: You choose the lines to highlight by writing this after the code fence and language: `{hl_lines=[1,2,3]}`.

## Get started

That's it! Start writing your own posts with `ssg new --title "Your Title"`.