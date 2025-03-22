<script lang="ts">
	import { getAvailableLanguages, getCurrentLanguage, setLanguage, type Language } from '$lib/i18n';
	import Section from './Section.svelte';

	function handleLanguageChange(event: Event) {
		const select = event.target as HTMLSelectElement;
		const selectedLanguage = select.value as Language;
		console.log(`Selected language: ${selectedLanguage}`);
		setLanguage(selectedLanguage);
	}
</script>

<Section
	id="appearance"
	title="Appearance"
	description="These settings are stored in your browser."
>
	<div class="flex flex-col space-y-4">
		<fieldset class="fieldset">
			<legend class="fieldset-legend">Language</legend>
			<select onchange={handleLanguageChange} value={getCurrentLanguage()} class="select">
				{#each getAvailableLanguages() as lang}
					<option value={lang.id}>{lang.name}</option>
				{/each}
			</select>
		</fieldset>
	</div>
</Section>
