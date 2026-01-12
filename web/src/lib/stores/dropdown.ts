import { writable } from 'svelte/store';

type DropdownType = 'language' | 'theme' | null;

/**
 * Store to manage which dropdown is currently open.
 * Only one dropdown can be open at a time.
 */
export const activeDropdown = writable<DropdownType>(null);

/**
 * Open a dropdown and close any other open dropdown
 */
export function openDropdown(type: DropdownType) {
	activeDropdown.set(type);
}

/**
 * Close the currently open dropdown
 */
export function closeDropdown() {
	activeDropdown.set(null);
}

/**
 * Toggle a dropdown - if it's already open, close it; otherwise open it and close others
 */
export function toggleDropdown(type: DropdownType) {
	activeDropdown.update(current => current === type ? null : type);
}
