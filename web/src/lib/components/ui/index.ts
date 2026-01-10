// UI Components exports
export { default as LoadingSpinner } from './LoadingSpinner.svelte';
export { default as ScoreDisplay } from './ScoreDisplay.svelte';
export { default as ProgressBar } from './ProgressBar.svelte';
export { default as Keyboard } from './Keyboard.svelte';
export { default as RegionDropdown } from './RegionDropdown.svelte';

// Re-export types from Keyboard component
export type KeyboardLayout = import('./Keyboard.svelte').KeyboardLayout;
export type KeyState = import('./Keyboard.svelte').KeyState;
