import { expect, test } from '@playwright/test';
import { openMode } from './app';

test('calculates a loan, edits a field with the pad and computes ROI', async ({ page }) => {
	await openMode(page, 'Finance');

	await page.getByRole('button', { name: 'Calculate loan' }).click();
	const result = page.locator('.result');
	await expect(result.locator('.hero')).toContainText('$1,580.17');
	await expect(result).toContainText('Total interest');

	const principal = page.locator('[data-field="principal"] input');
	await principal.fill('300000');
	await page.getByRole('button', { name: 'Calculate loan' }).click();
	await expect(result.locator('.hero')).toContainText('$1,896.20');

	const pad = page.getByRole('group', { name: 'Finance input keypad' });
	await pad.getByRole('button', { name: '5', exact: true }).click();
	await pad.getByRole('button', { name: 'Apply value' }).click();
	await expect(principal).toHaveValue('5');
	await page.getByRole('button', { name: 'Calculate loan' }).click();
	await expect(result.locator('.hero')).toContainText('$0.03');

	await page.getByRole('tab', { name: 'ROI' }).click();
	await page.locator('[data-field="initial"] input').fill('10000');
	await page.locator('[data-field="final"] input').fill('16000');
	await page.getByRole('button', { name: 'Calculate ROI' }).click();
	await expect(page.getByText('60%', { exact: true })).toBeVisible();
});
