# Wiki Documentation Format Specification

This document defines the standard format for game mode documentation in the wiki.

## File Structure

Each game mode documentation file follows this structure:

```markdown
**Purpose:** [One-line description of what the game mode does]

**Rules & Behavior:**

- [Rule 1]
- [Rule 2]
- [Rule 3]
...
```

## Format Rules

1. **No Title Heading**: The page name already serves as the title, so do not include a `# Title` heading at the top.

2. **Purpose Section**: 
   - Must start with `**Purpose:**` followed by a single-line description
   - Should be concise and clear about what the game mode does

3. **Rules & Behavior Section**:
   - Must start with `**Rules & Behavior:**`
   - Each rule should be a bullet point starting with `-`
   - Rules should be written in present tense
   - Rules should be specific and actionable
   - Rules should reflect the actual implementation in the code

## Information to Extract from Code

The documentation generator extracts the following information from game logic files:

### Core Game Properties
- `roundCount`: Number of rounds in a game (extracted from `roundCount` field in Logic struct)
- `purpose`: Description of what the game does (from comments or inferred)

### Game-Specific Rules
- Scoring mechanics (when score increases)
- Round mechanics (how rounds work, when they end)
- Input methods (multiple choice, text input, etc.)
- Validation rules (what constitutes a valid guess/answer)
- Special features (region filtering, difficulty modes, etc.)
- Error handling (what happens on invalid input)
- Completion conditions (when game ends)

### Common Patterns to Document
- Region filtering support
- Locale/translation support
- Country selection logic (random, no repeats, etc.)
- Answer validation (case-insensitive, multiple formats, etc.)
- Feedback mechanisms (correct/incorrect, hints, etc.)

## Example

```markdown
**Purpose:** Guess countries based on interesting facts about them.

**Rules & Behavior:**

- Game consists of 5 rounds
- Each round selects a random country that has facts available
- Facts are displayed one at a time (randomly selected from available facts for that country)
- Player has 3 tries per round to guess the correct country
- Each wrong guess reduces tries remaining
- Player can skip a round (counts as using all tries)
- Guess history is maintained showing each guess and the fact that was displayed
- Score increases by 1 only if you guess correctly within 3 tries
- Total rounds increase regardless of success or failure
- If you run out of tries, the correct answer is revealed and the round ends
- Only countries with facts data are included in the game
- Country matching supports multiple name formats (common, official, codes, translations)
- Game ends after 5 rounds total
- Facts are never repeated within the same round
```

## Generation Process

1. Parse Go game packages (`internal/games/*`) and handler files (`internal/api/handlers.go`, `internal/api/*_handlers.go`)
2. Extract game properties and rules
3. Parse handler files (`internal/api/handlers.go`) for API-specific behavior
4. Generate markdown files in the wiki format
5. Update wiki files (or create PR with changes)

## Usage

### AI-Powered Documentation Updates

The documentation system uses AI to analyze your code and automatically update documentation.

#### Local Development

To update documentation locally:

```bash
export GROQ_API_KEY=your_groq_api_key
make docs
```

This will:
1. Read Go logic files (`internal/games/*/logic.go`)
2. Read handler files (`internal/api/handlers.go`)
3. Read web API files (`web/src/lib/api/*Game.ts`)
4. Read current wiki documentation
5. Use AI to analyze code and compare with docs
6. Generate updated markdown files
7. Update files in the `../flagged-it.wiki` directory (if it exists)
8. Sync `Home.md` and `_Sidebar.md` so new game pages appear in the wiki navigation

**Note:** 
- The wiki must be cloned as a sibling directory to the main repository for local generation to work
- Alternatively, set the `WIKI_ROOT` environment variable: `WIKI_ROOT=/path/to/wiki make docs`
- You need a Groq API key (free tier available at https://console.groq.com)

#### Automated Updates

The GitHub Actions workflow (`.github/workflows/docs.yml`) automatically:
- Runs when game logic files change
- Uses AI to analyze code and generate documentation
- Creates a pull request with the changes

The workflow is triggered on:
- Push to `main` branch when `internal/games/**/*.go` or `internal/api/handlers.go` changes
- Manual trigger via `workflow_dispatch`

**Required Secret:** `GROQ_API_KEY` must be set in repository secrets for the workflow to run.

### Adding a New Game Mode

When adding a new game mode:

1. Create the game logic in `internal/games/<package-name>/` (any `.go` files, not only `logic.go`)
2. Add handler in `internal/api/handlers.go` or `internal/api/<name>_handlers.go`
3. Add web API file in `web/src/lib/api/<gameName>Game.ts` (if needed)
4. Run `node .github/scripts/docs.js` (with `GROQ_API_KEY` and `WIKI_ROOT` set) to generate documentation
5. The automated workflow will keep docs updated going forward

No manual configuration needed - the AI analyzes your code and generates appropriate documentation!

### Legacy Method (Code-Based Extraction)

If you prefer the old code-based extraction method (without AI), you can use:

```bash
make docs-legacy
```

This uses hardcoded rules to extract information. The AI-powered method is recommended as it's more flexible and accurate.
