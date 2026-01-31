#!/usr/bin/env node

/**
 * Translation script for automatically translating missing keys
 * 
 * This script:
 * 1. Reads en.json (source of truth)
 * 2. Reads all other locale files
 * 3. Finds missing keys per language
 * 4. Calls Groq API to translate missing keys (with exponential backoff retry)
 * 5. Updates locale files with new translations
 * 
 * Rate Limiting & Model Fallback:
 * - Uses 3-second delay between requests (20 req/min, well within free tier 30 req/min limit)
 * - Implements exponential backoff for 429 rate limit errors
 * - Automatic model fallback: tries multiple models if one hits rate limits
 *   - Primary: llama-3.3-70b-versatile (best quality)
 *   - Fallback 1: llama-3.1-70b-versatile (similar quality, different rate limit pool)
 *   - Fallback 2: llama-3.1-8b-instant (faster, often higher rate limits)
 *   - Fallback 3: mixtral-8x7b-32768 (alternative model)
 * - Retries each model up to 2 times before trying next model
 * - Respects Retry-After header if provided by API
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
// Different models may have different rate limit pools, so this increases success rate
const GROQ_MODELS = [
  'llama-3.3-70b-versatile',  // Primary: best quality
  'llama-3.1-70b-versatile',  // Fallback 1: similar quality, different rate limit pool
  'llama-3.1-8b-instant',     // Fallback 2: faster, often higher rate limits
  'mixtral-8x7b-32768'        // Fallback 3: alternative model
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
 * Find missing keys in a locale file compared to English
 */
function findMissingKeys(enKeys, localeData) {
  const missing = {};
  for (const key of Object.keys(enKeys)) {
    if (!(key in localeData) || !localeData[key] || localeData[key].trim() === '') {
      missing[key] = enKeys[key];
    }
  }
  return missing;
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Translate missing keys using Groq API with model fallback and exponential backoff retry
 */
async function translateKeys(missingKeys, targetLanguage) {
  if (Object.keys(missingKeys).length === 0) {
    return {};
  }

  console.log(`  📝 Translating ${Object.keys(missingKeys).length} keys to ${targetLanguage}...`);

  // Prepare the translation request
  // Format: key-value pairs as JSON
  const keysToTranslate = Object.entries(missingKeys)
    .map(([key, value]) => `"${key}": "${value}"`)
    .join(',\n    ');

  const prompt = `You are a professional translator. Translate the following JSON key-value pairs from English to ${targetLanguage}.

IMPORTANT INSTRUCTIONS:
1. Translate ONLY the values, keep the keys exactly as they are
2. Preserve all template variables like {{.Variable}} and {{Variable}} exactly as they appear
3. Preserve all printf-style placeholders like %d, %s, %.0f%% exactly as they appear
4. Preserve emojis and special characters
5. Return ONLY valid JSON, no explanations or markdown formatting
6. Maintain the same structure and formatting

JSON to translate:
{
    ${keysToTranslate}
}

Return the translated JSON:`;

  // Try each model in the fallback chain
  for (let modelIndex = 0; modelIndex < GROQ_MODELS.length; modelIndex++) {
    const model = GROQ_MODELS[modelIndex];
    const isLastModel = modelIndex === GROQ_MODELS.length - 1;
    const retriesPerModel = 2; // Retry each model 2 times before moving to next

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

        // Handle rate limiting (429) - try next model or retry with backoff
        if (response.status === 429) {
          const retryAfter = response.headers.get('retry-after');
          // Custom wait times: attempt 1 = 1s, attempt 2 = 5s, then +5s per additional attempt
          const waitTime = retryAfter 
            ? parseInt(retryAfter) * 1000 
            : (attempt === 0 ? 1000 : attempt * 5000); // 1s, 5s, 10s, 15s, ...
          
          // If not last attempt on this model, retry with backoff
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Rate limited (429) on ${model}. Waiting ${waitTime / 1000}s before retry ${attempt + 1}/${retriesPerModel}...`);
            await sleep(waitTime);
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
            const waitTime = Math.min(1000 * Math.pow(2, attempt), 5000);
            console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${response.status}. Retrying in ${waitTime / 1000}s...`);
            await sleep(waitTime);
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
          // Try next model if available
          if (!isLastModel) {
            console.warn(`  ⚠️  Trying next model...`);
            await sleep(2000);
            break;
          }
          return {};
        }
        
        // Validate that all keys are present
        const missingInResponse = [];
        for (const key of Object.keys(missingKeys)) {
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
        
        // For other errors, wait and retry or try next model
        if (attempt < retriesPerModel - 1) {
          const waitTime = Math.min(1000 * Math.pow(2, attempt), 10000); // Max 10 seconds
          console.warn(`  ⚠️  Error on ${model} (attempt ${attempt + 1}/${retriesPerModel}): ${error.message}. Retrying in ${waitTime / 1000}s...`);
          await sleep(waitTime);
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

    // Find missing keys
    const missingKeys = findMissingKeys(enData, localeData);

    if (Object.keys(missingKeys).length === 0) {
      console.log(`✓ ${locale}.json - No missing keys`);
      continue;
    }

    console.log(`\n📋 ${locale}.json - Found ${Object.keys(missingKeys).length} missing keys`);

    // Translate missing keys
    const translations = await translateKeys(missingKeys, LOCALE_TO_LANGUAGE[locale]);

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
    // Increased delay to 3 seconds to stay well within free tier limits (30 req/min)
    // This gives us ~20 requests per minute, leaving buffer for retries
    await sleep(3000);
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
