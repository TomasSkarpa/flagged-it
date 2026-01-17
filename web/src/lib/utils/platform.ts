/**
 * Platform detection utilities
 * Detects the user's operating system and platform
 */

/**
 * Detects if the device uses Apple emoji
 * Apple emoji are used on macOS and iOS devices (iPhone, iPad, iPod)
 * @returns true if the device is running macOS or iOS
 */
export function hasAppleEmoji(): boolean {
	if (typeof window === 'undefined' || typeof navigator === 'undefined') {
		return false;
	}
	
	const platform = navigator.platform || '';
	const userAgent = navigator.userAgent || '';
	
	// Check platform first (more reliable)
	if (/Mac|iPhone|iPad|iPod/.test(platform)) {
		return true;
	}
	
	// Fallback to user agent
	if (/Mac|iPhone|iPad|iPod/.test(userAgent)) {
		return true;
	}
	
	return false;
}
