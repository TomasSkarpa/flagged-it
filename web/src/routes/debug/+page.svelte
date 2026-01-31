<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

<svelte:head>
	<title>Debug - Flagged It</title>
</svelte:head>

<div class="min-h-screen p-4 md:p-8">
	<div class="max-w-4xl mx-auto">
		<h1 class="text-4xl md:text-5xl font-bold text-sandy-light mb-4">
			<span class="emoji-blue mr-2">🔧</span>
			<span class="gradient-text">Debug Browser</span>
		</h1>
		<p class="text-lg text-text-muted mb-8">
			Browse all available game data and assets.
		</p>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			{#each data.routes as route}
				<button 
					class="card-game text-left hover:ring-2 hover:ring-accent transition-all"
					on:click={() => goto(route.route)}
				>
					<div class="flex items-start gap-4">
						<div class="text-5xl leading-none pt-1">{route.icon}</div>
						<div class="flex-1">
							<h2 class="text-2xl font-bold text-sandy-light mb-2">{route.title}</h2>
							<p class="text-text-muted">{route.description}</p>
						</div>
					</div>
				</button>
			{/each}
		</div>
	</div>
</div>
