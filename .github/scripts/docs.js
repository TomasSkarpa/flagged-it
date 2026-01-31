#!/usr/bin/env node

/**
 * AI-Powered Documentation Updater
 * 
 * This script:
 * 1. Reads game logic files (Go and Svelte/TypeScript)
 * 2. Reads current wiki documentation
 * 3. Uses AI to analyze code and compare with docs
 * 4. Generates updated markdown files based on actual code implementation
 * 
 * Format follows WIKI_DOCUMENTATION_FORMAT.md specification
 */

import { readFileSync, writeFileSync, existsSync, readdirSync, statSync } from 'fs';
import { join, dirname, extname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine project root (two levels up from .github/scripts/)
const PROJECT_ROOT = join(__dirname, '../..');
const WIKI_ROOT = process.env.WIKI_ROOT || join(PROJECT_ROOT, '..', 'flagged-it.wiki');
const GAMES_DIR = join(PROJECT_ROOT, 'internal', 'games');
const HANDLERS_FILE = join(PROJECT_ROOT, 'internal', 'api', 'handlers.go');
const WEB_GAMES_DIR = join(PROJECT_ROOT, 'web', 'src', 'lib', 'api');

// Groq API configuration
const GROQ_API_URL = 'https://api.groq.com/openai/v1/chat/completions';
const GROQ_API_KEY = process.env.GROQ_API_KEY;

const GROQ_MODELS = [
  'llama-3.3-70b-versatile',
  'llama-3.1-70b-versatile',
  'llama-3.1-8b-instant',
  'mixtral-8x7b-32768'
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
  ].filter(Boolean);
}

/**
 * Find handler name in handlers file
 */
function findHandlerName(packageName, handlerContent) {
  const patterns = getHandlerNamePatterns(packageName);
  
  for (const pattern of patterns) {
    const regex = new RegExp(`type\\s+${pattern}\\s+struct`, 'i');
    if (regex.test(handlerContent)) {
      return pattern;
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
  // Try common patterns
  const patterns = [
    toCamelCase(packageName) + 'Game.ts',
    packageName.replace('_', '') + 'Game.ts',
    packageName + 'Game.ts',
    // Special cases
    packageName === 'guessing' ? 'worldleGame.ts' : null,
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
 * Get all game directories - automatically discovers games
 */
function getGameDirectories() {
  const games = [];
  const entries = readdirSync(GAMES_DIR);
  const handlerContent = readFile(HANDLERS_FILE) || '';
  
  for (const entry of entries) {
    const gamePath = join(GAMES_DIR, entry);
    if (statSync(gamePath).isDirectory()) {
      const logicFile = join(gamePath, 'logic.go');
      if (existsSync(logicFile)) {
        // Auto-detect handler name
        const handlerName = findHandlerName(entry, handlerContent);
        
        // Auto-detect web API file
        const webApiFile = findWebApiFile(entry);
        
        games.push({
          name: entry,
          logicFile: logicFile,
          handlerName: handlerName,
          webApiFile: webApiFile
        });
      }
    }
  }
  
  return games;
}

/**
 * Read game code files
 */
function readGameCode(game) {
  const code = {
    goLogic: readFile(game.logicFile),
    goHandler: null,
    webApi: null
  };

  // Read handler file (check for game-specific handler)
  const handlerContent = readFile(HANDLERS_FILE);
  if (handlerContent && game.handlerName) {
    // Extract relevant handler section
    const handlerRegex = new RegExp(
      `(type\\s+${game.handlerName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}[\\s\\S]*?)(?=type\\s+\\w+Handler|type\\s+\\w+Handler|$)`,
      'i'
    );
    const match = handlerContent.match(handlerRegex);
    if (match) {
      // Try to get just the handler methods, not the entire section
      const handlerMethodsRegex = new RegExp(
        `(type\\s+${game.handlerName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}[\\s\\S]*?func\\s+\\(h\\s+\\*${game.handlerName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\)[\\s\\S]*?)(?=type\\s+\\w+Handler|func\\s+\\(h\\s+\\*\\w+Handler\\)|$)`,
        'i'
      );
      const methodsMatch = handlerContent.match(handlerMethodsRegex);
      if (methodsMatch) {
        code.goHandler = methodsMatch[1];
      } else {
        code.goHandler = match[1].substring(0, 8000); // Limit size
      }
    } else {
      // Fallback: include entire handlers file if we can't extract
      code.goHandler = handlerContent.substring(0, 8000);
    }
  }

  // Read web API file if it exists
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
      if (file.endsWith('.md') && file !== 'Home.md' && file !== '_Sidebar.md' && file !== '_Footer.md') {
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
    'facts': 'Facts'
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
      codeContext.push(`=== Go Logic (package: ${packageName}) ===\n${code.goLogic.substring(0, 8000)}`);
    }
    if (code.goHandler) {
      codeContext.push(`=== Go Handler (handlers.go) ===\n${code.goHandler.substring(0, 8000)}`);
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
1. Analyze the Go logic file to understand the game mechanics
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
          const retryAfter = response.headers.get('retry-after');
          const waitTime = retryAfter 
            ? parseInt(retryAfter) * 1000 
            : Math.min(1000 * Math.pow(2, attempt), 30000);
          
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Rate limited (429) on ${model}. Waiting ${waitTime / 1000}s...`);
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
 * Main function
 */
async function main() {
  console.log('🚀 Starting AI-powered documentation update...\n');

  if (!existsSync(WIKI_ROOT)) {
    console.error(`❌ Wiki directory not found: ${WIKI_ROOT}`);
    console.error('   Set WIKI_ROOT environment variable or ensure wiki is cloned as sibling directory');
    process.exit(1);
  }

  const games = getGameDirectories();
  console.log(`✓ Found ${games.length} games to process\n`);

  let filesUpdated = 0;
  let filesGenerated = 0;
  let filesSkipped = 0;

  for (const game of games) {
    console.log(`\n📋 Processing ${game.name}...`);

    // Read code files
    const code = readGameCode(game);
    if (!code.goLogic) {
      console.warn(`  ⚠️  No logic file found, skipping`);
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

    // Rate limiting: wait between API calls
    await sleep(3000);
  }

  console.log('\n' + '='.repeat(50));
  console.log(`✅ Documentation update complete!`);
  console.log(`   Files generated: ${filesGenerated}`);
  console.log(`   Files updated: ${filesUpdated}`);
  console.log(`   Files skipped: ${filesSkipped}`);
  console.log('='.repeat(50));
}

// Run the script
main().catch(error => {
  console.error('❌ Fatal error:', error);
  process.exit(1);
});
