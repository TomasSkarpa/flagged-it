# Flag color quiz — SVG annotations

Guessable regions are marked on Twemoji flag SVGs under `web/static/assets/twemoji_flags_cca2/`.

## Attributes

| Attribute | Required | Meaning |
|-----------|----------|---------|
| `data-fi-guess` | Yes | Stable id within the file (e.g. `1`, `2`). Used by API + client to pick the region that shows the player’s live color preview. |
| `data-fi-tier` | No | `easy`, `hard`, or `both`. Omitted means `both`. |
| `data-fi-merge-group` | No | Same fill but **must stay separate challenges** (e.g. Egypt field white vs emblem white): give each region its own token; paths merge only when fill **and** group match. Omit for normal flags (merge all shards of the same hex). |
| `id` | Optional | `fi-guess-{same as data-fi-guess}` for convenience; client may select by attribute. |

Apply annotations only to `<path>`, `<circle>`, or `<rect>` with a **solid** `fill="#RRGGBB"` or `#RGB`. Do not tag gradients or patterns.

## Merging duplicate colours

Run once after editing SVGs so each distinct `(fill + optional merge-group)` maps to one `data-fi-guess` id (preview paints every matching path via `querySelectorAll`; the server dedupes picks):

`go run ./internal/games/flagcolor/cmd/merge-flag-guess-colors [path/to/twemoji_flags_cca2]`

The tool strips `data-fi-merge-group` from output; add it again before re-running if you still need separate slots for the same hex.

Any country may be drawn for flag color mode; the server skips SVGs with no `data-fi-guess` regions (or none matching the round difficulty) and picks another.

## Lint

Run `go test ./internal/games/flagcolor/...` — tests validate parsing and scoring helpers.
