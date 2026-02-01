/**
 * Translation provider registry and factory
 * 
 * Manages available translation providers and provides a unified interface
 * Automatically checks provider availability and selects the best available option
 */

import * as groqProvider from './groq.js';
import * as geminiProvider from './gemini.js';

/**
 * Available providers registry
 */
const PROVIDERS = {
  groq: groqProvider,
  gemini: geminiProvider,
};

/**
 * Provider availability cache
 */
let availabilityCache = null;

/**
 * Check availability of all providers and cache results
 * @returns {Promise<Object>} Object mapping provider names to availability info
 */
export async function checkAllProviders() {
  if (availabilityCache !== null) {
    return availabilityCache;
  }

  console.log('🔍 Checking provider availability...\n');
  
  const results = {};
  
  // Check Groq
  try {
    const groqResult = await groqProvider.checkAvailability();
    results.groq = groqResult;
    if (groqResult.available) {
      console.log(`✓ Groq: Available (${groqResult.models.length} models)`);
      if (groqResult.totalAvailable) {
        console.log(`  Total models available: ${groqResult.totalAvailable}`);
      }
    } else {
      console.log(`✗ Groq: Unavailable - ${groqResult.error}`);
    }
  } catch (error) {
    results.groq = { available: false, models: [], error: error.message };
    console.log(`✗ Groq: Error - ${error.message}`);
  }

  // Check Gemini
  try {
    const geminiResult = await geminiProvider.checkAvailability();
    results.gemini = geminiResult;
    if (geminiResult.available) {
      console.log(`✓ Gemini: Available (${geminiResult.models.length} models)`);
      if (geminiResult.totalAvailable) {
        console.log(`  Total models available: ${geminiResult.totalAvailable}`);
      }
    } else {
      console.log(`✗ Gemini: Unavailable - ${geminiResult.error}`);
    }
  } catch (error) {
    results.gemini = { available: false, models: [], error: error.message };
    console.log(`✗ Gemini: Error - ${error.message}`);
  }

  console.log('');
  availabilityCache = results;
  return results;
}

/**
 * Get the first available provider (preference: Groq > Gemini)
 * @returns {Promise<Object>} Available provider module or null if none available
 */
export async function getAvailableProvider() {
  const results = await checkAllProviders();
  
  // Check if a specific provider is requested via environment variable
  const requestedProvider = process.env.TRANSLATION_PROVIDER;
  if (requestedProvider && PROVIDERS[requestedProvider]) {
    const requestedResult = results[requestedProvider];
    if (requestedResult?.available) {
      console.log(`✓ Using requested provider: ${requestedProvider}\n`);
      return PROVIDERS[requestedProvider];
    } else {
      console.warn(`⚠️  Requested provider '${requestedProvider}' is not available: ${requestedResult?.error || 'Unknown error'}`);
      console.warn(`   Falling back to first available provider...\n`);
    }
  }

  // Try providers in preference order
  const preferenceOrder = ['groq', 'gemini'];
  
  for (const providerName of preferenceOrder) {
    const result = results[providerName];
    if (result?.available) {
      console.log(`✓ Using provider: ${providerName}\n`);
      return PROVIDERS[providerName];
    }
  }

  return null;
}

/**
 * Get a translation provider by name (legacy function, use getAvailableProvider instead)
 * @param {string} providerName - Name of the provider ('groq' or 'gemini')
 * @returns {Object} Provider module with translateBatch, getBatchDelay, getProviderName
 * @throws {Error} If provider is not found
 */
export function getProvider(providerName) {
  if (!PROVIDERS[providerName]) {
    const available = Object.keys(PROVIDERS).join(', ');
    throw new Error(`Unknown provider: ${providerName}. Available providers: ${available}`);
  }

  return PROVIDERS[providerName];
}

/**
 * Get the default provider (legacy function, use getAvailableProvider instead)
 * @returns {Object} Default provider module
 * @deprecated Use getAvailableProvider() instead
 */
export function getDefaultProvider() {
  const requestedProvider = process.env.TRANSLATION_PROVIDER || 'groq';
  return getProvider(requestedProvider);
}

/**
 * List all available providers
 * @returns {string[]} Array of provider names
 */
export function listProviders() {
  return Object.keys(PROVIDERS);
}

/**
 * Check if a provider is available
 * @param {string} providerName - Name of the provider
 * @returns {boolean} True if provider exists
 */
export function hasProvider(providerName) {
  return providerName in PROVIDERS;
}
