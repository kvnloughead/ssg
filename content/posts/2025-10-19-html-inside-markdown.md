---
title: HTML Inside Markdown
date: 2025-10-18T00:00:00Z
description: You can use HTML inside your markdown files too.
tags: [markdown, HTML]
draft: false
---



You can also embed HTML in your markdown files.

> ⚠️ Warning:
> Not safe for use with user-generated content. With user generated content, make sure to sanitize is first.
> For sanitizing HTML, use: https://github.com/microcosm-cc/bluemonday

## Examples

### Toggles

```md
<details>
  <summary>Click me</summary>
  <p>Write the hidden text inside a single <code>div</code> tag to get the nice border-left indicating the contents of the toggle.</p>
</details>
```

Result:

<details>
  <summary>Click me</summary>
  <div>
    <p>Write the hidden text inside a single <code>div</code> tag to get the nice border-left indicating the contents of the toggle.</p>
  </div>
</details>

### Checkboxes

```md
<input type="checkbox" /> Step 1
<input type="checkbox" /> Step 2
<input type="checkbox" /> Step 3
```

Result:

<input type="checkbox" /> Step 1
<input type="checkbox" /> Step 2
<input type="checkbox" /> Step 3
