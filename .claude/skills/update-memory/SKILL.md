---
name: update-memory
description: Updates CLAUDE.md with new learnings, conventions, or context. Use when user says "remember this" or after discovering important project patterns.
---

# Update Memory

## When to Use
- User says "remember this" or "remember globally"
- A build/test command is discovered
- An important convention is identified
- A bug pattern is found and resolved

## Process
1. Determine scope: global (~/.claude/CLAUDE.md) or project (./CLAUDE.md)
2. Find the `## Learned` section
3. Append a concise bullet: `- [topic]: [what was learned]`
4. Do not duplicate existing entries
5. Keep each entry to one line
