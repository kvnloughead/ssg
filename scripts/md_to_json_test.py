#!/usr/bin/python3

"""
Unit tests for md_to_json.py

Run with: python3 -m unittest md_to_json_test.py
"""

import unittest
import tempfile
import os
import json
import shutil
from md_to_json import parse_frontmatter, process_markdown_file, process_posts_directory


class TestParseFrontmatter(unittest.TestCase):
    """Tests for the parse_frontmatter function"""

    def test_parse_basic_frontmatter(self):
        """Test parsing basic YAML frontmatter with string values"""
        content = """---
title: Test Post
date: 2025-10-19T00:00:00Z
description: A test post
---

This is the content."""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertEqual(frontmatter["title"], "Test Post")
        self.assertEqual(frontmatter["date"], "2025-10-19T00:00:00Z")
        self.assertEqual(frontmatter["description"], "A test post")
        self.assertEqual(markdown, "This is the content.")

    def test_parse_frontmatter_with_array(self):
        """Test parsing frontmatter with array values"""
        content = """---
tags: [python, testing, markdown]
categories: [dev, tutorial]
---

Content here."""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertEqual(frontmatter["tags"], ["python", "testing", "markdown"])
        self.assertEqual(frontmatter["categories"], ["dev", "tutorial"])

    def test_parse_frontmatter_with_booleans(self):
        """Test parsing frontmatter with boolean values"""
        content = """---
draft: true
published: false
featured: True
archived: False
---

Content."""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertTrue(frontmatter["draft"])
        self.assertFalse(frontmatter["published"])
        self.assertTrue(frontmatter["featured"])
        self.assertFalse(frontmatter["archived"])

    def test_parse_frontmatter_with_null(self):
        """Test parsing frontmatter with null values"""
        content = """---
author: null
reviewer: NULL
---

Content."""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertIsNone(frontmatter["author"])
        self.assertIsNone(frontmatter["reviewer"])

    def test_parse_no_frontmatter(self):
        """Test parsing markdown without frontmatter"""
        content = "Just plain markdown content."

        frontmatter, markdown = parse_frontmatter(content)

        self.assertEqual(frontmatter, {})
        self.assertEqual(markdown, content)

    def test_parse_empty_frontmatter(self):
        """Test parsing with empty frontmatter section"""
        content = """---
---

Content here."""

        frontmatter, markdown = parse_frontmatter(content)

        # Empty frontmatter (no key:value pairs) still returns empty dict
        self.assertEqual(frontmatter, {})
        # The content is returned as-is since the regex doesn't match empty frontmatter
        self.assertEqual(markdown, content)

    def test_parse_frontmatter_ignores_comments(self):
        """Test that comments in frontmatter are ignored"""
        content = """---
title: Test
# This is a comment
author: John
---

Content."""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertEqual(frontmatter["title"], "Test")
        self.assertEqual(frontmatter["author"], "John")
        self.assertNotIn("# This is a comment", frontmatter)

    def test_parse_multiline_content(self):
        """Test parsing with multiline markdown content"""
        content = """---
title: Multi-line Test
---

Line 1

Line 2

Line 3"""

        frontmatter, markdown = parse_frontmatter(content)

        self.assertEqual(frontmatter["title"], "Multi-line Test")
        self.assertIn("Line 1", markdown)
        self.assertIn("Line 2", markdown)
        self.assertIn("Line 3", markdown)


class TestProcessMarkdownFile(unittest.TestCase):
    """Tests for the process_markdown_file function"""

    def setUp(self):
        """Create a temporary directory for test files"""
        self.test_dir = tempfile.mkdtemp()

    def tearDown(self):
        """Clean up temporary directory"""
        shutil.rmtree(self.test_dir)

    def test_process_valid_markdown_file(self):
        """Test processing a valid markdown file"""
        test_file = os.path.join(self.test_dir, "test.md")
        content = """---
title: Test Post
tags: [test, markdown]
draft: false
---

# Hello World

This is a test post."""

        with open(test_file, "w", encoding="utf-8") as f:
            f.write(content)

        result = process_markdown_file(test_file)

        self.assertEqual(result["file"], "test.md")
        self.assertEqual(result["frontmatter"]["title"], "Test Post")
        self.assertEqual(result["frontmatter"]["tags"], ["test", "markdown"])
        self.assertFalse(result["frontmatter"]["draft"])
        self.assertIn("# Hello World", result["content"])

    def test_process_file_without_frontmatter(self):
        """Test processing a markdown file without frontmatter"""
        test_file = os.path.join(self.test_dir, "plain.md")
        content = "# Just Markdown\n\nNo frontmatter here."

        with open(test_file, "w", encoding="utf-8") as f:
            f.write(content)

        result = process_markdown_file(test_file)

        self.assertEqual(result["file"], "plain.md")
        self.assertEqual(result["frontmatter"], {})
        self.assertEqual(result["content"], content)

    def test_process_utf8_content(self):
        """Test processing markdown with UTF-8 characters"""
        test_file = os.path.join(self.test_dir, "utf8.md")
        content = """---
title: UTF-8 Test
---

Unicode characters: 你好世界 🚀 café"""

        with open(test_file, "w", encoding="utf-8") as f:
            f.write(content)

        result = process_markdown_file(test_file)

        self.assertIn("你好世界", result["content"])
        self.assertIn("🚀", result["content"])
        self.assertIn("café", result["content"])


class TestProcessPostsDirectory(unittest.TestCase):
    """Tests for the process_posts_directory function"""

    def setUp(self):
        """Create a temporary directory structure with test markdown files"""
        self.test_dir = tempfile.mkdtemp()

        # Create multiple markdown files
        self.files = {
            "post1.md": """---
title: First Post
date: 2025-10-19
---

First post content.""",
            "post2.md": """---
title: Second Post
date: 2025-10-20
tags: [test, example]
---

Second post content.""",
            "readme.txt": "Not a markdown file",  # Should be ignored
            "draft.md": """---
title: Draft Post
draft: true
---

Draft content.""",
        }

        for filename, content in self.files.items():
            filepath = os.path.join(self.test_dir, filename)
            with open(filepath, "w", encoding="utf-8") as f:
                f.write(content)

        # Create a subdirectory with a markdown file
        subdir = os.path.join(self.test_dir, "subdir")
        os.makedirs(subdir)
        with open(os.path.join(subdir, "nested.md"), "w", encoding="utf-8") as f:
            f.write("""---
title: Nested Post
---

Nested content.""")

    def tearDown(self):
        """Clean up temporary directory"""
        shutil.rmtree(self.test_dir)

    def test_process_all_markdown_files(self):
        """Test that all markdown files are processed"""
        posts = process_posts_directory(self.test_dir)

        # Should find 4 .md files (post1, post2, draft, nested)
        self.assertEqual(len(posts), 4)

        # Extract titles from all posts
        titles = [post["frontmatter"].get("title") for post in posts]

        self.assertIn("First Post", titles)
        self.assertIn("Second Post", titles)
        self.assertIn("Draft Post", titles)
        self.assertIn("Nested Post", titles)

    def test_process_ignores_non_markdown_files(self):
        """Test that non-.md files are ignored"""
        posts = process_posts_directory(self.test_dir)

        # readme.txt should not be in the results
        filenames = [post["file"] for post in posts]
        self.assertNotIn("readme.txt", filenames)

    def test_process_includes_nested_files(self):
        """Test that nested directories are processed"""
        posts = process_posts_directory(self.test_dir)

        # Find the nested post
        nested_posts = [p for p in posts if p["file"] == "nested.md"]
        self.assertEqual(len(nested_posts), 1)
        self.assertEqual(nested_posts[0]["frontmatter"]["title"], "Nested Post")

    def test_process_empty_directory(self):
        """Test processing an empty directory"""
        empty_dir = tempfile.mkdtemp()
        try:
            posts = process_posts_directory(empty_dir)
            self.assertEqual(posts, [])
        finally:
            shutil.rmtree(empty_dir)

    def test_process_directory_with_only_non_markdown(self):
        """Test directory containing only non-markdown files"""
        test_dir = tempfile.mkdtemp()
        try:
            # Create some non-markdown files
            with open(os.path.join(test_dir, "readme.txt"), "w") as f:
                f.write("Text file")
            with open(os.path.join(test_dir, "image.png"), "w") as f:
                f.write("fake image")

            posts = process_posts_directory(test_dir)
            self.assertEqual(posts, [])
        finally:
            shutil.rmtree(test_dir)


class TestEndToEnd(unittest.TestCase):
    """End-to-end tests simulating actual usage"""

    def setUp(self):
        """Create a realistic test directory structure"""
        self.test_dir = tempfile.mkdtemp()

    def tearDown(self):
        """Clean up"""
        shutil.rmtree(self.test_dir)

    def test_full_pipeline(self):
        """Test the full pipeline from files to JSON output"""
        # Create test files
        posts_content = [
            {
                "filename": "2025-10-19-first-post.md",
                "content": """---
title: Introduction Post
date: 2025-10-19T00:00:00Z
description: An introduction
tags: [blogging, golang]
draft: false
---

This is the first post.""",
            },
            {
                "filename": "2025-10-20-second-post.md",
                "content": """---
title: Second Post
date: 2025-10-20T00:00:00Z
tags: [update]
---

This is the second post.""",
            },
        ]

        for post in posts_content:
            filepath = os.path.join(self.test_dir, post["filename"])
            with open(filepath, "w", encoding="utf-8") as f:
                f.write(post["content"])

        # Process directory
        results = process_posts_directory(self.test_dir)

        # Verify results
        self.assertEqual(len(results), 2)

        # Verify JSON serialization works
        json_output = json.dumps(results, indent=2, ensure_ascii=False)
        self.assertIsInstance(json_output, str)

        # Parse it back
        parsed = json.loads(json_output)
        self.assertEqual(len(parsed), 2)


if __name__ == "__main__":
    unittest.main()
