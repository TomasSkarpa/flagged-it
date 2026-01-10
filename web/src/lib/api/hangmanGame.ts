import { getApiUrl, API_ENDPOINTS } from './config';
import type { Country } from '../types';
import { getCountryName as getTranslatedCountryName } from '$lib/utils/countryNames';

export interface HangmanGameState {
	currentWord: string; // Country name in user's language
	guessedWord: string[]; // Array of characters/underscores
	guessedLetters: string[]; // Letters that have been guessed
	wrongGuesses: number;
	maxWrongGuesses: number;
	score: number;
	total: number;
	isComplete: boolean;
	isWon: boolean;
	country: Country; // The country object for reference
}

export interface HangmanGuessResult {
	isValidGuess: boolean;
	isInWord: boolean;
	guessedWord: string[];
	wrongGuesses: number;
	isWon: boolean;
	isGameOver: boolean;
	score: number;
	total: number;
	isComplete: boolean;
	revealedWord?: string;
	error?: string;
}

// Client-side hangman game logic
export class HangmanGame {
	private countries: Country[] = [];
	private state: HangmanGameState;
	private locale: string = 'en';
	private maxRounds: number = 5;

	constructor(countries: Country[], locale: string = 'en') {
		this.countries = countries;
		this.locale = locale;
		this.state = {
			currentWord: '',
			guessedWord: [],
			guessedLetters: [],
			wrongGuesses: 0,
			maxWrongGuesses: 6,
			score: 0,
			total: 0,
			isComplete: false,
			isWon: false,
			country: {} as Country
		};
	}

	// Get country name in user's language
	private getCountryName(country: Country, locale: string): string {
		return getTranslatedCountryName(country, locale);
	}

	// Start a new round
	newRound(): void {
		if (this.state.total >= this.maxRounds) {
			this.state.isComplete = true;
			return;
		}

		if (this.countries.length === 0) {
			throw new Error('No countries available');
		}

		// Select random country
		const randomIndex = Math.floor(Math.random() * this.countries.length);
		const country = this.countries[randomIndex];
		
		// Get country name in user's language
		const countryName = this.getCountryName(country, this.locale).toUpperCase();
		
		this.state.country = country;
		this.state.currentWord = countryName;
		
		// Initialize guessed word with underscores
		this.state.guessedWord = countryName.split('').map(char => {
			if (char === ' ') {
				return ' ';
			}
			return '_';
		});

		this.state.guessedLetters = [];
		this.state.wrongGuesses = 0;
		this.state.isWon = false;
	}

	// Make a guess
	makeGuess(letter: string): HangmanGuessResult {
		if (this.state.isComplete || !this.state.currentWord) {
			return {
				isValidGuess: false,
				isInWord: false,
				guessedWord: this.state.guessedWord,
				wrongGuesses: this.state.wrongGuesses,
				isWon: false,
				isGameOver: false,
				score: this.state.score,
				total: this.state.total,
				isComplete: true,
				error: 'Game is complete'
			};
		}

		const upperLetter = letter.toUpperCase();

		// Check if already guessed
		if (this.state.guessedLetters.includes(upperLetter)) {
			return {
				isValidGuess: false,
				isInWord: false,
				guessedWord: this.state.guessedWord,
				wrongGuesses: this.state.wrongGuesses,
				isWon: false,
				isGameOver: false,
				score: this.state.score,
				total: this.state.total,
				isComplete: false,
				error: 'Letter already guessed'
			};
		}

		// Add to guessed letters
		this.state.guessedLetters.push(upperLetter);

		// Check if letter is in word
		const isInWord = this.state.currentWord.includes(upperLetter);

		if (isInWord) {
			// Reveal all instances of the letter
			for (let i = 0; i < this.state.currentWord.length; i++) {
				if (this.state.currentWord[i] === upperLetter) {
					this.state.guessedWord[i] = upperLetter;
				}
			}

			// Check if word is complete
			const isWon = !this.state.guessedWord.includes('_');

			if (isWon) {
				this.state.isWon = true;
				this.state.total++;
				this.state.score++;
			}

			return {
				isValidGuess: true,
				isInWord: true,
				guessedWord: [...this.state.guessedWord],
				wrongGuesses: this.state.wrongGuesses,
				isWon: isWon,
				isGameOver: false,
				score: this.state.score,
				total: this.state.total,
				isComplete: this.state.total >= this.maxRounds,
				revealedWord: isWon ? this.state.currentWord : undefined
			};
		} else {
			// Wrong guess
			this.state.wrongGuesses++;
			const isGameOver = this.state.wrongGuesses >= this.state.maxWrongGuesses;

			if (isGameOver) {
				this.state.total++;
			}

			return {
				isValidGuess: true,
				isInWord: false,
				guessedWord: [...this.state.guessedWord],
				wrongGuesses: this.state.wrongGuesses,
				isWon: false,
				isGameOver: isGameOver,
				score: this.state.score,
				total: this.state.total,
				isComplete: this.state.total >= this.maxRounds,
				revealedWord: isGameOver ? this.state.currentWord : undefined
			};
		}
	}

	// Get current state
	getState(): HangmanGameState {
		return { ...this.state };
	}

	// Reset game
	reset(): void {
		this.state = {
			currentWord: '',
			guessedWord: [],
			guessedLetters: [],
			wrongGuesses: 0,
			maxWrongGuesses: 6,
			score: 0,
			total: 0,
			isComplete: false,
			isWon: false,
			country: {} as Country
		};
	}

	// Update locale
	setLocale(locale: string): void {
		this.locale = locale;
	}
}
