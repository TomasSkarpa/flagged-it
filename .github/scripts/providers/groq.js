/**
 * Groq API translation provider
 * 
 * Features:
 * - Multiple model fallback chain
 * - Rate limit handling with retries
 * - Exponential backoff retry logic
 * - Model-specific error handling
 */

import { buildTranslationPrompt, parseTranslationResponse } from '../translate.js';

const GROQ_API_URL = 'https://api.groq.com/openai/v1/chat/completions';
const GROQ_MODELS_URL = 'https://api.groq.com/openai/v1/models';
const GROQ_API_KEY = process.env.GROQ_API_KEY;

// Preferred model order (will be filtered by availability)
const PREFERRED_MODELS = [
  'llama-3.3-70b-versatile',              // Primary: best quality (30 RPM, 12k TPM)
  'meta-llama/llama-4-scout-17b-16e-instruct', // Fallback 1: next-gen performance (30 RPM, 30k TPM)
  'qwen/qwen3-32b',                       // Fallback 2: excellent for coding/math (60 RPM, 6k TPM)
  'llama-3.1-8b-instant'                  // Fallback 3: pure speed (30 RPM, 6k TPM)
];

// Discovered available models (will be populated by checkAvailability)
let availableModels = [];

/**
 * Sleep for specified milliseconds
 * @param {number} ms - Milliseconds to sleep
 * @returns {Promise<void>}
 */
function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Check if Groq provider is available (API key exists and API is accessible)
 * @returns {Promise<{available: boolean, models: string[], error?: string}>}
 */
export async function checkAvailability() {
  if (!GROQ_API_KEY) {
    return { available: false, models: [], error: 'GROQ_API_KEY environment variable is not set' };
  }

  try {
    const response = await fetch(GROQ_MODELS_URL, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${GROQ_API_KEY}`,
      },
    });

    if (!response.ok) {
      return { 
        available: false, 
        models: [], 
        error: `Groq API returned ${response.status}: ${response.statusText}` 
      };
    }

    const data = await response.json();
    const allModels = data.data?.map(model => model.id) || [];
    
    // Filter to preferred models that are available, maintaining order
    const filteredModels = PREFERRED_MODELS.filter(model => allModels.includes(model));
    
    // If no preferred models are available, use all available models
    // If no models exist at all, provider is not available
    let models;
    if (filteredModels.length > 0) {
      models = filteredModels;
    } else if (allModels.length > 0) {
      models = allModels.slice(0, 4);
    } else {
      // No models available - provider cannot be used
      return { 
        available: false, 
        models: [], 
        error: 'No Groq models are available' 
      };
    }
    
    availableModels = models;
    
    return { 
      available: true, 
      models: models,
      totalAvailable: allModels.length
    };
  } catch (error) {
    return { 
      available: false, 
      models: [], 
      error: `Failed to check Groq availability: ${error.message}` 
    };
  }
}

/**
 * Get available models for this provider
 * @returns {string[]} Array of available model names
 */
export function getAvailableModels() {
  return availableModels.length > 0 ? availableModels : PREFERRED_MODELS;
}

/**
 * Translate a single batch of keys using Groq API with model fallback and retry logic
 * @param {Object} keyBatch - Object with keys to translate
 * @param {Object} contextBatch - Object with context hints for keys
 * @param {string} targetLanguage - Target language name (e.g., "Spanish")
 * @param {string} localeCode - Locale code (e.g., "es")
 * @param {Object} narrativeInstruction - Narrative instruction for locale (optional)
 * @param {number} batchNumber - Current batch number (for logging)
 * @param {number} totalBatches - Total number of batches (for logging)
 * @returns {Promise<Object>} Translated keys object
 */
export async function translateBatch(keyBatch, contextBatch, targetLanguage, localeCode, narrativeInstruction, batchNumber, totalBatches) {
  if (Object.keys(keyBatch).length === 0) {
    return {};
  }

  if (!GROQ_API_KEY) {
    throw new Error('GROQ_API_KEY environment variable is not set');
  }

  const batchInfo = totalBatches > 1 ? ` (batch ${batchNumber}/${totalBatches})` : '';
  console.log(`  📝 Translating ${Object.keys(keyBatch).length} keys${batchInfo}...`);

  const prompt = buildTranslationPrompt(keyBatch, contextBatch, targetLanguage, narrativeInstruction);
  const systemMessage = 'You are a professional translator. Return only valid JSON without any markdown formatting or explanations.';

  // Get available models (use discovered models if available, otherwise fallback to preferred)
  const modelsToTry = availableModels.length > 0 ? availableModels : PREFERRED_MODELS;

  // Try each model in the fallback chain
  for (let modelIndex = 0; modelIndex < modelsToTry.length; modelIndex++) {
    const model = modelsToTry[modelIndex];
    const isLastModel = modelIndex === modelsToTry.length - 1;
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
                content: systemMessage
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
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Rate limited (429) on ${model}. Waiting ${waitBetweenAttempts / 1000}s before retry ${attempt + 1}/${retriesPerModel}...`);
            await sleep(waitBetweenAttempts);
            continue;
          } else {
            console.warn(`  ⚠️  Rate limited (429) on ${model} after ${retriesPerModel} attempts. Trying next model...`);
            await sleep(2000); // Brief pause before switching models
            break; // Break inner loop to try next model
          }
        }

        if (!response.ok) {
          const errorText = await response.text();
          if (attempt === retriesPerModel - 1 && isLastModel) {
            throw new Error(`Groq API error: ${response.status} ${response.statusText} - ${errorText}`);
          }
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

        // Parse the translated JSON
        const translated = parseTranslationResponse(translatedText);

        if (!translated) {
          console.error(`  ❌ Failed to parse translation response from ${model}`);
          console.error(`  Response preview: ${translatedText.substring(0, 200)}...`);
          if (attempt < retriesPerModel - 1) {
            console.warn(`  ⚠️  Parse error on ${model} (attempt ${attempt + 1}/${retriesPerModel}). Retrying in ${waitBetweenAttempts / 1000}s...`);
            await sleep(waitBetweenAttempts);
            continue;
          }
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
 * Get delay between batches for rate limiting
 * @returns {number} Delay in milliseconds
 */
export function getBatchDelay() {
  return 2500; // 2.5s delay = 24 RPM, safely under 30 RPM limit
}

/**
 * Get provider name
 * @returns {string} Provider name
 */
export function getProviderName() {
  return 'Groq';
}
