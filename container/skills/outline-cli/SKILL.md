---
name: outline-cli
description: Manage Outline wiki documents and collections — create, read, update, delete docs, list collections, and search content. Use whenever the user asks about wiki pages, documentation, notes, knowledge base, or wants to look up or modify content in Outline. Also use when you need to store or retrieve structured information persistently.
allowed-tools: Bash(outline:*), Bash(/workspace/extra/outline-cli/bin/outline:*)
---

# Outline CLI

CLI tool for managing the Outline wiki at `https://outline.nocoders.ai`.

```
OUTLINE="/workspace/extra/outline-cli/bin/outline"
```

## Collections

```bash
# List all collections
$OUTLINE collections list

# List as JSON (for parsing)
$OUTLINE collections list --json

# Create collection
$OUTLINE collections create "Collection Name"

# Delete collection
$OUTLINE collections delete COLLECTION_ID
```

## Documents

```bash
# List docs in a collection
$OUTLINE docs list --collection COLLECTION_ID
$OUTLINE docs list --collection COLLECTION_ID --json

# List child docs under a parent
$OUTLINE docs list --parent PARENT_DOC_ID
$OUTLINE docs children PARENT_DOC_ID

# Show a document's content
$OUTLINE docs show DOC_ID
$OUTLINE docs show DOC_ID --json

# Create a document
$OUTLINE docs create --title "Title" --collection COLLECTION_ID
$OUTLINE docs create --title "Title" --collection COLLECTION_ID --text "Markdown content"

# Create a child document (nested under a parent)
$OUTLINE docs create --title "Title" --collection COLLECTION_ID --parent PARENT_DOC_ID

# Update a document
$OUTLINE docs update DOC_ID --title "New Title"
$OUTLINE docs update DOC_ID --text "New content"

# Delete a document
$OUTLINE docs delete DOC_ID
```

## Search

```bash
$OUTLINE search "query"
$OUTLINE search "query" --json --limit 10
```

## Tips

- Both full UUIDs and short IDs (first 8 chars) work as arguments
- Use `--json` when you need to parse output programmatically
- All documents are created as published (immediately visible)
- Content is Markdown format
- Use `--parent` to create hierarchical document structures
- Default project collection: `60fa3861-441d-4e8c-aa3d-4955063fd5d5`
