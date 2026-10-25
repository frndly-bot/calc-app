import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { defineConfig, devices } from '@playwright/test';

const repoRoot = fileURLToPath(new URL('..', import.meta.url));
const port = Number(process.env.CALC_E2E_PORT ?? 34116);
const baseURL = `http://localhost:${port}`;

// Each run gets an empty config dir, so history and settings never leak
// between test runs or into the developer's real ~/.config.
const configHome = mkdtempSync(join(tmpdir(), 'calc-app-e2e-'));

export default defineConfig({
	testDir: './e2e',
	fullyParallel: false,
	workers: 1,
	retries: process.env.CI ? 1 : 0,
	reporter: process.env.CI ? [['github'], ['list']] : 'list',
	timeout: 30_000,
	expect: { timeout: 8_000 },
	use: {
		baseURL,
		trace: 'on-first-retry'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
	webServer: {
		command: 'go run -tags server .',
		cwd: repoRoot,
		url: baseURL,
		reuseExistingServer: false,
		timeout: 180_000,
		stdout: 'ignore',
		stderr: 'pipe',
		env: {
			...process.env,
			CGO_ENABLED: '0',
			CALC_PORT: String(port),
			XDG_CONFIG_HOME: configHome
		}
	}
});
