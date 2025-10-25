#!/usr/bin/python3

"""
Converts markdown files with YAML frontmatter to JSON.
Loops through content/posts directory and processes all .md files.
"""

import os
import json
import re


def parse_frontmatter(content):
    """
    Parses YAML frontmatter from markdown content.

    Args:
        content: The full markdown file content as a string

    Returns:
        A tuple of (frontmatter_dict, markdown_content)
    """
    # Match content between --- delimiters
    pattern = r"^---\s*\n(.*?)\n---\s*\n(.*)"
    match = re.match(pattern, content, re.DOTALL)

    if not match:
        return {}, content

    frontmatter_text = match.group(1)
    markdown_content = match.group(2).strip()

    # Parse YAML frontmatter manually (simple key: value pairs)
    frontmatter = {}
    for line in frontmatter_text.split("\n"):
        line = line.strip()
        if not line or line.startswith("#"):
            continue

        # Handle "key: value" format
        if ":" in line:
            key, value = line.split(":", 1)
            key = key.strip()
            value = value.strip()

            # Handle arrays [item1, item2]
            if value.startswith("[") and value.endswith("]"):
                # Remove brackets and split by comma
                items = value[1:-1].split(",")
                value = [item.strip() for item in items]
            # Handle booleans
            elif value.lower() == "true":
                value = True
            elif value.lower() == "false":
                value = False
            # Handle null
            elif value.lower() == "null":
                value = None

            frontmatter[key] = value

    return frontmatter, markdown_content


def process_markdown_file(file_path):
    """
    Processes a single markdown file and returns its data as a dict.

    Args:
        file_path: Path to the markdown file

    Returns:
        Dictionary with flattened frontmatter fields and content
    """
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    frontmatter, markdown_content = parse_frontmatter(content)

    # Start with flattened frontmatter
    result = dict(frontmatter)

    # Convert tags array to space-separated string if present
    if "tags" in result and isinstance(result["tags"], list):
        result["tags"] = " ".join(result["tags"])

    # Add content as 'body' (standard field for tinysearch)
    result["body"] = markdown_content

    # Generate URL from filename (remove .md extension)
    # Assumes format like "2025-10-19-post-title.md" -> "/posts/post-title"
    filename = os.path.basename(file_path)
    slug = filename.replace(".md", "")
    result["url"] = f"/posts/{slug}"

    return result


def process_posts_directory(posts_dir):
    """
    Loops through the posts directory and processes all markdown files.

    Args:
        posts_dir: Path to the posts directory

    Returns:
        List of processed post dictionaries
    """
    posts = []

    for root, _, file_names in os.walk(posts_dir):
        for file_name in file_names:
            if file_name.endswith(".md"):
                file_path = os.path.join(root, file_name)
                try:
                    post_data = process_markdown_file(file_path)
                    posts.append(post_data)
                    print(f"Processed: {file_name}", file=__import__("sys").stderr)
                except Exception as e:
                    print(
                        f"Error processing {file_name}: {e}",
                        file=__import__("sys").stderr,
                    )

    return posts


if __name__ == "__main__":
    # Get the posts directory path relative to script location
    script_dir = os.path.dirname(os.path.abspath(__file__))
    posts_directory = os.path.join(script_dir, "..", "content", "posts")

    # Normalize the path to resolve .. properly
    posts_directory = os.path.abspath(posts_directory)

    # Process all markdown files
    posts = process_posts_directory(posts_directory)

    # Output as JSON to stdout
    print(json.dumps(posts, indent=2, ensure_ascii=False))
