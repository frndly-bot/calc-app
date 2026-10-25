<script lang="ts">
	import { onMount } from 'svelte';
	import { Events } from '@wailsio/runtime';
	import { fade, fly } from 'svelte/transition';
	import HistoryPanel from '#lib/components/HistoryPanel.svelte';
	import Icon from '#lib/components/Icon.svelte';
	import TabBar from '#lib/components/TabBar.svelte';
	import Toast from '#lib/components/Toast.svelte';
	import DeveloperMode from '#lib/modes/DeveloperMode.svelte';
	import FinanceMode from '#lib/modes/FinanceMode.svelte';
	import NormalMode from '#lib/modes/NormalMode.svelte';
	import ScientificMode from '#lib/modes/ScientificMode.svelte';
	import { history } from '#lib/stores/history.svelte';
	import {
		accentFor,
		mode,
		modeLabel,
		modes,
		type Mode
	} from '#lib/stores/mode.svelte';
	import { motion } from '#lib/stores/motion.svelte';
	import { settings } from '#lib/stores/settings.svelte';
	import { ui } from '#lib/stores/ui.svelte';

	const marks: Record<Mode, string> = {
		normal: '=',
		scientific: '∑',
		developer: '</>',
		finance: '$'
	};

	const shortcuts = $derived(
		mode.current === 'finance'
			? [
				{ key: 'Enter', label: 'calculate' },
				{ key: 'Esc', label: 'clear' }
			]
			: [
					{ key: 'Enter', label: 'equals' },
					{ key: 'Esc', label: 'clear' },
					{ key: '⌫', label: 'delete' }
				]
	);

	function isMode(value: string): value is Mode {
		return modes.some((item) => item.id === value);
	}

	onMount(() => {
		let cancelled = false;
		const cleanups: Array<() => void> = [];

		void (async () => {
			await settings.load();
			if (cancelled) return;
			if (isMode(settings.current.mode)) mode.set(settings.current.mode, false);
			await history.load();
		})();

		try {
			cleanups.push(Events.On('history:changed', () => void history.load()));
		} catch {
			// Events only exist inside the Wails host; the preview works without them.
		}

		return () => {
			cancelled = true;
			for (const cleanup of cleanups) cleanup();
		};
	});

	$effect(() => {
		const accent = accentFor(mode.current);
		document.documentElement.style.setProperty('--accent', accent);
		document.documentElement.style.setProperty(
			'--accent-contrast',
			mode.current === 'finance' ? '#1b1302' : '#ffffff'
		);
	});

	function onKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && ui.historyOpen) {
			ui.closeHistory();
			event.preventDefault();
		}
	}
</script>

<svelte:window onkeydown={onKeydown} />

<div class="app">
	<header class="top">
		<div class="brand">
			<span class="mark">{marks[mode.current]}</span>
			<div class="titles">
				<h1>Calc</h1>
				<p>{modeLabel(mode.current)}</p>
			</div>
		</div>
		<div class="actions">
			<button
				type="button"
				class="icon-button"
				aria-label={settings.current.theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
				title={settings.current.theme === 'dark' ? 'Light theme' : 'Dark theme'}
				onclick={() => void settings.toggleTheme()}
			>
				<Icon name={settings.current.theme === 'dark' ? 'sun' : 'moon'} />
			</button>
			<button
				type="button"
				class="icon-button history-button"
				aria-label="Calculation history"
				title="History"
				onclick={() => ui.openHistory()}
			>
				<Icon name="history" />
				{#if history.entries.length > 0}<span class="dot"></span>{/if}
			</button>
		</div>
	</header>

	<TabBar />

	<main class="stage">
		{#key mode.current}
			<div
				class="mode-view"
				in:fly={{ y: motion.reduced ? 0 : 10, duration: motion.reduced ? 0 : 220 }}
				out:fade={{ duration: motion.reduced ? 0 : 110 }}
			>
				{#if mode.current === 'normal'}
					<NormalMode />
				{:else if mode.current === 'scientific'}
					<ScientificMode />
				{:else if mode.current === 'developer'}
					<DeveloperMode />
				{:else}
					<FinanceMode />
				{/if}
			</div>
		{/key}
	</main>

	<footer class="shortcuts">
		{#each shortcuts as shortcut (shortcut.key)}
			<span><kbd>{shortcut.key}</kbd>{shortcut.label}</span>
		{/each}
	</footer>

	<HistoryPanel open={ui.historyOpen} onclose={() => ui.closeHistory()} />
	<Toast />
</div>

<style>
	.app {
		display: flex;
		flex-direction: column;
		gap: 10px;
		width: 100%;
		max-width: 460px;
		height: 100dvh;
		margin: 0 auto;
		padding: 14px 14px 9px;
		animation: window-in 0.42s cubic-bezier(0.22, 1, 0.36, 1) both;
	}

	@keyframes window-in {
		from {
			opacity: 0;
			transform: translateY(10px) scale(0.985);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-shrink: 0;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.mark {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 38px;
		height: 38px;
		border-radius: 12px;
		background: linear-gradient(
			145deg,
			color-mix(in srgb, var(--accent) 92%, white 8%),
			color-mix(in srgb, var(--accent) 70%, black 30%)
		);
		color: var(--accent-contrast);
		font-family: var(--mono);
		font-size: 15px;
		font-weight: 700;
		letter-spacing: -0.03em;
		box-shadow: 0 8px 22px color-mix(in srgb, var(--accent) 35%, transparent);
		transition:
			background 0.3s ease,
			box-shadow 0.3s ease;
	}

	.titles h1 {
		margin: 0;
		font-size: 17px;
		font-weight: 650;
		letter-spacing: -0.02em;
	}

	.titles p {
		margin: 1px 0 0;
		color: var(--muted);
		font-size: 10.5px;
		font-weight: 600;
		letter-spacing: 0.07em;
		text-transform: uppercase;
	}

	.actions {
		display: flex;
		gap: 7px;
		flex-shrink: 0;
	}

	.history-button {
		position: relative;
	}

	.dot {
		position: absolute;
		top: 7px;
		right: 8px;
		width: 5px;
		height: 5px;
		border-radius: 50%;
		background: var(--accent);
		box-shadow: 0 0 7px var(--accent);
	}

	.stage {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		overflow-x: hidden;
		overscroll-behavior: contain;
		padding: 1px;
		margin: -1px;
	}

	.mode-view {
		display: block;
	}

	.shortcuts {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 14px;
		flex-shrink: 0;
		color: var(--muted);
		font-size: 10px;
	}

	.shortcuts span {
		display: inline-flex;
		align-items: center;
		gap: 4px;
	}

	kbd {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-width: 18px;
		height: 17px;
		padding: 0 4px;
		border: 1px solid var(--border);
		border-radius: 5px;
		background: var(--surface);
		color: var(--text);
		font-family: var(--font);
		font-size: 9px;
		font-weight: 650;
		line-height: 1;
	}

	@media (max-height: 640px) {
		.shortcuts {
			display: none;
		}
	}

	@media (max-width: 380px) {
		.app {
			padding: 10px 10px 8px;
		}

		.mark {
			width: 34px;
			height: 34px;
		}
	}
</style>
