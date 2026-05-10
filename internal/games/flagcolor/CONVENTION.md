# Flag color quiz — SVG annotations

Guessable regions are marked on Twemoji flag SVGs under `web/static/assets/twemoji_flags_cca2/`.

## Attributes

| Attribute | Required | Meaning |
|-----------|----------|---------|
| `data-fi-guess` | Yes | Stable id within the file (e.g. `1`, `2`). Used by API + client to pick the region that shows the player’s live color preview. |
| `data-fi-tier` | No | `easy`, `hard`, or `both`. Omitted means `both`. |
| `id` | Optional | `fi-guess-{same as data-fi-guess}` for convenience; client may select by attribute. |

Apply annotations only to `<path>`, `<circle>`, or `<rect>` with a **solid** `fill="#RRGGBB"` or `#RGB`. Do not tag gradients or patterns.

Any country may be drawn for flag color mode; the server skips SVGs with no `data-fi-guess` regions (or none matching the round difficulty) and picks another.

## Lint

Run `go test ./internal/games/flagcolor/...` — tests validate parsing and scoring helpers.
