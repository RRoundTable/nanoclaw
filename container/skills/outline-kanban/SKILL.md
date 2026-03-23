---
name: outline-kanban
description: Manage project structure and tasks in Outline using nested documents — organize work into categories with child documents for individual items. Also handles automatic structure auditing, progress recording, and document restructuring for simplicity. Use whenever the user mentions tasks, project tracking, sprints, PRDs, roadmaps, backlog, user stories, progress updates, document cleanup, or wants to organize and track work items.
allowed-tools: Bash(outline:*), Bash(/workspace/extra/outline-cli/bin/outline:*)
---

# Outline Project Structure

Organize projects using Outline's nested document hierarchy. Top-level documents are categories, child documents are individual items.

```
OUTLINE="/workspace/extra/outline-cli/bin/outline"
```

## Document hierarchy

```
Collection
├── PRD/               (parent doc — product requirements)
│   ├── Feature A PRD  (child doc)
│   └── Feature B PRD
├── Roadmap/           (parent doc — timeline and milestones)
├── Sprint/            (parent doc — sprint plans)
│   ├── Sprint 1       (child doc — tasks as checklist)
│   └── Sprint 2
└── Backlog/           (parent doc — user stories and ideas)
    ├── User Story 1   (child doc)
    └── User Story 2
```

Max depth: 2 levels (category → item). Never nest deeper.

## Commands

```bash
# Create a category (top-level parent doc)
$OUTLINE docs create --title "Sprint" --collection $COLL

# Create an item under a category
$OUTLINE docs create --title "Sprint 1" --collection $COLL --parent SPRINT_DOC_ID --text "## Tasks
- [ ] Task 1
- [ ] Task 2"

# List items in a category
$OUTLINE docs children SPRINT_DOC_ID

# View an item
$OUTLINE docs show ITEM_DOC_ID

# Update an item
$OUTLINE docs update ITEM_DOC_ID --text "Updated content"

# Move an item (delete + recreate under new parent)
# Note: Outline API doesn't support reparenting — recreate the doc under the new parent
```

## Structure audit

Before starting work, scan the collection to detect structure issues. Run this check at the start of every task session.

```bash
# 1. List all top-level docs in the collection
$OUTLINE docs list --collection $COLL --json

# 2. For each top-level doc, list children
$OUTLINE docs children DOC_ID --json
```

**Issues to detect and fix:**

| Problem | Fix |
|---------|-----|
| Orphan docs (top-level but not a category) | Move under the right category (delete + recreate with --parent) |
| Empty categories (no children, no content) | Delete them |
| Duplicate categories (e.g. two "Sprint" docs) | Merge children into one, delete the duplicate |
| Deeply nested docs (child of a child) | Flatten to category → item (2 levels max) |
| Stale items (all tasks checked, sprint done) | Archive by prefixing title with `[Done]` or delete if no longer needed |
| Docs with no title or meaningless names | Rename based on content |

**Restructure procedure:**
1. List all top-level docs and their children
2. Identify issues from the table above
3. Fix each issue (smallest change possible)
4. Report what was changed

## Progress recording

Record progress directly in task documents using markdown checklists.

**When completing a task:**
```bash
# 1. Read current content
$OUTLINE docs show TASK_DOC_ID

# 2. Update with checked items and status note
$OUTLINE docs update TASK_DOC_ID --text "## Tasks
- [x] Task 1 — done (2026-03-23)
- [ ] Task 2
- [x] Task 3 — done (2026-03-23)

## Progress Log
- 2026-03-23: Completed Task 1 and Task 3. Task 2 blocked on API key."
```

**Rules for progress recording:**
- Check off completed items with `[x]` and add the date
- Add a `## Progress Log` section at the bottom of task docs — append, never overwrite
- Keep log entries to one line each: date + what happened
- When all tasks are done, update the doc title to include `[Done]`

## Typical workflow

1. **Audit structure** — scan collection, fix issues (every session start)
2. **Read current sprint** — check what tasks are assigned
3. **Do the work**
4. **Record progress** — update task doc with checked items and log entry
5. **Report** — brief summary to the user in chat
