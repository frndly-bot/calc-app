import { expect, test } from '@playwright/test';
import { display, openMode } from './app';

test('evaluates arithmetic with the keypad and the keyboard', async ({ page }) => {
	await openMode(page, 'Normal');
	const screen = display(page);

	await page.getByRole('button', { name: '1', exact: true }).click();
	await page.getByRole('button', { name: '2', exact: true }).click();
	await page.getByRole('button', { name: 'Add', exact: true }).click();
	await page.getByRole('button', { name: '3', exact: true }).click();
	await page.getByRole('button', { name: 'Equals' }).click();

	await expect(screen.locator('.expr')).toContainText('12+3 =');
	await expect(screen.locator('.result')).toHaveText('15');

	await page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
	await page.keyboard.press('7');
	await page.keyboard.press('*');
	await page.keyboard.press('6');
	await page.keyboard.press('Enter');

	await expect(screen.locator('.expr')).toContainText('7×6 =');
	await expect(screen.locator('.result')).toHaveText('42');
});

test('theme toggle applies immediately and survives a reload', async ({ page }) => {
	await openMode(page, 'Normal');
	const html = page.locator('html');

	await expect(html).toHaveAttribute('data-theme', 'dark');
	await page.getByRole('button', { name: 'Switch to light theme' }).click();
	await expect(html).toHaveAttribute('data-theme', 'light');

	await page.reload();
	await expect(html).toHaveAttribute('data-theme', 'light');

	await page.getByRole('button', { name: 'Switch to dark theme' }).click();
	await expect(html).toHaveAttribute('data-theme', 'dark');
});
