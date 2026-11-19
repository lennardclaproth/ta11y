import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
		runes: ({ filename }) => (filename.split(/[/\\]/).includes('node_modules') ? undefined : true)
	},
	kit: {
		// A single-page app served as static files (Firebase Hosting). Every route renders in the
		// browser (`ssr = false` in the root layout), and unknown paths fall back to index.html.
		adapter: adapter({ fallback: 'index.html' })
	}
};

export default config;
