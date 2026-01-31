#!/usr/bin/env node

/**
 * Translation script for automatically translating missing keys
 * 
 * This script:
 * 1. Reads en.json (source of truth)
 * 2. Reads all other locale files
 * 3. Finds missing keys per language
 * 4. Calls Groq API to translate missing keys in batches (max 30 keys per request)
 * 5. Updates locale files with new translations
 * 
 * Batching:
 * - Maximum 30 translations per API request to avoid token limits
 * - Large translation sets are automatically split into batches
 * - 2.5s delay between batches to respect rate limits
 * 
 * Rate Limiting & Model Fallback:
 * - Uses 2.5-second delay between batches (24 req/min, safely under 30 req/min limit)
 * - Retries each model up to 5 times with 5-second wait between attempts
 * - Automatic model fallback: tries next model if current model fails after 5 attempts
 *   - Primary: llama-3.3-70b-versatile (best quality, 30 RPM)
 *   - Fallback 1: llama-4-scout-17b (next-gen performance, 30 RPM)
 *   - Fallback 2: qwen3-32b (excellent for coding/math, 60 RPM)
 *   - Fallback 3: llama-3.1-8b-instant (pure speed, 30 RPM)
 * - Waits 5 seconds between retry attempts on the same model
 * - Switches to next model after 5 failed attempts
 */

import { readFileSync, writeFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

// Determine project root (two levels up from .github/scripts/)
const PROJECT_ROOT = join(__dirname, '../..');

// Path to translation files
const LOCALES_DIR = join(PROJECT_ROOT, 'web/src/lib/translations/locales');
const EN_FILE = join(LOCALES_DIR, 'en.json');

// Locale code to language name mapping for Groq API
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
 * Translate a single batch of keys using Groq API with model fallback and exponential backoff retry
 */
async function translateKeyBatch(keyBatch, contextBatch, targetLanguage, batchNumber, totalBatches) {
  if (Object.keys(keyBatch).length === 0) {
    return {};
  }

  const batchInfo = totalBatches > 1 ? ` (batch ${batchNumber}/${totalBatches})` : '';
  console.log(`  📝 Translating ${Object.keys(keyBatch).length} keys${batchInfo}...`);

  // Prepare the translation request with context hints
  // Format: key-value pairs as JSON, with context hints included in the prompt
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

  const prompt = `You are a professional translator. Translate the following JSON key-value pairs from English to ${targetLanguage}.${contextInstructions}

IMPORTANT INSTRUCTIONS:
1. Translate ONLY the values, keep the keys exactly as they are
2. Preserve all template variables like {{.Variable}} and {{Variable}} exactly as they appear
3. Preserve all printf-style placeholders like %d, %s, %.0f%% exactly as they appear
4. Preserve emojis and special characters
5. Return ONLY valid JSON, no explanations or markdown formatting
6. Maintain the same structure and formatting
7. Use the context hints (if provided) to guide your translation style and meaning

JSON to translate:
{
    ${keysToTranslate}
}

Return the translated JSON:`;

  // Try each model in the fallback chain
  for (let modelIndex = 0; modelIndex < GROQ_MODELS.length; modelIndex++) {
    const model = GROQ_MODELS[modelIndex];
    const isLastModel = modelIndex === GROQ_MODELS.length - 1;
    const retriesPerModel = 5; // Retry each model 5 times before moving to next
    const waitBetweenAttempts = 5000; // Wait 5 seconds between attempts

    console.log(`  🔄 Trying model: ${model}${modelIndex > 0 ? ' (fallback)' : ''}`);

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
                content: 'You are a professional translator. Return only valid JSON without any markdown formatting or explanations.'
              },
              {
                role: 'user',
                content: prompt
              }
            ],
            temperature: 0.3,
            max_tokens: 4000,
          }),
        });

        // Handle rate limiting (429) - retry with 5s wait
        if (response.status === 429) {
          // If not last attempt on this model, retry after 5s
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Rate limited (429) on ${model}. Waiting ${waitBetweenAttempts / 1000}s before retry ${attempt + 1}/${retriesPerModel}...`);
            await sleep(waitBetweenAttempts);
            continue;
          } else {
            // Last attempt on this model failed - try next model
            console.warn(`  ⚠️  Rate limited (429) on ${model} after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000); // Brief pause before switching models
            break; // Break inner loop to try next model
          }
        }

        if (!response.ok) {
          const errorText = await response.text();
          // For non-429 errors, if it's the last attempt on last model, throw
          if (attempt === retriesPerModel - 1 && isLastModel) {
            throw new Error(`Groq API error: ${response.status} ${response.statusText} - ${errorText}`);
          }
          // Otherwise, try next attempt or next model
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${response.status}. Retrying in ${waitBetweenAttempts / 1000}s...`);
            await sleep(waitBetweenAttempts);
            continue;
          } else {
            console.warn(`  ⚠️  Error on ${model} after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000);
            break;
          }
        }

        const data = await response.json();
        const translatedText = data.choices[0]?.message?.content?.trim();

        if (!translatedText) {
          throw new Error('No translation received from API');
        }

        // Clean up the response - remove markdown code blocks if present
        let cleanedText = translatedText;
        if (cleanedText.startsWith('```json')) {
          cleanedText = cleanedText.replace(/^```json\n?/, '').replace(/\n?```$/, '');
        } else if (cleanedText.startsWith('```')) {
          cleanedText = cleanedText.replace(/^```\n?/, '').replace(/\n?```$/, '');
        }

        // Parse the translated JSON
        let translated;
        try {
          translated = JSON.parse(cleanedText);
        } catch (parseError) {
          console.error(`  ❌ Failed to parse translation response from ${model}: ${parseError.message}`);
          console.error(`  Response preview: ${cleanedText.substring(0, 200)}...`);
          // If not last attempt, retry after 5s
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Parse error on ${model} (attempt ${attempt + 1}/${retriesPerModel}). Retrying in ${waitBetweenAttempts / 1000}s...`);
            await sleep(waitBetweenAttempts);
            continue;
          }
          // Try next model if available
          if (!isLastModel) {
            console.warn(`  ⚠️  Parse error after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000);
            break;
          }
          return {};
        }
        
        // Validate that all keys are present
        const missingInResponse = [];
        for (const key of Object.keys(keyBatch)) {
          if (!(key in translated)) {
            missingInResponse.push(key);
          }
        }

        if (missingInResponse.length > 0) {
          console.warn(`  ⚠️  Warning: ${missingInResponse.length} keys missing in translation response from ${model}`);
          // Still return what we got - partial translation is better than nothing
        }

        console.log(`  ✅ Successfully translated using ${model}`);
        return translated;
      } catch (error) {
        // If it's the last attempt on the last model, give up
        if (attempt === retriesPerModel - 1 && isLastModel) {
          console.error(`  ❌ Error translating to ${targetLanguage} after trying all models:`, error.message);
          return {};
        }
        
        // For other errors, wait 5s and retry or try next model
        if (attempt < retriesPerModel - 1) {
          console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${error.message}. Retrying in ${waitBetweenAttempts / 1000}s...`);
          await sleep(waitBetweenAttempts);
        } else {
          console.warn(`  ⚠️  Error on ${model} after ${retriesPerModel} attempts: ${error.message}. Trying next model...`);
          await sleep(2000);
          break; // Try next model
        }
      }
    }
  }

  // If we get here, all models failed
  console.error(`  ❌ All models failed for ${targetLanguage}`);
  return {};
}

/**
 * Translate missing keys using Groq API with batching (max 30 keys per request)
 * Uses model fallback and exponential backoff retry
 */
async function translateKeys(missingKeys, contexts, targetLanguage) {
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

  // Process each batch sequentially
  for (let i = 0; i < keyChunks.length; i++) {
    const keyBatch = keyChunks[i];
    const contextBatch = contextChunks[i];
    
    const batchTranslations = await translateKeyBatch(
      keyBatch,
      contextBatch,
      targetLanguage,
      i + 1,
      totalBatches
    );

    // Merge batch translations into result
    Object.assign(allTranslations, batchTranslations);

    // Rate limiting: wait between batches (except for the last batch)
    if (i < keyChunks.length - 1) {
      await sleep(2500); // 2.5s delay between batches
    }
  }

  return allTranslations;
}

/**
 * Main function
 */
async function main() {
  console.log('🚀 Starting translation process...\n');

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

    // Translate missing keys (pass contexts for translation hints)
    const translations = await translateKeys(missingKeys, contexts, LOCALE_TO_LANGUAGE[locale]);

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
    // Minimum RPM is 30 (llama-3.3-70b, llama-4-scout, llama-3.1-8b)
    // Using 2.5s delay = 24 RPM, safely under 30 RPM limit to prevent blocking
    await sleep(2500);
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
