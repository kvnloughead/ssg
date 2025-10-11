 ssg/
  ├── cmd/
  │   └── ssg/
  │       └── main.go           # CLI entry point
  ├── internal/
  │   ├── parser/               # Markdown parsing logic
  │   ├── renderer/             # Template rendering
  │   └── builder/              # Site building orchestration
  ├── content/
  │   └── posts/                # Markdown files
  │       └── 2024-01-01-my-post.md
  ├── templates/
  │   ├── base.html
  │   ├── post.html
  │   └── index.html
  ├── static/                   # CSS, JS, images
  │   ├── css/
  │   └── images/
  ├── public/                   # Generated output (gitignored)
  └── config.yaml              # Site configuration