#!/usr/bin/env node

/**
 * Translation script for automatically translating missing keys
 * 
 * This script:
 * 1. Reads en.json (source of truth)
 * 2. Reads all other locale files
 * 3. Finds missing keys per language
 * 4. Calls translation provider API (Groq/Gemini) to translate missing keys in batches (max 30 keys per request)
 * 5. Updates locale files with new translations
 * 
 * Provider Selection:
 * - Default provider: Gemini (can be changed via TRANSLATION_PROVIDER environment variable)
 * - Available providers: 'gemini', 'groq'
 * - Set TRANSLATION_PROVIDER=groq to use Groq API instead
 * 
 * Batching:
 * - Maximum 30 translations per API request to avoid token limits
 * - Large translation sets are automatically split into batches
 * - Provider-specific delays between batches to respect rate limits
 * 
 * Rate Limiting & Model Fallback:
 * - Provider-specific retry logic and model fallback chains
 * - Retries each model up to 5 times with 5-second wait between attempts
 * - Automatic model fallback: tries next model if current model fails after 5 attempts
 * - Waits 5 seconds between retry attempts on the same model
 * - Switches to next model after 5 failed attempts
 */

import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { getAvailableProvider } from './providers/index.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine project root (two levels up from .github/scripts/)
const PROJECT_ROOT = join(__dirname, '../..');

// Path to translation files
const LOCALES_DIR = join(PROJECT_ROOT, 'web/src/lib/translations/locales');
const EN_FILE = join(LOCALES_DIR, 'en.json');

// Locale code to language name mapping for translation providers
const LOCALE_TO_LANGUAGE = {
  'es': 'Spanish',
  'fr': 'French',
  'de': 'German',
  'nl': 'Dutch',
  'nb': 'Norwegian Bokmål',
  'da': 'Danish',
  'sv': 'Swedish',
  'fi': 'Finnish',
  'pt': 'Portuguese',
  'tr': 'Turkish',
  'ro': 'Romanian',
  'hu': 'Hungarian',
  'hr': 'Croatian',
  'cs': 'Czech',
  'sk': 'Slovak',
  'pl': 'Polish',
  'it': 'Italian',
  'id': 'Indonesian',
  'ms': 'Malay',
  'fil': 'Filipino',
  'sw': 'Swahili',
  'vi': 'Vietnamese',
  'ru': 'Russian',
  'zh': 'Chinese (Simplified)',
  'ko': 'Korean',
  'ja': 'Japanese',
  'ar': 'Arabic',
  'hi': 'Hindi',
  'th': 'Thai',
  'uk': 'Ukrainian',
  'he': 'Hebrew',
  'el': 'Greek'
};

// Locale code to narrative instruction mapping for translation tone/style
const LOCALE_TO_NARRATIVE = {
  // Informal Group (T-form)
  'cs': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'sk': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'hr': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'pl': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'ru': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'uk': 'Use informal singular (T-form). Avoid formal plural unless addressing multiple people.',
  'de': 'Use informal (Du-form). Avoid the formal Sie.',
  'nl': 'Use informal (Du-form). Avoid the formal Sie.',
  'fr': 'Use informal (Tu). Avoid the formal Vous.',
  'it': 'Use informal (Tu). Avoid the formal Lei.',
  'es': 'Use informal (Tú). Avoid the formal Usted.',
  'pt': 'Use informal (Tu). Avoid the formal Você.',
  'da': 'Use informal. (Formal address is rare in modern digital contexts here).',
  'sv': 'Use informal. (Formal address is rare in modern digital contexts here).',
  'nb': 'Use informal. (Formal address is rare in modern digital contexts here).',
  'fi': 'Use informal. (Formal address is rare in modern digital contexts here).',
  'hu': 'Use informal singular.',
  'tr': 'Use informal singular.',
  'ro': 'Use informal singular.',
  // Neutral/Polite Group
  'en': 'Use active, imperative verbs. Friendly and direct.',
  'ja': 'Use polite neutral (e.g., Desu/Masu in Japanese). Avoid "dictionary form" (too blunt) or "honorifics" (too stiff).',
  'ko': 'Use polite neutral (Haeyo-che style). Avoid the very formal Hasipsio-che or blunt Panmal.',
  'zh': 'Use a friendly, standard neutral tone. Avoid overly formal officialese.',
  'ar': 'Use Modern Standard Arabic (MSA). Use the masculine singular imperative as the neutral default.',
  'hi': 'Use friendly, polite imperative. Use standard polite particles common in apps.',
  'th': 'Use friendly, polite imperative. Use standard polite particles common in apps.',
  'vi': 'Use friendly, polite imperative. Use standard polite particles common in apps.',
  'id': 'Use standard friendly imperative. Focus on clarity and engagement.',
  'ms': 'Use standard friendly imperative. Focus on clarity and engagement.',
  'fil': 'Use standard friendly imperative. Focus on clarity and engagement.',
  'sw': 'Use standard friendly imperative. Focus on clarity and engagement.',
  // Greek — orthography and grammar rules for UI strings (do not rely on English title case)
  'el':
    'Modern Greek (Ελληνικά): Use polite plural or neutral app tone consistent with the English source (match formality of the English string). ' +
    'CAPITALIZATION: Use sentence-style capitalization only. Capitalize the first word of each complete sentence or standalone headline; do NOT use English-style title case—keep common nouns mid-sentence in lowercase (e.g. «Μαντεύστε την πρωτεύουσα», not «...την Πρωτεύουσα»). ' +
    'GRAMMAR: Enforce correct gender, number, and case agreement—articles (ο/η/το, του/της/του, etc.) and adjectives must agree with their nouns (masculine/feminine/neuter). Examples: neuter «σχήμα» takes «το σχήμα», never «την σχήμα»; feminine «χώρα» takes «η χώρα» / «τη χώρα» as context requires. Proofread every noun phrase. ' +
    'Leave true proper names (official country/city names, brand «Flagged It») as appropriate for Greek; do not introduce English fragments except established loanwords where natural.'
};

// Translation provider (will be initialized in main function)
let translationProvider = null;

/**
 * Read and parse JSON file
 */
function readJSON(filePath) {
  try {
    const content = readFileSync(filePath, 'utf-8');
    return JSON.parse(content);
  } catch (error) {
    console.error(`❌ Error reading ${filePath}:`, error.message);
    return null;
  }
}

/**
 * Write JSON file with proper formatting
 */
function writeJSON(filePath, data) {
  try {
    const content = JSON.stringify(data, null, 2) + '\n';
    writeFileSync(filePath, content, 'utf-8');
    return true;
  } catch (error) {
    console.error(`❌ Error writing ${filePath}:`, error.message);
    return false;
  }
}

/**
 * Extract value from translation entry (handles both string and object formats)
 */
function getTranslationValue(entry) {
  if (typeof entry === 'string') {
    return entry;
  }
  if (typeof entry === 'object' && entry !== null && 'value' in entry) {
    return entry.value;
  }
  return entry;
}

/**
 * Extract context hint from translation entry (if available)
 */
function getTranslationContext(entry) {
  if (typeof entry === 'object' && entry !== null && '_context' in entry) {
    return entry._context;
  }
  return null;
}

/**
 * Find missing keys in a locale file compared to English
 * Returns object with keys mapped to { value, context } structure
 */
function findMissingKeys(enKeys, localeData) {
  const missing = {};
  const missingWithContext = {};
  
  for (const key of Object.keys(enKeys)) {
    const enValue = getTranslationValue(enKeys[key]);
    const enContext = getTranslationContext(enKeys[key]);
    
    // Check if key is missing or empty in locale file
    const localeValue = localeData[key];
    const localeTranslatedValue = getTranslationValue(localeValue);
    
    if (!(key in localeData) || !localeTranslatedValue || localeTranslatedValue.trim() === '') {
      missing[key] = enValue;
      if (enContext) {
        missingWithContext[key] = enContext;
      }
    }
  }
  
  return { missing, contexts: missingWithContext };
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Split an object into chunks of specified size
 */
function chunkObject(obj, chunkSize) {
  const entries = Object.entries(obj);
  const chunks = [];
  for (let i = 0; i < entries.length; i += chunkSize) {
    chunks.push(Object.fromEntries(entries.slice(i, i + chunkSize)));
  }
  return chunks;
}

/**
 * Build translation prompt with context hints and narrative instructions
 * @param {Object} keyBatch - Object with keys to translate
 * @param {Object} contextBatch - Object with context hints for keys
 * @param {string} targetLanguage - Target language name (e.g., "Spanish")
 * @param {string|null} narrativeInstruction - Narrative instruction for locale (optional)
 * @returns {string} Formatted translation prompt
 */
export function buildTranslationPrompt(keyBatch, contextBatch, targetLanguage, narrativeInstruction = null) {
  // Prepare the translation request with context hints
  const keysToTranslate = Object.entries(keyBatch)
    .map(([key, value]) => {
      const context = contextBatch[key];
      if (context) {
        return `"${key}": "${value}" // Context: ${context}`;
      }
      return `"${key}": "${value}"`;
    })
    .join(',\n    ');

  const contextInstructions = Object.keys(contextBatch).length > 0
    ? `\n\nTRANSLATION CONTEXT HINTS (use these to guide your translation):
${Object.entries(contextBatch).map(([key, context]) => `- "${key}": ${context}`).join('\n')}`
    : '';

  const narrativeSection = narrativeInstruction
    ? `\n\nNARRATIVE STYLE INSTRUCTION:\n${narrativeInstruction}`
    : '';

  const prompt = `You are a professional translator. Translate the following JSON key-value pairs from English to ${targetLanguage}.${contextInstructions}${narrativeSection}

IMPORTANT INSTRUCTIONS:
1. Translate ONLY the values, keep the keys exactly as they are
2. Preserve all template variables like {{.Variable}} and {{Variable}} exactly as they appear
3. Preserve all printf-style placeholders like %d, %s, %.0f%% exactly as they appear
4. Preserve emojis and special characters
5. Return ONLY valid JSON, no explanations or markdown formatting
6. Maintain the same structure and formatting
7. Use the context hints (if provided) to guide your translation style and meaning${narrativeSection ? '\n8. Follow the narrative style instruction above for the appropriate tone and formality level' : ''}

JSON to translate:
{
    ${keysToTranslate}
}

Return the translated JSON:`;

  return prompt;
}

/**
 * Clean and parse JSON response from API
 * Removes markdown code blocks if present
 * @param {string} text - Raw response text from API
 * @returns {Object|null} Parsed JSON object or null if parsing fails
 */
export function parseTranslationResponse(text) {
  if (!text) {
    return null;
  }

  let cleanedText = text.trim();
  
  // Remove markdown code blocks if present
  if (cleanedText.startsWith('```json')) {
    cleanedText = cleanedText.replace(/^```json\n?/, '').replace(/\n?```$/, '');
  } else if (cleanedText.startsWith('```')) {
    cleanedText = cleanedText.replace(/^```\n?/, '').replace(/\n?```$/, '');
  }

  try {
    return JSON.parse(cleanedText);
  } catch (error) {
    return null;
  }
}

/**
 * Translate a single batch of keys using the configured translation provider
 * @param {Object} keyBatch - Object with keys to translate
 * @param {Object} contextBatch - Object with context hints for keys
 * @param {string} targetLanguage - Target language name (e.g., "Spanish")
 * @param {string} localeCode - Locale code (e.g., "es")
 * @param {number} batchNumber - Current batch number (for logging)
 * @param {number} totalBatches - Total number of batches (for logging)
 * @returns {Promise<Object>} Translated keys object
 */
async function translateKeyBatch(keyBatch, contextBatch, targetLanguage, localeCode, batchNumber, totalBatches) {
  if (Object.keys(keyBatch).length === 0) {
    return {};
  }

  // Get narrative instruction for this locale (if available)
  const narrativeInstruction = LOCALE_TO_NARRATIVE[localeCode] || null;

  // Use provider's translateBatch function
  return await translationProvider.translateBatch(
    keyBatch,
    contextBatch,
    targetLanguage,
    localeCode,
    narrativeInstruction,
    batchNumber,
    totalBatches
  );
}

/**
 * Translate missing keys using the configured translation provider with batching (max 30 keys per request)
 * Uses provider-specific model fallback and exponential backoff retry
 * @param {Object} missingKeys - Object with missing keys to translate
 * @param {Object} contexts - Object with context hints for keys
 * @param {string} targetLanguage - Target language name (e.g., "Spanish")
 * @param {string} localeCode - Locale code (e.g., "es")
 * @returns {Promise<Object>} Translated keys object
 */
async function translateKeys(missingKeys, contexts, targetLanguage, localeCode) {
  if (Object.keys(missingKeys).length === 0) {
    return {};
  }

  const MAX_KEYS_PER_REQUEST = 30;
  const totalKeys = Object.keys(missingKeys).length;
  
  console.log(`  📝 Translating ${totalKeys} keys to ${targetLanguage}...`);

  // Split keys into chunks of MAX_KEYS_PER_REQUEST
  const keyChunks = chunkObject(missingKeys, MAX_KEYS_PER_REQUEST);
  const contextChunks = keyChunks.map(chunk => {
    const chunkContexts = {};
    for (const key of Object.keys(chunk)) {
      if (contexts[key]) {
        chunkContexts[key] = contexts[key];
      }
    }
    return chunkContexts;
  });

  const totalBatches = keyChunks.length;
  const allTranslations = {};

  // Get provider-specific batch delay
  const batchDelay = translationProvider.getBatchDelay();

  // Process each batch sequentially
  for (let i = 0; i < keyChunks.length; i++) {
    const keyBatch = keyChunks[i];
    const contextBatch = contextChunks[i];
    
    const batchTranslations = await translateKeyBatch(
      keyBatch,
      contextBatch,
      targetLanguage,
      localeCode,
      i + 1,
      totalBatches
    );

    // Merge batch translations into result
    Object.assign(allTranslations, batchTranslations);

    // Rate limiting: wait between batches (except for the last batch)
    if (i < keyChunks.length - 1) {
      await sleep(batchDelay);
    }
  }

  return allTranslations;
}

/**
 * Main function
 */
async function main() {
  console.log('🚀 Starting translation process...\n');

  // Initialize translation provider (checks availability and selects best option)
  translationProvider = await getAvailableProvider();
  
  if (!translationProvider) {
    console.error('❌ No translation providers available. Please check your API keys:');
    console.error('   - GROQ_API_KEY for Groq');
    console.error('   - GEMINI_API_KEY for Gemini');
    process.exit(1);
  }

  // Read English file (source of truth)
  const enData = readJSON(EN_FILE);
  if (!enData) {
    console.error('❌ Failed to read en.json');
    process.exit(1);
  }

  console.log(`✓ Loaded en.json with ${Object.keys(enData).length} keys\n`);

  // Get all locale files
  const localeFiles = Object.keys(LOCALE_TO_LANGUAGE);
  let totalTranslated = 0;
  let filesUpdated = 0;

  for (const locale of localeFiles) {
    const localeFile = join(LOCALES_DIR, `${locale}.json`);
    const localeData = readJSON(localeFile);

    if (!localeData) {
      console.warn(`⚠️  Skipping ${locale}.json (file not found or invalid)`);
      continue;
    }

    // Find missing keys (now returns { missing, contexts })
    const { missing: missingKeys, contexts } = findMissingKeys(enData, localeData);

    if (Object.keys(missingKeys).length === 0) {
      console.log(`✓ ${locale}.json - No missing keys`);
      continue;
    }

    console.log(`\n📋 ${locale}.json - Found ${Object.keys(missingKeys).length} missing keys`);

    // Translate missing keys (pass contexts for translation hints and locale code for narrative instructions)
    const translations = await translateKeys(missingKeys, contexts, LOCALE_TO_LANGUAGE[locale], locale);

    if (Object.keys(translations).length === 0) {
      console.warn(`  ⚠️  No translations received, skipping ${locale}.json`);
      continue;
    }

    // Merge translations into locale data
    const updatedData = { ...localeData };
    for (const [key, value] of Object.entries(translations)) {
      updatedData[key] = value;
    }

    // Write updated file
    if (writeJSON(localeFile, updatedData)) {
      console.log(`  ✓ Updated ${locale}.json with ${Object.keys(translations).length} translations`);
      filesUpdated++;
      totalTranslated += Object.keys(translations).length;
    }

    // Rate limiting: wait between API calls to respect rate limits
    // Use provider-specific delay
    await sleep(translationProvider.getBatchDelay());
  }

  console.log('\n' + '='.repeat(50));
  console.log(`✅ Translation complete!`);
  console.log(`   Files updated: ${filesUpdated}`);
  console.log(`   Total keys translated: ${totalTranslated}`);
  console.log('='.repeat(50));
}

// Run the script
main().catch(error => {
  console.error('❌ Fatal error:', error);
  process.exit(1);
});
