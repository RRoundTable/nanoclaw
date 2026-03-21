---
name: outline-cli
description: Manage Outline wiki documents and collections — create, read, update, delete docs, list collections, and search content. Use whenever the user asks about wiki pages, documentation, notes, knowledge base, or wants to look up or modify content in Outline. Also use when you need to store or retrieve structured information persistently.
allowed-tools: Bash(python3:*)
---

# Outline CLI

CLI tool for managing the Outline wiki at `https://outline.nocoders.ai`.

```
CLI_PATH="/workspace/group/outline-cli/outline.py"
```

## Collections

```bash
# List all collections
python3 $CLI_PATH collections list
# Output: [short_id] Name (N docs) /urlId

# List as JSON (for parsing)
python3 $CLI_PATH collections list --json

# Create collection
python3 $CLI_PATH collections create "Collection Name"

# Delete collection
python3 $CLI_PATH collections delete COLLECTION_ID
```

## Documents

```bash
# List docs in a collection
python3 $CLI_PATH docs list --collection COLLECTION_ID
python3 $CLI_PATH docs list --collection COLLECTION_ID --json

# Show a document's content
python3 $CLI_PATH docs show DOC_ID
python3 $CLI_PATH docs show DOC_ID --json

# Create a document
python3 $CLI_PATH docs create --title "Title" --collection COLLECTION_ID
python3 $CLI_PATH docs create --title "Title" --collection COLLECTION_ID --text "Markdown content"

# Update a document
python3 $CLI_PATH docs update DOC_ID --title "New Title"
python3 $CLI_PATH docs update DOC_ID --text "New content"

# Delete a document
python3 $CLI_PATH docs delete DOC_ID
```

## Search

```bash
python3 $CLI_PATH search "query"
python3 $CLI_PATH search "query" --json --limit 10
```

## Tips

- Both full UUIDs and short IDs (first 8 chars) work as arguments
- Use `--json` when you need to parse output programmatically
- All documents are created as published (immediately visible)
- Content is Markdown format
- Default project collection: `60fa3861-441d-4e8c-aa3d-4955063fd5d5`

## Inline parsing pattern

```bash
python3 $CLI_PATH docs list --collection ID --json | python3 -c "
import sys, json
docs = json.load(sys.stdin)
for d in docs:
    print(d['id'][:8], d['title'])
"
```
