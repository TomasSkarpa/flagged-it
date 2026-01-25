<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { locale } from '$lib/stores/locale';
	import { getKeyboardLayoutForLocale } from '$lib/utils/keyboardLayout';

	import type { KeyboardLayout, KeyState } from './keyboardTypes';
	
	// Re-export types for use in other components
	export type { KeyboardLayout, KeyState };

	export let layout: KeyboardLayout | null = null; // null = auto-detect from locale
	export let keyStates: Record<string, KeyState> = {};
	export let disabled: boolean = false;
	
	// Auto-detect layout from locale if not explicitly set
	// Ensure we always have a valid layout, fallback to 'english' if locale is unavailable
	let effectiveLayout: KeyboardLayout = 'english';
	$: effectiveLayout = layout || (typeof $locale !== 'undefined' && $locale ? getKeyboardLayoutForLocale($locale) : 'english');

	const dispatch = createEventDispatcher<{
		keypress: { key: string };
	}>();

	// Keyboard layouts with offset info (true means offset second row like QWERTY)
	const layouts: Record<KeyboardLayout, { rows: string[][]; offset?: boolean }> = {
		// QWERTY-based layouts (offset second row)
		english: {
			rows: [
				['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
				['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'],
				['Z', 'X', 'C', 'V', 'B', 'N', 'M']
			],
			offset: true
		},
		indonesian: {
			// Indonesian uses standard Latin (QWERTY)
			rows: [
				['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
				['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L'],
				['Z', 'X', 'C', 'V', 'B', 'N', 'M']
			],
			offset: true
		},
		spanish: {
			// Spanish with special characters
			rows: [
				['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'I', 'O', 'P'],
				['A', 'S', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'Ñ'],
				['Z', 'X', 'C', 'V', 'B', 'N', 'M']
			],
			offset: true
		},
		turkish: {
			// Turkish Q layout (QWERTY-based with Turkish characters)
			// Note: Turkish has both İ/i (dotted) and I/ı (dotless)
			rows: [
				['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'Ü', 'I', 'ı', 'O', 'Ö', 'P'],
				['A', 'S', 'D', 'F', 'G', 'Ğ', 'H', 'J', 'K', 'L', 'Ş'],
				['Z', 'X', 'C', 'Ç', 'V', 'B', 'N', 'M', 'İ', 'i']
			],
			offset: true
		},
		vietnamese: {
			// Vietnamese QWERTY with diacritics (Vietnamese Telex/VNI input method compatible)
			// Includes: Ă, Â, Ê, Ô, Ơ, Ư, Đ
			rows: [
				['Q', 'W', 'E', 'R', 'T', 'Y', 'U', 'Ư', 'I', 'O', 'Ô', 'Ơ', 'P'],
				['A', 'Ă', 'Â', 'S', 'D', 'Đ', 'F', 'G', 'H', 'J', 'K', 'L'],
				['Z', 'X', 'C', 'V', 'B', 'N', 'M', 'Ê']
			],
			offset: true
		},
		polish: {
			// Polish QWERTZ layout with all diacritics
			// Includes: Ą, Ć, Ę, Ł, Ń, Ó, Ś, Ź, Ż
			rows: [
				['Q', 'W', 'E', 'Ę', 'R', 'T', 'Y', 'U', 'Ó', 'I', 'O', 'P'],
				['A', 'Ą', 'S', 'Ś', 'D', 'Ć', 'F', 'G', 'H', 'J', 'K', 'L', 'Ł'],
				['Z', 'Ź', 'Ż', 'X', 'C', 'V', 'B', 'N', 'Ń', 'M']
			],
			offset: true
		},
		czech: {
			// Czech QWERTZ layout with all diacritics
			// Includes: Á, Č, Ď, É, Ě, Í, Ň, Ó, Ř, Š, Ť, Ú, Ů, Ý, Ž
			rows: [
				['Q', 'W', 'E', 'É', 'R', 'Ř', 'T', 'Ť', 'Z', 'Ž', 'U', 'Ú', 'I', 'Í', 'O', 'Ó', 'P', 'Ý'],
				['A', 'Á', 'S', 'Š', 'D', 'Ď', 'F', 'G', 'H', 'J', 'K', 'L'],
				['Y', 'X', 'C', 'Č', 'V', 'B', 'N', 'Ň', 'M', 'Ů', 'Ě']
			],
			offset: true
		},
		// Non-QWERTY layouts (no offset)
		korean: {
			// Korean Dubeolsik (두벌식) layout - includes all consonants and vowels (basic + compound)
			rows: [
				['ㅃ', 'ㅉ', 'ㄸ', 'ㄲ', 'ㅆ', 'ㅛ', 'ㅕ', 'ㅑ', 'ㅐ', 'ㅔ', 'ㅒ', 'ㅖ'],
				['ㅁ', 'ㄴ', 'ㅇ', 'ㄹ', 'ㅎ', 'ㅗ', 'ㅓ', 'ㅏ', 'ㅣ', 'ㅘ', 'ㅙ', 'ㅚ'],
				['ㅋ', 'ㅌ', 'ㅊ', 'ㅍ', 'ㅠ', 'ㅜ', 'ㅡ', 'ㅝ', 'ㅞ', 'ㅟ', 'ㅢ', 'ㅂ', 'ㅈ', 'ㄷ', 'ㄱ', 'ㅅ']
			],
			offset: false
		},
		japanese: {
			// Japanese Hiragana layout - includes all basic, dakuten, handakuten, and small tsu
			// Basic hiragana (あいうえお order)
			rows: [
				['あ', 'か', 'が', 'さ', 'ざ', 'た', 'だ', 'な', 'は', 'ば', 'ぱ', 'ま', 'や', 'ら', 'わ', 'ん'],
				['い', 'き', 'ぎ', 'し', 'じ', 'ち', 'ぢ', 'に', 'ひ', 'び', 'ぴ', 'み', 'ゆ', 'り', 'を'],
				['う', 'く', 'ぐ', 'す', 'ず', 'つ', 'づ', 'っ', 'ぬ', 'ふ', 'ぶ', 'ぷ', 'む', 'よ', 'る'],
				['え', 'け', 'げ', 'せ', 'ぜ', 'て', 'で', 'ね', 'へ', 'べ', 'ぺ', 'め', 'れ'],
				['お', 'こ', 'ご', 'そ', 'ぞ', 'と', 'ど', 'の', 'ほ', 'ぼ', 'ぽ', 'も', 'ろ']
			],
			offset: false
		},
		russian: {
			// Russian ЙЦУКЕН (JCUKEN) layout
			rows: [
				['Й', 'Ц', 'У', 'К', 'Е', 'Н', 'Г', 'Ш', 'Щ', 'З', 'Х', 'Ъ'],
				['Ф', 'Ы', 'В', 'А', 'П', 'Р', 'О', 'Л', 'Д', 'Ж', 'Э'],
				['Я', 'Ч', 'С', 'М', 'И', 'Т', 'Ь', 'Б', 'Ю', 'Ё']
			],
			offset: false
		},
		ukrainian: {
			// Ukrainian ЙЦУКЕН (JCUKEN) layout - includes Ukrainian-specific letters (ґ, є, і, ї)
			rows: [
				['Й', 'Ц', 'У', 'К', 'Е', 'Н', 'Г', 'Ш', 'Щ', 'З', 'Х', 'Ї'],
				['Ф', 'І', 'В', 'А', 'П', 'Р', 'О', 'Л', 'Д', 'Ж', 'Є'],
				['Я', 'Ч', 'С', 'М', 'И', 'Т', 'Ь', 'Б', 'Ю', 'Ґ']
			],
			offset: false
		},
		arabic: {
			// Arabic keyboard layout (QWERTY-based Arabic)
			rows: [
				['ض', 'ص', 'ث', 'ق', 'ف', 'غ', 'ع', 'ه', 'خ', 'ح', 'ج', 'د'],
				['ش', 'س', 'ي', 'ب', 'ل', 'ا', 'ت', 'ن', 'م', 'ك', 'ط'],
				['ئ', 'ء', 'ؤ', 'ر', 'لا', 'ى', 'ة', 'و', 'ز', 'ظ']
			],
			offset: false
		},
		thai: {
			// Thai Kedmanee layout - includes all 44 consonants and 14 vowels
			// Consonants: ก ข ฃ ค ฅ ฆ ง จ ฉ ช ซ ฌ ญ ฎ ฏ ฐ ฑ ฒ ณ ด ต ถ ท ธ น บ ป ผ ฝ พ ฟ ภ ม ย ร ฤ ล ฦ ว ศ ษ ส ห ฬ อ ฮ
			// Vowels: ะ ั า ำ ิ ี ึ ื ุ ู เ แ โ ใ ไ
			rows: [
				['ๆ', 'ไ', 'ำ', 'พ', 'ะ', 'ร', 'น', 'ย', 'บ', 'ล', 'ฃ', 'ฟ'],
				['ห', 'ก', 'ด', 'เ', '้', '่', 'า', 'ส', 'ว', 'ง', 'ผ', 'ป'],
				['แ', 'อ', 'ิ', 'ื', 'ท', 'ม', 'ใ', 'ฝ', 'ฅ', 'ถ', 'ค', 'ต', 'จ'],
				['ข', 'ฆ', 'ฉ', 'ช', 'ซ', 'ฌ', 'ญ', 'ฎ', 'ฏ', 'ฐ', 'ฑ', 'ฒ', 'ณ', 'ธ', 'ภ', 'ฤ', 'ฦ', 'ศ', 'ษ', 'ฬ', 'ฮ', 'ั', 'ี', 'ึ', 'ุ', 'ู', 'โ']
			],
			offset: false
		},
		greek: {
			// Greek keyboard layout - includes both uppercase and lowercase
			// Note: σ (sigma) and ς (final sigma) are both included as they're different forms
			rows: [
				['Α', 'Β', 'Γ', 'Δ', 'Ε', 'Ζ', 'Η', 'Θ', 'Ι', 'Κ', 'Λ', 'Μ'],
				['Ν', 'Ξ', 'Ο', 'Π', 'Ρ', 'Σ', 'Τ', 'Υ', 'Φ', 'Χ', 'Ψ'],
				['Ω', 'α', 'β', 'γ', 'δ', 'ε', 'ζ', 'η', 'θ', 'ι', 'κ', 'λ', 'μ'],
				['ν', 'ξ', 'ο', 'π', 'ρ', 'σ', 'ς', 'τ', 'υ', 'φ', 'χ', 'ψ', 'ω']
			],
			offset: false
		},
		hebrew: {
			// Hebrew keyboard layout
			rows: [
				['/', '\'', 'ק', 'ר', 'א', 'ט', 'ו', 'ן', 'ם', 'פ'],
				['ש', 'ד', 'ג', 'כ', 'ע', 'י', 'ח', 'ל', 'ך', 'ף'],
				['ז', 'ס', 'ב', 'ה', 'נ', 'מ', 'צ', 'ת', 'ץ']
			],
			offset: false
		}
	};

	function handleKeyClick(key: string) {
		if (disabled || keyStates[key] === 'disabled') {
			return;
		}
		dispatch('keypress', { key });
	}

	function getKeyClass(key: string): string {
		const state = keyStates[key] || 'default';
		const baseClass = 'key';
		
		switch (state) {
			case 'correct':
				return `${baseClass} key-correct`;
			case 'incorrect':
				return `${baseClass} key-incorrect`;
			case 'disabled':
				return `${baseClass} key-disabled`;
			default:
				return baseClass;
		}
	}

	$: currentLayoutData = layouts[effectiveLayout] || layouts.english;
	$: currentLayout = currentLayoutData ? currentLayoutData.rows : [];
	$: hasOffset = currentLayoutData ? (currentLayoutData.offset || false) : false;
	$: isRTL = effectiveLayout === 'arabic' || effectiveLayout === 'hebrew';
	$: needsLargeKeys = ['korean', 'japanese', 'arabic', 'hebrew', 'thai', 'greek'].includes(effectiveLayout);
</script>

<div class="keyboard" class:disabled={disabled} class:keyboard-large-keys={needsLargeKeys} dir={isRTL ? 'rtl' : 'ltr'}>
	{#each currentLayout as row, rowIndex}
		<div class="keyboard-row" class:keyboard-row-offset={hasOffset && rowIndex === 1}>
			{#each row as key}
			{@const keyState = keyStates[key] || 'default'}
			<button
				class="key"
				class:key-correct={keyState === 'correct'}
				class:key-incorrect={keyState === 'incorrect'}
				class:key-disabled={keyState === 'disabled'}
				on:click={() => handleKeyClick(key)}
				disabled={disabled || keyState === 'disabled'}
				type="button"
			>
				{key}
			</button>
		{/each}
		</div>
	{/each}
</div>

<style>
	.keyboard {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
		padding: 0.5rem;
		background: var(--color-surface);
		border-radius: 0.75rem;
		user-select: none;
	}

	.keyboard.disabled {
		opacity: 0.6;
		/* Don't use pointer-events: none here - let individual buttons handle their disabled state */
	}

	.keyboard-row {
		display: flex;
		justify-content: center;
		gap: 0.25rem;
		flex-wrap: wrap;
	}

	.keyboard-row-offset {
		margin-left: 1.5rem; /* Offset for QWERTY-style layouts (second row) */
	}

	/* RTL support for Arabic and Hebrew */
	.keyboard[dir="rtl"] .keyboard-row {
		direction: rtl;
	}

	.keyboard[dir="rtl"] .keyboard-row-offset {
		margin-left: 0;
		margin-right: 1.5rem;
	}

	.key {
		min-width: 2.25rem;
		height: 2.5rem;
		padding: 0.375rem 0.5rem;
		border: 2px solid rgba(255, 255, 255, 0.1);
		border-radius: 0.5rem;
		background: var(--color-surface-light);
		color: var(--color-text-light);
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		display: flex;
		align-items: center;
		justify-content: center;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
	}

	/* Special font sizing for non-Latin scripts */
	.keyboard[dir="rtl"] .key,
	.keyboard-large-keys .key {
		font-size: 0.9375rem;
		min-width: 2.5rem;
	}

	.key:hover:not(:disabled) {
		background: var(--color-surface);
		border-color: var(--color-primary);
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
	}

	.key:active:not(:disabled) {
		transform: translateY(0);
	}

	.key:disabled {
		cursor: not-allowed;
		opacity: 0.5;
	}

	:global(:root.light) .key {
		border-color: rgba(0, 0, 0, 0.15);
	}

	.key-correct {
		background: var(--color-success) !important;
		border-color: var(--color-success) !important;
		color: white !important;
		box-shadow: 0 0 20px rgba(16, 185, 129, 0.4);
	}

	.key-incorrect {
		background: var(--color-error) !important;
		border-color: var(--color-error) !important;
		color: white !important;
		box-shadow: 0 0 20px rgba(239, 68, 68, 0.4);
	}

	.key-disabled {
		background: var(--color-surface-dark) !important;
		border-color: rgba(255, 255, 255, 0.05) !important;
		color: var(--color-text-muted) !important;
		opacity: 0.4;
	}

	:global(:root.light) .key-disabled {
		border-color: rgba(0, 0, 0, 0.08) !important;
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		.key {
			min-width: 1.875rem;
			height: 2.25rem;
			padding: 0.25rem 0.375rem;
			font-size: 0.6875rem;
		}

		.keyboard-row {
			gap: 0.25rem;
		}
	}

	/* Special handling for longer keyboards (Japanese, Russian) */
	:global(.keyboard-row:has(button:nth-child(10))) {
		gap: 0.25rem;
	}

	:global(.keyboard-row:has(button:nth-child(12))) {
		gap: 0.2rem;
	}
</style>
