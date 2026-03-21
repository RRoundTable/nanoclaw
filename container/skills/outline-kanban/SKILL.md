---
name: outline-kanban
description: Manage kanban boards and project tasks in Outline — list task boards, create tasks, move tasks between statuses (todo, in-progress, review, done, blocked). Use whenever the user mentions tasks, kanban, project tracking, task management, to-do items, work status, or wants to organize and track work items.
allowed-tools: Bash(python3:*)
---

# Outline Kanban

Kanban board management built on top of Outline wiki. Tasks are regular Outline documents with emoji-prefixed titles to represent status.

```
KANBAN="/workspace/group/outline-cli/skills/kanban.py"
```

## Status values

| Command key | Emoji | Meaning |
|-------------|-------|---------|
| `todo` | 📋 | To do |
| `progress` | 🔄 | In progress |
| `review` | 👀 | In review |
| `done` | ✅ | Done |
| `blocked` | 🚫 | Blocked |

`doing` is an alias for `progress`. Status keys are case-insensitive.

## Commands

```bash
# View the kanban board (grouped by status)
python3 $KANBAN list [COLLECTION_ID]

# Move a task to a different status
python3 $KANBAN move DOC_ID progress
python3 $KANBAN move DOC_ID done
python3 $KANBAN move DOC_ID blocked

# Create a new task with initial status
python3 $KANBAN create "Task name" todo COLLECTION_ID
python3 $KANBAN create "Build REST API" progress COLLECTION_ID
```

## Board output example

```
📋 TODO (2)
  [def67890] Build REST API
  [ghi01234] Write documentation
🔄 IN PROGRESS (1)
  [jkl56789] Setup CI/CD pipeline
✅ DONE (3)
  [mno23456] Initial project setup
```

## Typical workflow

1. **View board**: `python3 $KANBAN list COLLECTION_ID`
2. **Pick a task**: note the short ID from the listing
3. **Start work**: `python3 $KANBAN move DOC_ID progress`
4. **Complete work**: `python3 $KANBAN move DOC_ID done`

## Default collection

The default project collection ID is `60fa3861-441d-4e8c-aa3d-4955063fd5d5`. Use this when no specific collection is mentioned.
