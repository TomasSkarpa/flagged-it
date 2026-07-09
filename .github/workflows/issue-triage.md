---
name: Issue Triage
description: Label new issues and leave a brief, helpful first response.
emoji: "👀"
labels: ["support", "triage"]
on:
  issues:
    types: [opened]
  reaction: "eyes"
permissions:
  contents: read
  issues: read
  copilot-requests: write
safe-outputs:
  add-labels:
    allowed: [bug, enhancement, question, documentation]
    max: 1
  add-comment:
    max: 1
---

# Triage New Issues

When a new issue is opened, classify it and leave a concise first response.

## Labeling

Choose exactly one of these existing labels:

- `bug`
- `enhancement`
- `question`
- `documentation`

Use these hints:

- `bug` for broken behavior, regressions, crashes, wrong answers, missing assets, or incorrect gameplay
- `enhancement` for feature ideas, new game modes, UX improvements, account systems, scoring ideas, or roadmap items
- `question` for compatibility, usage, setup, or clarification requests
- `documentation` for README, setup instructions, contribution docs, or unclear developer guidance

## Comment style

Write one short comment that is easy to scan:

1. Greet the author by username
2. Restate your understanding in one sentence
3. Ask at most two concrete questions only if essential information is missing
4. End with one short sentence about next-step expectation

## Guardrails

- Keep the comment under 120 words
- Do not be overly cheerful or verbose
- Do not ask generic boilerplate questions
- If the report is already clear, do not ask for more information
- Never promise a fix date
