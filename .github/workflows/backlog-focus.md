---
name: Backlog Focus
description: Create a short weekly backlog issue that highlights the most valuable work to focus on next.
emoji: "🧭"
labels: ["planning", "product"]
on:
  schedule: weekly
  workflow_dispatch:
permissions:
  contents: read
  issues: read
  pull-requests: read
  copilot-requests: write
engine:
  model: mini
tools:
  github:
    mode: gh-proxy
    toolsets: [issues, pull_requests, repos]
safe-outputs:
  create-issue:
    max: 1
---

# Weekly Backlog Focus

Review the current open issues and open pull requests in this repository, with extra attention to product-facing work.

Your goal is to help a solo maintainer decide what to do next without reading the whole backlog.

Produce at most one new GitHub issue only when there is something meaningfully useful to summarize.

## What to analyze

- Open issues, especially unlabeled or long-open enhancement ideas
- Open pull requests, if any
- Recent backlog-focus issues, so you do not create repetitive summaries

## Output requirements

If you create an issue, title it:

`Backlog Focus - <date>`

Keep the body short and structured:

1. `Snapshot` — 2 sentences max describing the current state of the repo
2. `Do Now` — exactly 3 bullets, each with issue number, title, and one short reason
3. `Can Wait` — up to 3 bullets for lower-priority items
4. `One Recommended Move` — one concrete next action the maintainer could do in the next 30 to 90 minutes

## Guardrails

- Optimize for clarity over completeness
- Do not restate every issue
- Prefer actionable work over ambitious future ideas
- If there is no meaningful change since the most recent open backlog-focus issue, do nothing
- If the backlog is tiny or already obvious, do nothing
- Keep the whole issue under 220 words
