#!/usr/bin/env node

/**
 * AI-Powered Documentation Updater
 * 
 * This script:
 * 1. Reads game source from internal/games/* and handler files (internal/api/*_handlers.go)
 * 2. Reads current wiki documentation
 * 3. Uses AI to analyze code and compare with docs
 * 4. Generates updated markdown files based on actual code implementation
 * 
 * Format follows WIKI_DOCUMENTATION_FORMAT.md specification
 */

import { readFileSync, writeFileSync, existsSync, readdirSync, statSync } from 'fs';
import { join, dirname, basename } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine project root (two levels up from .github/scripts/)
const PROJECT_ROOT = join(__dirname, '../..');
const WIKI_ROOT = process.env.WIKI_ROOT || join(PROJECT_ROOT, '..', 'flagged-it.wiki');
const GAMES_DIR = join(PROJECT_ROOT, 'internal', 'games');
const API_DIR = join(PROJECT_ROOT, 'internal', 'api');
const HANDLERS_FILE = join(API_DIR, 'handlers.go');
const WEB_GAMES_DIR = join(PROJECT_ROOT, 'web', 'src', 'lib', 'api');

/** Handler type name -> internal/games package or game key */
const HANDLER_TO_GAME = {
  WorldleGameHandler: 'guessing',
  FlagGameHandler: 'flag',
  FlagColorGameHandler: 'flagcolor',
};

const NON_GAME_HANDLERS = new Set(['DebugHandler']);

const WIKI_SPECIAL_FILES = new Set(['Home.md', '_Sidebar.md', '_Footer.md']);

// Groq API configuration
const GROQ_API_URL = 'https://api.groq.com/openai/v1/chat/completions';
const GROQ_API_KEY = process.env.GROQ_API_KEY;

// Model fallback chain - try models in order if one fails or hits rate limits
// Rate limits: All models have 30 RPM minimum, so we wait 2.5s between requests (24 RPM = safe buffer)
const GROQ_MODELS = [
  'llama-3.3-70b-versatile',              // Primary: best quality (30 RPM, 12k TPM)
  'meta-llama/llama-4-scout-17b-16e-instruct', // Fallback 1: next-gen performance (30 RPM, 30k TPM)
  'qwen/qwen3-32b',                       // Fallback 2: excellent for coding/math (60 RPM, 6k TPM)
  'llama-3.1-8b-instant'                  // Fallback 3: pure speed (30 RPM, 6k TPM)
];

if (!GROQ_API_KEY) {
  console.error('❌ Error: GROQ_API_KEY environment variable is not set');
  process.exit(1);
}

/**
 * Read file content
 */
function readFile(filePath) {
  try {
    return readFileSync(filePath, 'utf-8');
  } catch (error) {
    return null;
  }
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Convert snake_case to camelCase
 */
function toCamelCase(str) {
  return str.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
}

/**
 * Convert package name to potential handler name patterns
 */
function getHandlerNamePatterns(packageName) {
  // Convert package name to potential handler names
  // e.g., "higher_lower" -> ["HigherLowerHandler", "HigherLowerGameHandler"]
  const parts = packageName.split('_').map(p => 
    p.charAt(0).toUpperCase() + p.slice(1)
  );
  const camelName = parts.join('');
  
  return [
    `${camelName}GameHandler`,
    `${camelName}Handler`,
    // Special cases
    packageName === 'guessing' ? 'WorldleGameHandler' : null,
    packageName === 'flag' ? 'FlagGameHandler' : null,
    packageName === 'flagcolor' ? 'FlagColorGameHandler' : null,
  ].filter(Boolean);
}

/**
 * Convert CamelCase to snake_case
 */
function camelToSnakeCase(str) {
  return str.replace(/([a-z0-9])([A-Z])/g, '$1_$2').toLowerCase();
}

/**
 * Derive game key from a handler struct name
 */
function handlerNameToGameName(handlerName) {
  if (HANDLER_TO_GAME[handlerName]) {
    return HANDLER_TO_GAME[handlerName];
  }
  const base = handlerName.replace(/GameHandler$/, '').replace(/Handler$/, '');
  return camelToSnakeCase(base);
}

/**
 * Normalize game keys for deduplication (flag_color vs flagcolor)
 */
function normalizeGameKey(name) {
  return name.replace(/_/g, '').toLowerCase();
}

/**
 * List non-test Go source files in a game package directory
 */
function listPackageGoFiles(packageDir) {
  try {
    return readdirSync(packageDir)
      .filter((file) => file.endsWith('.go') && !file.endsWith('_test.go'))
      .map((file) => join(packageDir, file))
      .sort((a, b) => {
        if (a.endsWith('logic.go')) return -1;
        if (b.endsWith('logic.go')) return 1;
        return a.localeCompare(b);
      });
  } catch {
    return [];
  }
}

/**
 * Read handlers.go and internal/api/*_handlers.go
 */
function readAllHandlerFiles() {
  const paths = [HANDLERS_FILE];
  try {
    for (const entry of readdirSync(API_DIR)) {
      if (entry.endsWith('_handlers.go')) {
        paths.push(join(API_DIR, entry));
      }
    }
  } catch {
    // API directory missing
  }

  return paths
    .filter((filePath) => existsSync(filePath))
    .map((filePath) => ({ file: filePath, content: readFile(filePath) || '' }));
}

/**
 * Scan handler files for game API handler structs
 */
function scanGameHandlers(handlerFiles) {
  const handlers = [];
  const seen = new Set();

  for (const { file, content } of handlerFiles) {
    const regex = /type\s+(\w+Handler)\s+struct/g;
    let match;
    while ((match = regex.exec(content)) !== null) {
      const handlerName = match[1];
      if (NON_GAME_HANDLERS.has(handlerName) || seen.has(handlerName)) {
        continue;
      }
      seen.add(handlerName);
      handlers.push({ handlerName, handlerFile: file });
    }
  }

  return handlers;
}

/**
 * Find handler struct name for a game package across all handler files
 */
function findHandlerForPackage(packageName, handlerFiles) {
  for (const { file, content } of handlerFiles) {
    const handlerName = findHandlerName(packageName, content);
    if (handlerName) {
      return { handlerName, handlerFile: file };
    }
  }
  return null;
}

/**
 * Extract a handler struct and its methods from a Go source file
 */
function extractHandlerSection(handlerContent, handlerName) {
  if (!handlerContent || !handlerName) {
    return null;
  }

  const escaped = handlerName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const handlerMethodsRegex = new RegExp(
    `(type\\s+${escaped}[\\s\\S]*?func\\s+\\(h\\s+\\*${escaped}\\)[\\s\\S]*?)(?=type\\s+\\w+Handler|\\Z)`,
    'i'
  );
  const methodsMatch = handlerContent.match(handlerMethodsRegex);
  if (methodsMatch) {
    return methodsMatch[1].substring(0, 8000);
  }

  const handlerRegex = new RegExp(
    `(type\\s+${escaped}[\\s\\S]*?)(?=type\\s+\\w+Handler|\\Z)`,
    'i'
  );
  const match = handlerContent.match(handlerRegex);
  if (match) {
    return match[1].substring(0, 8000);
  }

  return handlerContent.substring(0, 8000);
}

/**
 * Find handler name in handlers file
 */
function findHandlerName(packageName, handlerContent) {
  const patterns = getHandlerNamePatterns(packageName);
  
  for (const pattern of patterns) {
    const regex = new RegExp(`type\\s+(${pattern})\\s+struct`, 'i');
    const match = handlerContent.match(regex);
    if (match) {
      return match[1];
    }
  }
  
  // Fallback: search for any handler with package name in it
  const searchTerm = packageName.replace('_', '');
  const handlerRegex = new RegExp(`type\\s+([A-Za-z]*${searchTerm}[A-Za-z]*Handler)\\s+struct`, 'i');
  const match = handlerContent.match(handlerRegex);
  if (match) {
    return match[1];
  }
  
  return null;
}

/**
 * Find web API file by convention
 */
function findWebApiFile(packageName) {
  const patterns = [
    // Special cases first (exact filenames)
    packageName === 'guessing' ? 'worldleGame.ts' : null,
    packageName === 'flagcolor' ? 'flagColorGame.ts' : null,
    toCamelCase(packageName) + 'Game.ts',
    packageName.replace('_', '') + 'Game.ts',
    packageName + 'Game.ts',
  ].filter(Boolean);
  
  for (const pattern of patterns) {
    const filePath = join(WEB_GAMES_DIR, pattern);
    if (existsSync(filePath)) {
      return filePath;
    }
  }
  
  // Try scanning directory for files that might match
  try {
    const files = readdirSync(WEB_GAMES_DIR);
    const searchTerm = packageName.replace('_', '').toLowerCase();
    for (const file of files) {
      if (file.toLowerCase().includes(searchTerm) && file.endsWith('.ts')) {
        return join(WEB_GAMES_DIR, file);
      }
    }
  } catch (e) {
    // Directory doesn't exist or can't read
  }
  
  return null;
}

/**
 * Discover games from internal/games packages and internal/api handlers
 */
function discoverGames() {
  const games = new Map();
  const handlerFiles = readAllHandlerFiles();

  for (const entry of readdirSync(GAMES_DIR)) {
    const packageDir = join(GAMES_DIR, entry);
    if (!statSync(packageDir).isDirectory()) {
      continue;
    }

    const goFiles = listPackageGoFiles(packageDir);
    if (goFiles.length === 0) {
      continue;
    }

    const handler = findHandlerForPackage(entry, handlerFiles);
    games.set(entry, {
      name: entry,
      packageDir,
      goFiles,
      handlerName: handler?.handlerName ?? null,
      handlerFile: handler?.handlerFile ?? null,
      webApiFile: findWebApiFile(entry),
    });
  }

  for (const { handlerName, handlerFile } of scanGameHandlers(handlerFiles)) {
    const gameName = handlerNameToGameName(handlerName);
    const normalized = normalizeGameKey(gameName);

    let alreadyDiscovered = false;
    for (const [key, game] of games) {
      if (
        normalizeGameKey(key) === normalized ||
        game.handlerName === handlerName
      ) {
        if (!game.handlerName) {
          game.handlerName = handlerName;
          game.handlerFile = handlerFile;
        }
        alreadyDiscovered = true;
        break;
      }
    }
    if (alreadyDiscovered) {
      continue;
    }

    games.set(gameName, {
      name: gameName,
      packageDir: null,
      goFiles: [],
      handlerName,
      handlerFile,
      webApiFile: findWebApiFile(gameName),
    });
  }

  return [...games.values()].sort((a, b) => a.name.localeCompare(b.name));
}

/**
 * Read game code files
 */
function readGameCode(game) {
  const code = {
    goLogic: null,
    goHandler: null,
    webApi: null,
    handlerSource: game.handlerFile ? basename(game.handlerFile) : 'handlers.go',
  };

  if (game.goFiles?.length) {
    const parts = [];
    let totalLength = 0;
    const maxLength = 12000;

    for (const filePath of game.goFiles) {
      const content = readFile(filePath);
      if (!content) {
        continue;
      }
      const chunk = `=== ${basename(filePath)} ===\n${content}`;
      if (totalLength + chunk.length > maxLength) {
        break;
      }
      parts.push(chunk);
      totalLength += chunk.length;
    }

    if (parts.length > 0) {
      code.goLogic = parts.join('\n\n');
    }
  }

  if (game.handlerName && game.handlerFile) {
    const handlerContent = readFile(game.handlerFile);
    code.goHandler = extractHandlerSection(handlerContent, game.handlerName);
  }

  if (game.webApiFile && existsSync(game.webApiFile)) {
    code.webApi = readFile(game.webApiFile);
  }

  return code;
}

/**
 * Read current wiki documentation
 */
function readCurrentDoc(gameWikiName) {
  const docFile = join(WIKI_ROOT, `${gameWikiName}.md`);
  if (existsSync(docFile)) {
    return readFileSync(docFile, 'utf-8');
  }
  return null;
}

/**
 * Generate wiki filename from package name - uses convention but can be overridden
 */
function generateWikiFileName(packageName, code) {
  // Check if there's an existing wiki file that matches
  // This handles cases where the naming might be different
  try {
    const wikiFiles = readdirSync(WIKI_ROOT);
    // Look for files that might match this game
    const searchTerms = [
      packageName.toLowerCase(),
      packageName.replace('_', '').toLowerCase(),
      packageName.replace('_', '-').toLowerCase()
    ];
    
    for (const file of wikiFiles) {
      if (file.endsWith('.md') && !WIKI_SPECIAL_FILES.has(file)) {
        const fileBase = file.replace('.md', '').toLowerCase();
        for (const term of searchTerms) {
          if (fileBase.includes(term) || term.includes(fileBase)) {
            return file.replace('.md', '');
          }
        }
      }
    }
  } catch (e) {
    // Can't read wiki directory, continue with convention
  }
  
  // Use convention mapping for known games
  const nameMap = {
    'guessing': 'Worldle',
    'flag': 'Guess-By-Flag',
    'shape': 'Shape-Game',
    'capital': 'Capital-City',
    'higher_lower': 'Higher-Lower',
    'hangman': 'Hangman',
    'facts': 'Facts',
    'flagcolor': 'Flag-Color',
  };
  
  // If we have a mapping, use it; otherwise convert package name
  if (nameMap[packageName]) {
    return nameMap[packageName];
  }
  
  // Convert package_name to Title-Case
  return packageName
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join('-');
}

/**
 * Call Groq API to analyze code and generate documentation
 */
async function generateDocWithAI(packageName, code, currentDoc) {
  const codeContext = [];
  
    if (code.goLogic) {
      codeContext.push(`=== Go Game Package (package: ${packageName}) ===\n${code.goLogic.substring(0, 12000)}`);
    }
    if (code.goHandler) {
      codeContext.push(`=== Go Handler (${code.handlerSource}) ===\n${code.goHandler.substring(0, 8000)}`);
    }
    if (code.webApi) {
      codeContext.push(`=== Web API (TypeScript) ===\n${code.webApi.substring(0, 4000)}`);
    }

  const currentDocSection = currentDoc 
    ? `\n\n=== Current Documentation ===\n${currentDoc}`
    : '\n\n=== No existing documentation ===';

  const prompt = `You are a technical documentation expert. Analyze the game code below and generate/update the wiki documentation following the format specification.

GAME PACKAGE NAME: ${packageName}

FORMAT SPECIFICATION:
- Start with **Purpose:** followed by a one-line description
- Then **Rules & Behavior:** section with bullet points
- No title heading (page name is the title)
- Rules should be in present tense
- Rules should reflect the ACTUAL implementation in the code
- Be specific about: rounds, scoring, input methods, validation, special features, error handling

CODE TO ANALYZE:
${codeContext.join('\n\n')}${currentDocSection}

INSTRUCTIONS:
1. Analyze the Go game package source to understand the game mechanics
2. Check the handler file for API-specific behavior (region filtering, locale support, difficulty modes)
3. Review web API file for frontend-specific behavior
4. Compare with current documentation and update it to match the code
5. Generate markdown following the exact format specified above
6. Include ALL important rules and behaviors from the code
7. Be accurate - only document what's actually implemented
8. The game package name is "${packageName}" - use this context to understand the game type

Return ONLY the markdown content, no explanations or code blocks.`;

  // Try each model in the fallback chain
  for (let modelIndex = 0; modelIndex < GROQ_MODELS.length; modelIndex++) {
    const model = GROQ_MODELS[modelIndex];
    const isLastModel = modelIndex === GROQ_MODELS.length - 1;
    const retriesPerModel = 2;

    for (let attempt = 0; attempt < retriesPerModel; attempt++) {
      try {
        const response = await fetch(GROQ_API_URL, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${GROQ_API_KEY}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            model: model,
            messages: [
              {
                role: 'system',
                content: 'You are a technical documentation expert. Return only valid markdown documentation, no explanations or code blocks.'
              },
              {
                role: 'user',
                content: prompt
              }
            ],
            temperature: 0.2,
            max_tokens: 4000,
          }),
        });

        if (response.status === 429) {
          // Custom wait times: attempt 1 = 1s, attempt 2 = 5s, then +5s per additional attempt
          // Ignore Retry-After header and use our custom pattern instead
          const waitTime = attempt === 0 ? 1000 : attempt * 5000; // 1s, 5s, 10s, 15s, ...
          
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Rate limited (429) on ${model}. Waiting ${waitTime / 1000}s before retry ${attempt + 1}/${retriesPerModel}...`);
            await sleep(waitTime);
            continue;
          } else {
            console.warn(`  ⚠️  Rate limited on ${model} after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000);
            break;
          }
        }

        if (!response.ok) {
          const errorText = await response.text();
          if (attempt === retriesPerModel - 1 && isLastModel) {
            throw new Error(`Groq API error: ${response.status} ${response.statusText} - ${errorText}`);
          }
          if (attempt < retriesPerModel - 1) {
            const waitTime = Math.min(1000 * Math.pow(2, attempt), 5000);
            console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${response.status}. Retrying...`);
            await sleep(waitTime);
            continue;
          } else {
            console.warn(`  ⚠️  Error on ${model} after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000);
            break;
          }
        }

        const data = await response.json();
        let markdown = data.choices[0]?.message?.content?.trim();

        if (!markdown) {
          throw new Error('No documentation received from API');
        }

        // Clean up markdown - remove code blocks if present
        if (markdown.startsWith('```markdown')) {
          markdown = markdown.replace(/^```markdown\n?/, '').replace(/\n?```$/, '');
        } else if (markdown.startsWith('```')) {
          markdown = markdown.replace(/^```\n?/, '').replace(/\n?```$/, '');
        }

        console.log(`  ✅ Successfully generated using ${model}`);
        return markdown.trim() + '\n';
      } catch (error) {
        if (attempt === retriesPerModel - 1 && isLastModel) {
          console.error(`  ❌ Error after trying all models:`, error.message);
          return null;
        }
        
        if (attempt < retriesPerModel - 1) {
          const waitTime = Math.min(1000 * Math.pow(2, attempt), 10000);
          console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${error.message}. Retrying...`);
          await sleep(waitTime);
        } else {
          console.warn(`  ⚠️  Error on ${model} after ${retriesPerModel} attempts: ${error.message}. Trying next model...`);
          await sleep(2000);
          break;
        }
      }
    }
  }

  return null;
}

/**
 * GitHub wiki page title from filename (Flag-Color.md -> Flag Color)
 */
function wikiDisplayTitle(fileBase) {
  return fileBase.replace(/-/g, ' ');
}

/**
 * Extract one-line Purpose from a game wiki page
 */
function parsePurposeFromDoc(content) {
  if (!content) {
    return '';
  }
  const match = content.match(/\*\*Purpose:\*\*\s*(.+)/);
  return match ? match[1].trim() : '';
}

/**
 * List game wiki pages (excludes Home, sidebar, footer)
 */
function listGameWikiPages() {
  return readdirSync(WIKI_ROOT)
    .filter((file) => file.endsWith('.md') && !WIKI_SPECIAL_FILES.has(file))
    .map((file) => {
      const fileBase = file.replace('.md', '');
      return {
        fileBase,
        displayTitle: wikiDisplayTitle(fileBase),
        path: join(WIKI_ROOT, file),
      };
    });
}

/**
 * Read existing [[Page Title]] order from Home.md Game Modes section
 */
function parseHomeGameOrder(homeContent) {
  if (!homeContent) {
    return [];
  }

  const sectionMatch = homeContent.match(/## Game Modes\n([\s\S]*?)(?=\n## |\nUse the sidebar|$)/);
  if (!sectionMatch) {
    return [];
  }

  const titles = [];
  const linkRegex = /^- \[\[([^\]]+)\]\]/gm;
  let match;
  while ((match = linkRegex.exec(sectionMatch[1])) !== null) {
    titles.push(match[1]);
  }
  return titles;
}

/**
 * Order wiki pages: keep Home.md order, append new pages alphabetically
 */
function orderWikiPages(pages, preferredTitles) {
  const byTitle = new Map(pages.map((page) => [page.displayTitle, page]));
  const ordered = [];

  for (const title of preferredTitles) {
    if (byTitle.has(title)) {
      ordered.push(byTitle.get(title));
      byTitle.delete(title);
    }
  }

  const remaining = [...byTitle.values()].sort((a, b) =>
    a.displayTitle.localeCompare(b.displayTitle)
  );
  return [...ordered, ...remaining];
}

/**
 * Replace the Game Modes bullet list in Home.md
 */
function replaceHomeGameModesSection(homeContent, gameLines) {
  const newSection = `## Game Modes\n\n${gameLines.join('\n')}\n`;
  const sectionRegex = /## Game Modes\n[\s\S]*?(?=\n## |\nUse the sidebar|$)/;

  if (sectionRegex.test(homeContent)) {
    return homeContent.replace(sectionRegex, newSection);
  }

  return `${homeContent.trim()}\n\n${newSection}`;
}

/**
 * Sync Home.md and _Sidebar.md with all game wiki pages
 */
function updateWikiNavigation() {
  const pages = listGameWikiPages();
  if (pages.length === 0) {
    return { updated: false, changedFiles: [] };
  }

  const homePath = join(WIKI_ROOT, 'Home.md');
  const sidebarPath = join(WIKI_ROOT, '_Sidebar.md');
  const homeContent = readFile(homePath) || '';
  const ordered = orderWikiPages(pages, parseHomeGameOrder(homeContent));

  const gameLines = ordered.map((page) => {
    const purpose = parsePurposeFromDoc(readFile(page.path));
    return purpose
      ? `- [[${page.displayTitle}]] - ${purpose}`
      : `- [[${page.displayTitle}]]`;
  });

  const newHome = replaceHomeGameModesSection(homeContent, gameLines);
  const newSidebar = `## Games\n\n- [[Home]]\n${ordered.map((page) => `- [[${page.displayTitle}]]`).join('\n')}\n`;

  const changedFiles = [];

  if (newHome.trim() !== homeContent.trim()) {
    writeFileSync(homePath, newHome, 'utf-8');
    changedFiles.push('Home.md');
  }

  const sidebarContent = readFile(sidebarPath) || '';
  if (newSidebar.trim() !== sidebarContent.trim()) {
    writeFileSync(sidebarPath, newSidebar, 'utf-8');
    changedFiles.push('_Sidebar.md');
  }

  return { updated: changedFiles.length > 0, changedFiles };
}

/**
 * Main function
 */
async function main() {
  console.log('🚀 Starting AI-powered documentation update...\n');

  if (!existsSync(WIKI_ROOT)) {
    console.error(`❌ Wiki directory not found: ${WIKI_ROOT}`);
    console.error('   Set WIKI_ROOT environment variable or ensure wiki is cloned as sibling directory');
    process.exit(1);
  }

  const games = discoverGames();
  console.log(`✓ Found ${games.length} games to process\n`);

  let filesUpdated = 0;
  let filesGenerated = 0;
  let filesSkipped = 0;

  for (const game of games) {
    console.log(`\n📋 Processing ${game.name}...`);

    // Read code files
    const code = readGameCode(game);
    if (!code.goLogic && !code.goHandler) {
      console.warn(`  ⚠️  No game source or handler found, skipping`);
      filesSkipped++;
      continue;
    }

    // Generate wiki filename (let AI determine or use convention)
    const gameWikiName = generateWikiFileName(game.name, code);
    console.log(`  📝 Wiki file: ${gameWikiName}.md`);

    // Read current documentation
    const currentDoc = readCurrentDoc(gameWikiName);

    // Generate updated documentation with AI
    console.log(`  🤖 Analyzing code with AI...`);
    const newDoc = await generateDocWithAI(game.name, code, currentDoc);

    if (!newDoc) {
      console.error(`  ❌ Failed to generate documentation`);
      filesSkipped++;
      continue;
    }

    // Write updated documentation
    const docFile = join(WIKI_ROOT, `${gameWikiName}.md`);
    const shouldUpdate = !currentDoc || currentDoc.trim() !== newDoc.trim();

    if (shouldUpdate) {
      try {
        writeFileSync(docFile, newDoc, 'utf-8');
        if (currentDoc) {
          console.log(`  ✅ Updated ${gameWikiName}.md`);
          filesUpdated++;
        } else {
          console.log(`  ✨ Created ${gameWikiName}.md`);
          filesGenerated++;
        }
      } catch (error) {
        console.error(`  ❌ Error writing ${docFile}:`, error.message);
        filesSkipped++;
      }
    } else {
      console.log(`  ✓ ${gameWikiName}.md - No changes needed`);
    }

    // Rate limiting: wait between API calls to respect rate limits
    // Minimum RPM is 30 (llama-3.3-70b, llama-4-scout, llama-3.1-8b)
    // Using 2.5s delay = 24 RPM, safely under 30 RPM limit to prevent blocking
    await sleep(2500);
  }

  console.log('\n📚 Syncing wiki navigation (Home.md, _Sidebar.md)...');
  const nav = updateWikiNavigation();
  if (nav.updated) {
    console.log(`  ✅ Updated: ${nav.changedFiles.join(', ')}`);
  } else {
    console.log('  ✓ Navigation already up to date');
  }

  console.log('\n' + '='.repeat(50));
  console.log(`✅ Documentation update complete!`);
  console.log(`   Files generated: ${filesGenerated}`);
  console.log(`   Files updated: ${filesUpdated}`);
  console.log(`   Navigation updated: ${nav.changedFiles.join(', ') || 'none'}`);
  console.log(`   Files skipped: ${filesSkipped}`);
  console.log('='.repeat(50));
}

// Run the script
main().catch(error => {
  console.error('❌ Fatal error:', error);
  process.exit(1);
});
