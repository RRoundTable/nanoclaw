---
name: claude-session
description: Manage Claude Code subprocess sessions for development and planning tasks. Use when delegating work to claude -p, resuming previous sessions, running parallel tasks with worktrees, or managing session lifecycle. Provides session creation, resume, done-tagging, and message routing patterns.
allowed-tools: Bash(claude:*), Bash(cd:*)
---

# Claude Code Session Management

Delegate work to `claude -p` as a subprocess. Sessions persist on disk and can be resumed.

## Core Commands

```bash
# New task — creates a new session in the project dir
cd /workspace/extra/my-playground/src/<project>
claude --dangerously-skip-permissions -p "<task description>" -n "<project>: <short task name>"

# Continue previous work (resumes most recent session in current dir)
cd /workspace/extra/my-playground/src/<project>
claude --dangerously-skip-permissions -p "<follow-up instruction>" --continue

# Resume a specific session by ID
claude --dangerously-skip-permissions -p "<follow-up>" --resume <session-id>

# Mark session as done (tag with [done] in name + generate summary)
claude --dangerously-skip-permissions -p "summarize what was done in this session" \
  --resume <session-id> -n "[done] <project>: <short task name>"
```

## When to use what

| Situation | Command |
|-----------|---------|
| New project or new task | `claude -p "<task>" -n "<name>"` |
| Follow-up on same task | `claude -p "<follow-up>" --continue` |
| User says "done" | `claude -p "summarize" --resume <id> -n "[done] <name>"` |
| Parallel tasks on same repo | `claude -p "<task>" --worktree <branch>` |
| Big task, break into steps | Multiple `claude -p` calls with `--continue` |

## Session Storage

Sessions are auto-organized by Claude Code at `~/.claude/projects/<dir-slug>/`:
- `<session-id>.jsonl` — transcript (one per session)
- `<session-id>/` — subagent logs and tool results

The dir-slug is the `cwd` path with `/` replaced by `-`. Each project dir gets its own session namespace.

## Message Routing

On each incoming user message:

1. **Identify the project** the message relates to
2. **cd to that project dir** (`/workspace/extra/my-playground/src/<project>`)
3. **Decide**:
   - Continuing previous work → use `--continue`
   - New unrelated task → fresh `claude -p` with `-n`
   - User says "done" / "완료" → tag session with `[done]`
4. **Report** the claude output to the user

## Worktree for Parallel Work

For git repos only — isolates work on separate branches:

```bash
cd /workspace/extra/my-playground/src/<project>

# Task A on its own branch
claude --dangerously-skip-permissions -p "<task-A>" --worktree feature-a -n "<project>: task A"

# Task B on another branch (parallel, isolated)
claude --dangerously-skip-permissions -p "<task-B>" --worktree feature-b -n "<project>: task B"
```

## Splitting Big Requests

Break large requests into sequential `claude -p` calls in the same dir:

```bash
cd /workspace/extra/my-playground/src/<project>

# Step 1: Setup
claude --dangerously-skip-permissions -p "set up the project structure and install dependencies" -n "<project>: initial setup"

# Step 2: Continue with next piece
claude --dangerously-skip-permissions -p "implement the API endpoints" --continue

# Step 3: Continue with testing
claude --dangerously-skip-permissions -p "add tests for all endpoints" --continue
```

Each `--continue` resumes the same session, so context carries over.

## Rules

- **Never write code directly** — always delegate to `claude -p`
- **One subject per session** — don't mix unrelated tasks
- **Name sessions descriptively** with `-n` — format: `<project>: <task>`
- **Tag done sessions** — use `-n "[done] ..."` when user confirms completion
- **Report progress** — after each `claude -p` call, summarize the output to the user
- **Record in Outline** — update sprint/task docs with progress at each milestone
