<script lang="ts">
	import { page } from '$app/stores';
	import LanguageSelector from './LanguageSelector.svelte';
	import ThemeToggle from './ThemeToggle.svelte';
	import { t } from '$lib/translations';
	import { locale } from '$lib/stores/locale';
	
	let mobileMenuOpen = false;
	
	// Reactive translations - will update when locale changes
	$: currentLocale = $locale;
	$: navLinks = [
		{ href: '/', label: t('nav.games', undefined, currentLocale), icon: '🎮' },
		{ href: '/library', label: t('nav.library', undefined, currentLocale), icon: '📚' },
		{ href: '/scoreboard', label: t('nav.scoreboard', undefined, currentLocale), icon: '🏆' }
	];
	$: homeTitle = t('home.hero.title', undefined, currentLocale);
	
	function toggleMobileMenu() {
		mobileMenuOpen = !mobileMenuOpen;
	}
	
	function closeMobileMenu() {
		mobileMenuOpen = false;
	}
</script>

<nav class="nav-container">
	<div class="nav-content">
		<!-- Logo -->
		<a href="/" class="nav-logo" on:click={closeMobileMenu}>
			<span class="logo-icon">🌍</span>
			<span class="logo-text">{homeTitle}</span>
		</a>
		
		<!-- Desktop Navigation -->
		<div class="nav-links-desktop">
			{#each navLinks as link}
				<a 
					href={link.href} 
					class="nav-link"
					class:active={$page.url.pathname === link.href}
				>
					<span class="nav-link-icon">{link.icon}</span>
					<span>{link.label}</span>
				</a>
			{/each}
		</div>
		
		<!-- Right side controls -->
		<div class="nav-controls">
			<LanguageSelector />
			<ThemeToggle />
			
			<!-- Mobile menu button -->
			<button 
				class="mobile-menu-btn"
				on:click={toggleMobileMenu}
				aria-label="Toggle menu"
			>
				{#if mobileMenuOpen}
					<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				{:else}
					<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<line x1="3" y1="12" x2="21" y2="12"></line>
						<line x1="3" y1="6" x2="21" y2="6"></line>
						<line x1="3" y1="18" x2="21" y2="18"></line>
					</svg>
				{/if}
			</button>
		</div>
	</div>
	
	<!-- Mobile Navigation Menu -->
	{#if mobileMenuOpen}
		<div class="mobile-menu">
			{#each navLinks as link}
				<a 
					href={link.href} 
					class="mobile-nav-link"
					class:active={$page.url.pathname === link.href}
					on:click={closeMobileMenu}
				>
					<span class="nav-link-icon">{link.icon}</span>
					<span>{link.label}</span>
				</a>
			{/each}
		</div>
	{/if}
</nav>

<style>
	.nav-container {
		position: sticky;
		top: 0;
		z-index: 50;
		background: rgba(var(--nav-bg-rgb, 10, 14, 39), 0.85);
		backdrop-filter: blur(12px);
		border-bottom: 1px solid rgba(255, 255, 255, 0.08);
	}
	
	.nav-content {
		max-width: 80rem;
		margin: 0 auto;
		padding: 0 1rem;
		height: 4rem;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
	}
	
	.nav-logo {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		text-decoration: none;
		font-weight: 700;
		font-size: 1.25rem;
		color: var(--color-text-light);
		transition: opacity 0.2s;
	}
	
	.nav-logo:hover {
		opacity: 0.8;
	}
	
	.logo-icon {
		font-size: 1.5rem;
	}
	
	.logo-text {
		background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}
	
	.nav-links-desktop {
		display: none;
		align-items: center;
		gap: 0.5rem;
	}
	
	@media (min-width: 768px) {
		.nav-links-desktop {
			display: flex;
		}
	}
	
	.nav-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: 0.5rem;
		text-decoration: none;
		color: var(--color-text-muted);
		font-weight: 500;
		font-size: 0.875rem;
		transition: all 0.2s;
	}
	
	.nav-link:hover {
		color: var(--color-text-light);
		background: rgba(255, 255, 255, 0.05);
	}
	
	:global(:root.light) .nav-link:hover {
		background: rgba(0, 0, 0, 0.06);
	}
	
	.nav-link.active {
		color: var(--color-primary-light);
		background: rgba(99, 102, 241, 0.1);
	}
	
	.nav-link-icon {
		font-size: 1rem;
	}
	
	.nav-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	
	.nav-controls > * {
		flex-shrink: 0;
	}
	
	.mobile-menu-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 0.5rem;
		min-width: 2.5rem;
		height: 2.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		background: rgba(255, 255, 255, 0.05);
		color: var(--color-text-muted);
		cursor: pointer;
		border-radius: 0.5rem;
		transition: all 0.2s;
	}
	
	.mobile-menu-btn:hover {
		color: var(--color-text-light);
		background: rgba(255, 255, 255, 0.05);
	}
	
	:global(:root.light) .mobile-menu-btn {
		border-color: rgba(0, 0, 0, 0.15);
		background: rgba(0, 0, 0, 0.04);
	}
	
	:global(:root.light) .mobile-menu-btn:hover {
		background: rgba(0, 0, 0, 0.08);
		border-color: rgba(0, 0, 0, 0.25);
	}
	
	@media (min-width: 768px) {
		.mobile-menu-btn {
			display: none;
		}
	}
	
	.mobile-menu {
		display: flex;
		flex-direction: column;
		padding: 0.5rem;
		border-top: 1px solid rgba(255, 255, 255, 0.08);
		animation: slideDown 0.2s ease-out;
	}
	
	@keyframes slideDown {
		from {
			opacity: 0;
			transform: translateY(-10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
	
	@media (min-width: 768px) {
		.mobile-menu {
			display: none;
		}
	}
	
	.mobile-nav-link {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		border-radius: 0.5rem;
		text-decoration: none;
		color: var(--color-text-muted);
		font-weight: 500;
		transition: all 0.2s;
	}
	
	.mobile-nav-link:hover {
		color: var(--color-text-light);
		background: rgba(255, 255, 255, 0.05);
	}
	
	:global(:root.light) .mobile-nav-link:hover {
		background: rgba(0, 0, 0, 0.06);
	}
	
	.mobile-nav-link.active {
		color: var(--color-primary-light);
		background: rgba(99, 102, 241, 0.1);
	}
	
	:global(:root.light) .mobile-menu {
		border-top-color: rgba(0, 0, 0, 0.1);
	}
</style>
