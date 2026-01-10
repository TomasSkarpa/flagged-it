/** @type {import('tailwindcss').Config} */
export default {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			colors: {
				// Primary Palette - Modern Dark Theme
				'primary': {
					DEFAULT: '#6366F1',
					light: '#818CF8',
					dark: '#4F46E5'
				},
				'secondary': {
					DEFAULT: '#8B5CF6',
					light: '#A78BFA',
					dark: '#7C3AED'
				},
				'accent': {
					DEFAULT: '#06B6D4',
					light: '#22D3EE',
					dark: '#0891B2'
				},
				// Background Colors
				'bg': {
					DEFAULT: '#0A0E27',
					light: '#111827',
					dark: '#050810'
				},
				'surface': {
					DEFAULT: '#1E293B',
					light: '#334155',
					dark: '#0F172A'
				},
				// Text Colors
				'text': {
					DEFAULT: '#F8FAFC',
					light: '#FFFFFF',
					muted: '#94A3B8',
					dark: '#64748B'
				},
				// Status Colors
				'success': {
					DEFAULT: '#10B981',
					light: '#34D399',
					dark: '#059669'
				},
				'error': {
					DEFAULT: '#EF4444',
					light: '#F87171',
					dark: '#DC2626'
				},
				'warning': {
					DEFAULT: '#F59E0B',
					light: '#FBBF24',
					dark: '#D97706'
				},
				'info': {
					DEFAULT: '#3B82F6',
					light: '#60A5FA',
					dark: '#2563EB'
				},
				// Legacy color mappings for backward compatibility
				'ocean': {
					DEFAULT: '#0A0E27',
					light: '#111827',
					dark: '#050810'
				},
				'terracotta': {
					DEFAULT: '#6366F1',
					light: '#818CF8',
					dark: '#4F46E5'
				},
				'sage': {
					DEFAULT: '#10B981',
					light: '#34D399',
					dark: '#059669'
				},
				'sandy': {
					DEFAULT: '#F8FAFC',
					light: '#FFFFFF',
					dark: '#94A3B8'
				},
				'slate': {
					DEFAULT: '#64748B',
					light: '#94A3B8',
					dark: '#0F172A'
				},
				'sunset': {
					DEFAULT: '#F59E0B',
					light: '#FBBF24',
					dark: '#D97706'
				},
				'pin-red': {
					DEFAULT: '#EF4444',
					light: '#F87171',
					dark: '#DC2626'
				},
				'sky': {
					DEFAULT: '#06B6D4',
					light: '#22D3EE',
					dark: '#0891B2'
				},
				'bronze': {
					DEFAULT: '#F59E0B',
					light: '#FBBF24',
					dark: '#D97706'
				}
			},
			fontFamily: {
				'heading': ['Montserrat', 'Poppins', 'system-ui', 'sans-serif'],
				'body': ['Inter', 'Open Sans', 'system-ui', 'sans-serif'],
				'mono': ['Roboto Mono', 'Consolas', 'monospace']
			},
			fontSize: {
				'game-title': ['3rem', { lineHeight: '1.2', fontWeight: '700' }],
				'section': ['2rem', { lineHeight: '1.3', fontWeight: '700' }],
				'card-title': ['1.5rem', { lineHeight: '1.4', fontWeight: '600' }]
			},
			spacing: {
				'18': '4.5rem',
				'88': '22rem'
			},
			borderRadius: {
				'card': '1.5rem',
				'button': '0.75rem'
			},
			boxShadow: {
				'card': '0 4px 20px rgba(0, 0, 0, 0.4)',
				'card-hover': '0 8px 30px rgba(0, 0, 0, 0.5)',
				'glow': '0 0 24px rgba(99, 102, 241, 0.5)',
				'glow-accent': '0 0 24px rgba(6, 182, 212, 0.4)'
			},
			backdropBlur: {
				'card': '10px'
			}
		}
	},
	plugins: []
};

