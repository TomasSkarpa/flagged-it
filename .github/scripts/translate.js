#!/usr/bin/env node

/**
 * Translation script for automatically translating missing keys
 * 
 * This script:
 * 1. Reads en.json (source of truth)
 * 2. Reads all other locale files
 * 3. Finds missing keys per language
 * 4. Calls Groq API to translate missing keys
 * 5. Updates locale files with new translations
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
const GROQ_MODEL = 'llama-3.3-70b-versatile';
const GROQ_API_KEY = process.env.GROQ_API_KEY;

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
 * Translate missing keys using Groq API
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

  try {
    const response = await fetch(GROQ_API_URL, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${GROQ_API_KEY}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model: GROQ_MODEL,
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

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Groq API error: ${response.status} ${response.statusText} - ${errorText}`);
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
      console.error(`  ❌ Failed to parse translation response: ${parseError.message}`);
      console.error(`  Response preview: ${cleanedText.substring(0, 200)}...`);
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
      console.warn(`  ⚠️  Warning: ${missingInResponse.length} keys missing in translation response`);
      // Still return what we got - partial translation is better than nothing
    }

    return translated;
  } catch (error) {
    console.error(`  ❌ Error translating to ${targetLanguage}:`, error.message);
    // Return empty object on error - we'll skip this language
    return {};
  }
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

    // Rate limiting: wait a bit between API calls to respect rate limits
    await new Promise(resolve => setTimeout(resolve, 1000));
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
