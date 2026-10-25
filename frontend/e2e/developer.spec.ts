import { expect, test } from '@playwright/test';
import { display, openMode } from './app';

test('converts bases, flips bits and applies bitwise operators', async ({ page }) => {
	await openMode(page, 'Developer');
	const screen = display(page);
	const dec = page.locator('.readout', { hasText: 'DEC' });
	const bin = page.locator('.readout', { hasText: 'BIN' });

	await expect(dec).toContainText('0');

	await page.getByRole('button', { name: 'F', exact: true }).click();
	await expect(dec).toContainText('15');
	await expect(bin).toContainText('1111');

	await page.getByRole('button', { name: 'Bit 0, set' }).click();
	await expect(dec).toContainText('14');

	await page.getByRole('button', { name: '64', exact: true }).click();
	await expect(screen.locator('.expr')).toContainText('64-bit · unsigned');
	await page.getByRole('button', { name: 'UNSIGNED', exact: true }).click();
	await expect(screen.locator('.expr')).toContainText('64-bit · signed');

	await page.getByRole('button', { name: '32', exact: true }).click();
	await page.getByRole('button', { name: 'SIGNED', exact: true }).click();
	await expect(screen.locator('.expr')).toContainText('32-bit · unsigned');

	await page.getByRole('button', { name: 'Clear', exact: true }).click();
	await page.getByRole('button', { name: '5', exact: true }).click();
	await page.getByRole('button', { name: 'AND', exact: true }).click();
	await page.getByRole('button', { name: '3', exact: true }).click();
	await page.getByRole('button', { name: 'Equals' }).click();

	await expect(dec).toContainText('1');
});
