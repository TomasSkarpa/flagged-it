# AGENTS.md

Short guide for anyone (human or agent) working on flagged-it.

## What it is

Geography quiz games. Go API + SvelteKit frontend. Optional desktop build with Wails.

| Where | What |
|---|---|
| `cmd/web` | API server entry (`go run main.go -dev`) |
| `internal/api` | HTTP routes (`server.go`) and game handlers (`handlers.go`, `*_handlers.go`) |
| `internal/games/<game>` | Game logic, one folder per game |
| `internal/data`, `assets/` | Country data, flags, shapes (embedded) |
| `web/src/routes/<game>` | Game pages |
| `web/src/lib/api` | Frontend API clients, one per game |
| `web/src/lib/components/ui` | Shared UI kit (preview at `/ui-kit`) |
| `web/src/lib/translations/locales` | UI strings |

## Run

```bash
make setup   # install deps
make dev     # API :8080 + web :5173
make check   # go vet/fmt + svelte check — run before every PR
```

## Things to know

- Game state lives on the server in in-memory sessions. The client never gets the answer before it submits one.
- Hangman is the exception: it runs fully in the browser.
- Multiplayer ("party") lives in `internal/party` (rooms, rounds, scoring) and `web/src/routes/party`. Live updates use server-sent events. Check it end to end with `go run ./cmd/party-smoke -api <url>`.
- Add new UI strings to `en.json` only. A workflow translates the other locales and opens a PR.
- Wiki game pages are generated from `internal/games` by a workflow. Don't edit them by hand.
- The server's systemd unit and nginx config live in the IaC repo, not here.

## Conventions

- Branch: kebab-case, e.g. `country-guessing-game`.
- Commit: `feat:` `fix:` `refactor:` `enhance:` `chore:` `docs:` + short lowercase message.

## How we work (new features and big reworks)

Everything is tracked in GitHub Issues so there is one log of what changed, in which PR, and why.
Use `gh` for all of it. Small fixes can skip steps 1–2.

1. **Issue first.** Use the template. Title `🌍 Feature Request: <name>` (or a plain title for bugs).
   Write: what and why → **Acceptance criteria** → **Open questions** → **Requires** (#issue).
2. **Big work = epic.** Add the `epic` label and split it into sub-issues (one PR each).
3. **Start.** Comment on the issue with a short plan: 2–4 bullets.
4. **Branch + PR.** One issue per PR. The PR body says `Closes #N`, then **What**, **How**, **Why**.
5. **Log on the issue** when the PR is opened and when it is merged. Use this format:

   ```
   **Update** — PR #123 (open|merged)
   - Done: what changed, in one or two lines
   - Why: the reason, or what it unblocks
   - Next: the next step, or "nothing, closing"
   ```

6. **Close.** Merging closes the issue. Tick the sub-issue in the epic.

Keep every issue, comment and PR short and plain. Use bullets, not essays.
