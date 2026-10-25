import { expect, type Page } from '@playwright/test';

export async function openApp(page: Page) {
	await page.goto('/');
	// Settings load on mount; the theme attribute appears once the backend
	// has answered and the persisted mode has been restored.
	await expect(page.locator('html')).toHaveAttribute('data-theme', /^(dark|light)$/);
}

export async function openMode(
	page: Page,
	label: 'Normal' | 'Scientific' | 'Developer' | 'Finance'
) {
	await openApp(page);
	const tab = page.getByRole('button', { name: label });
	await tab.click();
	await expect(tab).toHaveAttribute('aria-pressed', 'true');
}

export function display(page: Page) {
	return page.locator('.display');
}

export function historyDialog(page: Page) {
	return page.getByRole('dialog', { name: 'Calculation history' });
}

export async function resetHistory(page: Page) {
	await page.getByRole('button', { name: 'Calculation history' }).click();
	const dialog = historyDialog(page);
	await expect(dialog).toBeVisible();
	const clear = dialog.getByRole('button', { name: 'Clear all' });
	if (await clear.isEnabled()) await clear.click();
	await dialog.getByRole('button', { name: 'Close history' }).click();
	await expect(dialog).toBeHidden();
}
